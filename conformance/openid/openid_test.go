//go:build openid

// Package openid runs the FAPI 1.0 Advanced test plan of the OpenID conformance suite, with the Australian CDR
// profile, against the authorisation server of the data holder.
package openid

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

const (
	planName   = "fapi1-advanced-final-test-plan"
	suiteAlias = "cdr-lab"
	// The FIXED_OTPS of wattle-bank in compose.yaml gives this OTP to the customer that the browser of the suite
	// authenticates as.
	customerEmail = "openid@mail.test"
	fixedOTP      = "111111"
	// scope has a scope for the resource endpoint (Get Customer) and a scope that needs an account selection.
	scope = "openid common:customer.basic:read bank:accounts.basic:read"
	// cdrVersion is the version of Get Customer.
	cdrVersion = "1"
)

var planVariant = map[string]string{
	"fapi_profile":             "consumerdataright_au",
	"client_auth_type":         "private_key_jwt",
	"fapi_auth_request_method": "pushed",
	"fapi_response_mode":       "jarm",
}

func TestTheDataHolderPassesTheFAPI1AdvancedPlanWithTheCDRProfile(t *testing.T) {
	ready(t)

	clients := newClients(t)
	serveJWKS(t, clients)

	holder := clients[0].register(t)
	clients[1].register(t)

	customerNumber := newCustomer(t)

	s := suite{t: t, baseURL: labtest.Env(t, "OPENID_SUITE_URL"), http: labtest.Client(t)}
	p := s.createPlan(planName, planVariant, configuration(t, clients, holder, customerNumber))

	publicURL := labtest.Env(t, "OPENID_SUITE_PUBLIC_URL")
	t.Logf("plan: %s/plan-detail.html?plan=%s", publicURL, p.ID)

	var failed, review []string

	for _, module := range p.Modules {
		id, info := s.run(p.ID, module)
		t.Logf("%-90s %-11s %-8s %s/log-detail.html?log=%s",
			module.TestModule, info.Status, info.Result, publicURL, id)

		switch {
		case info.Status != statusFinished:
			failed = append(failed, module.TestModule)
		case info.Result == resultFailed, info.Result == resultWarning:
			if !onlyAcceptedFindings(t, s.findings(id)) {
				failed = append(failed, module.TestModule)
			}
		case info.Result == resultReview:
			review = append(review, module.TestModule)
		}
	}

	if len(review) > 0 {
		t.Logf("these modules need a person to examine the screenshot in their log: %s", strings.Join(review, ", "))
	}

	if len(failed) > 0 {
		t.Errorf("these modules failed, have a warning or did not finish: %s", strings.Join(failed, ", "))
	}
}

func configuration(t *testing.T, clients []*client, holder dataHolder, customerNumber string) map[string]any {
	t.Helper()

	pages := consumerPages{authorise: holder.issuer + "/authorize*", customerNumber: customerNumber}

	return map[string]any{
		"alias":       suiteAlias,
		"description": "cdr-lab data holder",
		"server": map[string]any{
			"discoveryUrl": holder.issuer + "/.well-known/openid-configuration",
			// The CDS requires this field (CDS: OpenID Provider Configuration endpoint), and RFC 8414 does not have it.
			"allow_unexpected_metadata_fields": []string{"cdr_arrangement_revocation_endpoint"},
		},
		"client":  suiteClient(clients[0]),
		"mtls":    clients[0].mtls(t),
		"client2": suiteClient(clients[1]),
		"mtls2":   clients[1].mtls(t),
		"resource": map[string]any{
			"resourceUrl": holder.resourceBaseURI + "/cds-au/v1/common/customer",
			"cdrVersion":  cdrVersion,
		},
		"browser": []any{pages.consent()},
		"override": map[string]any{
			"fapi1-advanced-final-user-rejects-authentication": map[string]any{
				"browser": []any{pages.cancel()},
			},
			"fapi1-advanced-final-par-ensure-reused-request-uri-prior-to-auth-completion-succeeds": map[string]any{
				"browser": []any{pages.visitOnly(), pages.consent()},
			},
		},
	}
}

func suiteClient(c *client) map[string]any {
	return map[string]any{
		"client_id": c.clientID,
		"scope":     scope,
		"jwks":      jose.JSONWebKeySet{Keys: []jose.JSONWebKey{c.privateJWK()}},
	}
}

// newCustomer creates the customer with the fixed OTP and one open account, for the account selection of the consent.
func newCustomer(t *testing.T) string {
	t.Helper()

	var customer struct {
		CustomerNumber string `json:"customerNumber"`
	}
	operate(t, "/v1/customers", map[string]any{
		"lastName": "Conformance", "emailAddress": customerEmail,
	}, &customer)

	accountNumber, err := rand.Int(rand.Reader, big.NewInt(90_000_000))
	require.NoError(t, err)

	operate(t, "/v1/customers/"+customer.CustomerNumber+"/accounts", map[string]any{
		"displayName":     "Everyday",
		"productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
		"productName":     "Everyday Account",
		"bsb":             "372200",
		"accountNumber":   fmt.Sprintf("%d", 10_000_000+accountNumber.Int64()),
	}, nil)

	return customer.CustomerNumber
}

func operate(t *testing.T, path string, body, into any) {
	t.Helper()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodPost, labtest.Env(t, "DATAHOLDER_OPERATOR_URL")+path, bytes.NewReader(data),
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+labtest.Env(t, "DATAHOLDER_OPERATOR_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	reply, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(reply))

	if into != nil {
		require.NoError(t, json.Unmarshal(reply, into))
	}
}
