//go:build conformance

package register

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// This data recipient of services/register/seed/example.json is only
// for the caller status test, which changes its status.
const (
	statusTestRecipient = "41090368-a571-41df-a2c2-a9ad32dfa512"
	statusTestProduct   = "8bbe8ce3-3d90-44b2-b97d-383d7fa7e579"
)

const (
	dataHolderBrandsPath = "/cdr-register/v1/all/data-holders/brands"
	// The seed file has 12 data holder brands with a dataHolderBrandId, and 2 with only an interimId.
	seedRegisteredBrands = 12
)

// The CDS OpenAPI specification gives only these statuses for Get Data Holder Brands.
var specifiedBrandStatuses = []int{
	http.StatusOK,
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusNotAcceptable,
}

type secureReply struct {
	status int
	header http.Header
	body   []byte
}

func secureCall(t *testing.T, cert tls.Certificate, target string, header http.Header) secureReply {
	t.Helper()
	return secureRequest(t, cert, target, header, specifiedBrandStatuses)
}

// secureRequest validates the response against the OpenAPI specification when its status is in validated.
func secureRequest(t *testing.T, cert tls.Certificate, target string, header http.Header, validated []int) secureReply {
	t.Helper()

	httpClient := &http.Client{Timeout: labtest.RequestTimeout, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      labtest.CDRCABundle(t),
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}}
	if !strings.HasPrefix(target, "https://") {
		target = labtest.Env(t, "REGISTER_SECURE_URL") + target
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	require.NoError(t, err)

	req.Header = header
	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	if slices.Contains(validated, resp.StatusCode) {
		require.NoError(t, spec(t).ValidateResponse(resp))
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return secureReply{status: resp.StatusCode, header: resp.Header, body: body}
}

func bearerV3(token string) http.Header {
	return http.Header{"Authorization": {"Bearer " + token}, "X-V": {"3"}}
}

type brandPage struct {
	Data []struct {
		DataHolderBrandID string   `json:"dataHolderBrandId"`
		Industries        []string `json:"industries"`
		Status            string   `json:"status"`
		LegalEntity       struct {
			LegalEntityID string `json:"legalEntityId"`
			Status        string `json:"status"`
		} `json:"legalEntity"`
		EndpointDetail struct {
			ResourceBaseURI string `json:"resourceBaseUri"`
			InfosecBaseURI  string `json:"infosecBaseUri"`
		} `json:"endpointDetail"`
		AuthDetails []struct {
			RegisterUType string `json:"registerUType"`
			JWKSEndpoint  string `json:"jwksEndpoint"`
		} `json:"authDetails"`
		LastUpdated time.Time `json:"lastUpdated"`
	} `json:"data"`
	Links map[string]string `json:"links"`
	Meta  struct {
		TotalRecords int `json:"totalRecords"`
		TotalPages   int `json:"totalPages"`
	} `json:"meta"`
}

func brands(t *testing.T, cert tls.Certificate, token, target string) brandPage {
	t.Helper()
	r := secureCall(t, cert, target, bearerV3(token))
	require.Equal(t, http.StatusOK, r.status, string(r.body))
	assert.Equal(t, "3", r.header.Get("x-v"))
	assert.NotEmpty(t, r.header.Get("x-fapi-interaction-id"))

	var page brandPage
	require.NoError(t, json.Unmarshal(r.body, &page))

	return page
}

func requireSecureCDSError(t *testing.T, r secureReply, status int, code, title string) {
	t.Helper()
	requireCDSError(t, response{Response: &http.Response{StatusCode: r.status}, body: r.body}, status, code, title)
}

// Verifies: REG-0300, REG-0333
func TestADataRecipientGetsTheDataHolderBrandsOfEachIndustryWithTheirEndpoints(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)

	tests := map[string]int{"all": seedRegisteredBrands, "banking": 5, "energy": 3, "telco": 3, "non-bank-lending": 2}
	for industry, want := range tests {
		t.Run(industry, func(t *testing.T) {
			page := brands(t, cert, token, "/cdr-register/v1/"+industry+"/data-holders/brands")

			assert.Len(t, page.Data, want)
			assert.Equal(t, want, page.Meta.TotalRecords)

			statuses := map[string]bool{}
			for _, b := range page.Data {
				statuses[b.Status] = true
				assert.NotEmpty(t, b.DataHolderBrandID, "a brand with only an interimId is not in the list")
				assert.NotEmpty(t, b.EndpointDetail.InfosecBaseURI)
				assert.NotEmpty(t, b.EndpointDetail.ResourceBaseURI)
				assert.NotEmpty(t, b.LegalEntity.LegalEntityID)
				require.Len(t, b.AuthDetails, 1)
				assert.Equal(t, "SIGNED-JWT", b.AuthDetails[0].RegisterUType)
				assert.NotEmpty(t, b.AuthDetails[0].JWKSEndpoint)

				if industry != "all" {
					assert.Contains(t, b.Industries, industry)
				}
			}

			if industry == "all" {
				assert.Equal(
					t,
					map[string]bool{"ACTIVE": true, "INACTIVE": true, "REMOVED": true},
					statuses,
					"the list has brands in each status",
				)
			}
		})
	}
}

// Verifies: REG-0316
func TestGetDataHolderBrandsRefusesAnIndustryThatIsNotValid(t *testing.T) {
	cert := activeTokenClient(t)

	r := secureCall(
		t,
		cert,
		"/cdr-register/v1/insurance/data-holders/brands",
		bearerV3(accessToken(t, idpTestActiveProduct, cert)),
	)

	requireSecureCDSError(t, r, http.StatusBadRequest, "urn:au-cds:error:cds-all:Field/Invalid", "Invalid Field")
}

// Verifies: REG-0330, REG-0332
func TestGetDataHolderBrandsPagesFollowTheCDSPaginationRules(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)
	secureBase := labtest.Env(t, "REGISTER_SECURE_URL")

	var ids []string

	target := dataHolderBrandsPath + "?page-size=4"
	for pageNumber := 1; target != ""; pageNumber++ {
		require.LessOrEqual(t, pageNumber, 3, "the next links stop at the last page")
		page := brands(t, cert, token, target)

		assert.Equal(t, seedRegisteredBrands, page.Meta.TotalRecords)
		assert.Equal(t, 3, page.Meta.TotalPages)

		for name, link := range page.Links {
			assert.True(
				t,
				strings.HasPrefix(link, secureBase+dataHolderBrandsPath+"?"),
				"%s is a URL of the secure APIs: %s",
				name,
				link,
			)
			assert.Contains(t, link, "page-size=4")
		}

		assert.Contains(t, page.Links["self"], "page="+strconv.Itoa(pageNumber))
		assert.Equal(t, pageNumber > 1, page.Links["prev"] != "", "prev is present only after the first page")
		assert.Equal(t, pageNumber < 3, page.Links["next"] != "", "next is present only before the last page")
		assert.NotEmpty(t, page.Links["first"])
		assert.NotEmpty(t, page.Links["last"])

		for _, b := range page.Data {
			ids = append(ids, b.DataHolderBrandID)
		}

		target = page.Links["next"]
	}

	assert.Len(t, ids, seedRegisteredBrands)
	assert.True(t, slices.IsSorted(ids), "the brands are in the order of their IDs")
	assert.Len(t, slices.Compact(slices.Clone(ids)), seedRegisteredBrands, "no brand is on two pages")
}

func TestGetDataHolderBrandsDefaultsToTheFirstPageOf25(t *testing.T) {
	cert := activeTokenClient(t)

	page := brands(t, cert, accessToken(t, idpTestActiveProduct, cert), dataHolderBrandsPath)

	assert.Contains(t, page.Links["self"], "page=1")
	assert.Contains(t, page.Links["self"], "page-size=25")
	assert.Equal(t, 1, page.Meta.TotalPages)
}

// Verifies: REG-0331
func TestGetDataHolderBrandsUpdatedSinceGivesOnlyTheBrandsWithALaterOrEqualUpdate(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)
	all := brands(t, cert, token, dataHolderBrandsPath)
	require.NotEmpty(t, all.Data)

	oldest := all.Data[0].LastUpdated
	for _, b := range all.Data {
		if b.LastUpdated.Before(oldest) {
			oldest = b.LastUpdated
		}
	}

	since := func(ts time.Time) string {
		return dataHolderBrandsPath + "?" + url.Values{"updated-since": {ts.UTC().Format(time.RFC3339)}}.Encode()
	}

	fromOldest := brands(t, cert, token, since(oldest.Truncate(time.Second)))
	future := brands(t, cert, token, since(time.Now().Add(time.Hour)))

	assert.Equal(t, seedRegisteredBrands, fromOldest.Meta.TotalRecords)
	assert.Contains(t, fromOldest.Links["self"], "updated-since=")
	assert.Empty(t, future.Data)
	assert.Equal(t, 0, future.Meta.TotalRecords)
	assert.Equal(t, 0, future.Meta.TotalPages, "totalPages is 0 when totalRecords is 0")
}

func TestGetDataHolderBrandsRefusesQueryParametersThatAreNotValid(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)

	tests := map[string]struct {
		query       string
		status      int
		code, title string
	}{
		"page-size above 1000": {
			"page-size=1001",
			400,
			"urn:au-cds:error:cds-all:Field/InvalidPageSize",
			"Invalid Page Size",
		},
		"page-size zero": {
			"page-size=0",
			400,
			"urn:au-cds:error:cds-all:Field/InvalidPageSize",
			"Invalid Page Size",
		},
		"page zero":         {"page=0", 400, "urn:au-cds:error:cds-all:Field/Invalid", "Invalid Field"},
		"page out of range": {"page=2", 422, "urn:au-cds:error:cds-all:Field/InvalidPage", "Invalid Page"},
		"updated-since not a date": {
			"updated-since=yesterday",
			400,
			"urn:au-cds:error:cds-all:Field/InvalidDateTime",
			"Invalid Date",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := secureCall(t, cert, dataHolderBrandsPath+"?"+tt.query, bearerV3(token))

			requireSecureCDSError(t, r, tt.status, tt.code, tt.title)
		})
	}
}

func TestGetDataHolderBrandsAcceptsTheMaximumPageSize(t *testing.T) {
	cert := activeTokenClient(t)

	page := brands(t, cert, accessToken(t, idpTestActiveProduct, cert), dataHolderBrandsPath+"?page-size=1000")

	assert.Len(t, page.Data, seedRegisteredBrands)
}

// Verifies: REG-0313
func TestGetDataHolderBrandsServesOnlyVersion3(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)
	authorization := "Bearer " + token

	missing := secureCall(t, cert, dataHolderBrandsPath, http.Header{"Authorization": {authorization}})
	old := secureCall(t, cert, dataHolderBrandsPath, http.Header{"Authorization": {authorization}, "X-V": {"2"}})

	requireSecureCDSError(
		t,
		missing,
		http.StatusBadRequest,
		"urn:au-cds:error:cds-all:Header/Missing",
		"Missing Required Header",
	)
	requireSecureCDSError(
		t,
		old,
		http.StatusNotAcceptable,
		"urn:au-cds:error:cds-all:Header/UnsupportedVersion",
		"Unsupported Version",
	)
}

// Verifies: REG-0320, REG-0321
func TestGetDataHolderBrandsRefusesARequestWithoutAValidToken(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)
	otherCert := labtest.ClientCertificate(t, "PKI_UNREGISTERED_CLIENT_CERT", "PKI_UNREGISTERED_CLIENT_KEY")

	tests := map[string]struct {
		cert      tls.Certificate
		header    http.Header
		challenge string
	}{
		"no token":                  {cert, http.Header{"X-V": {"3"}}, `Bearer`},
		"a token that is not a JWT": {cert, bearerV3("not-a-token"), `Bearer error="invalid_token"`},
		"a token that is changed": {
			cert,
			bearerV3(token[:len(token)-4] + "AAAA"),
			`Bearer error="invalid_token"`,
		},
		"a token of a different client certificate": {otherCert, bearerV3(token), `Bearer error="invalid_token"`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := secureCall(t, tt.cert, dataHolderBrandsPath, tt.header)

			assert.Equal(t, http.StatusUnauthorized, r.status, string(r.body))
			assert.Equal(t, tt.challenge, r.header.Get("WWW-Authenticate"))
			assert.Empty(t, r.body)
		})
	}
}

// Verifies: REG-0323
func TestGetDataHolderBrandsRefusesACallerThatIsNotActive(t *testing.T) {
	cert := tokenClient(t, statusTestRecipient, statusTestProduct, "PKI_STATUS_CLIENT_CERT", "PKI_STATUS_CLIENT_KEY")
	token := accessToken(t, statusTestProduct, cert)
	status := func() secureReply { return secureCall(t, cert, dataHolderBrandsPath, bearerV3(token)) }
	operator := func(path string) {
		t.Helper()
		code, p := operate(t, path)
		require.Equal(t, http.StatusNoContent, code, p.Detail)
	}
	refused := func(detail string) {
		t.Helper()

		var r secureReply

		require.Eventually(
			t,
			func() bool { r = status(); return r.status == http.StatusForbidden },
			propagation,
			50*time.Millisecond,
			"the token is refused",
		)
		requireSecureCDSError(
			t,
			r,
			http.StatusForbidden,
			"urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive",
			"ADR Status Is Not Active",
		)
		assert.Contains(t, string(r.body), detail)
	}
	accepted := func() {
		t.Helper()
		require.Eventually(
			t,
			func() bool { return status().status == http.StatusOK },
			propagation,
			50*time.Millisecond,
			"the token is accepted",
		)
	}
	accepted()

	operator(productAction(statusTestRecipient, statusTestProduct, "deactivate"))
	refused("INACTIVE")
	operator(productAction(statusTestRecipient, statusTestProduct, "activate"))
	accepted()

	operator(recipientAction(statusTestRecipient, "suspend"))
	refused("INACTIVE")
	operator(recipientAction(statusTestRecipient, "reinstate"))
	// The test leaves the participants ACTIVE, so that it can run again.
	operator(productAction(statusTestRecipient, statusTestProduct, "activate"))
	accepted()
}
