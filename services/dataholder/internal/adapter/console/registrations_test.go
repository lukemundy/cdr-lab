package console_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// The method is a constant, because gosec takes a literal value of a field with "token" in its name for a credential.
const privateKeyJWT = "private_key_jwt"

func registration(clientID, softwareID string) domain.Registration {
	return domain.Registration{
		ClientID:                    clientID,
		ClientIDIssuedAt:            1792454400,
		ClientName:                  "Recipient Web",
		ClientDescription:           "A data recipient for tests",
		ClientURI:                   "https://recipient.example",
		LegalEntityID:               "legal-entity-1",
		LegalEntityName:             "Recipient Pty Ltd",
		OrgID:                       "brand-1",
		OrgName:                     "Recipient",
		RedirectURIs:                []string{"https://recipient.example/cb", "https://recipient.example/cb2"},
		LogoURI:                     "https://recipient.example/logo.png",
		JWKSURI:                     "https://recipient.example/jwks",
		RevocationURI:               "https://recipient.example/revoke",
		RecipientBaseURI:            "https://recipient.example/cdr",
		TokenEndpointAuthMethod:     privateKeyJWT,
		TokenEndpointAuthSigningAlg: domain.PS256,
		GrantTypes: []domain.GrantType{
			domain.GrantClientCredentials,
			domain.GrantAuthorizationCode,
		},
		ResponseTypes:                  []string{"code"},
		ApplicationType:                "web",
		IDTokenSignedResponseAlg:       domain.PS256,
		AuthorizationSignedResponseAlg: domain.PS256,
		RequestObjectSigningAlg:        domain.PS256,
		SoftwareStatement:              "header.ssa-zebracorn.signature",
		SoftwareID:                     softwareID,
		SoftwareRoles:                  "data-recipient-software-product",
		Scope:                          "openid bank:accounts.basic:read cdr:registration",
	}
}

func (r records) register(t *testing.T, registration domain.Registration) {
	t.Helper()

	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), r.events, registration.SoftwareID, domain.RegisterClient{Registration: registration},
	))
}

type registrationTimes struct {
	RegisteredAt time.Time `json:"registeredAt"`
	LastUpdated  time.Time `json:"lastUpdated"`
}

// Verifies: DH-1642
func TestTheRegistrationsGiveEachRegistrationWithItsMetadataTimesAndRegisterStatusAndNoSSA(t *testing.T) {
	r := newRecords(t)
	r.register(t, registration("client-1", "software-1"))

	second := registration("client-2", "software-2")
	second.ClientName = "Another Web"
	second.LegalEntityID, second.LegalEntityName = "", ""
	r.register(t, second)

	require.NoError(t, statusstore.New(r.pool, nil).Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive,
	}, `"v1"`))

	resp := get(t, r.ports, "/api/registrations")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.NotContains(t, resp.body, "zebracorn", "the console shows no SSA")

	var list struct {
		Data []json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &list))
	require.Len(t, list.Data, 2)

	var times [2]registrationTimes
	for i := range times {
		require.NoError(t, json.Unmarshal(list.Data[i], &times[i]))
	}

	assert.Contains(t, string(list.Data[0]), `"client_id":"client-2"`, "the list is in the order of the client_name")
	assert.False(t, times[1].RegisteredAt.IsZero())
	assert.Equal(t, times[1].RegisteredAt, times[1].LastUpdated)

	assert.JSONEq(t, `{
		"registeredAt": "`+times[1].RegisteredAt.Format(time.RFC3339Nano)+`",
		"lastUpdated": "`+times[1].LastUpdated.Format(time.RFC3339Nano)+`",
		"softwareProductStatus": "ACTIVE", "arrangementCount": 0, "currentArrangementCount": 0,
		"metadata": {
			"client_id": "client-1", "client_id_issued_at": 1792454400, "client_name": "Recipient Web",
			"client_description": "A data recipient for tests", "client_uri": "https://recipient.example",
			"legal_entity_id": "legal-entity-1", "legal_entity_name": "Recipient Pty Ltd",
			"org_id": "brand-1", "org_name": "Recipient",
			"redirect_uris": ["https://recipient.example/cb", "https://recipient.example/cb2"],
			"logo_uri": "https://recipient.example/logo.png", "jwks_uri": "https://recipient.example/jwks",
			"revocation_uri": "https://recipient.example/revoke",
			"recipient_base_uri": "https://recipient.example/cdr",
			"token_endpoint_auth_method": "private_key_jwt", "token_endpoint_auth_signing_alg": "PS256",
			"grant_types": ["client_credentials", "authorization_code"], "response_types": ["code"],
			"application_type": "web", "id_token_signed_response_alg": "PS256",
			"authorization_signed_response_alg": "PS256", "request_object_signing_alg": "PS256",
			"software_id": "software-1", "software_roles": "data-recipient-software-product",
			"scope": "openid bank:accounts.basic:read cdr:registration"
		}
	}`, string(list.Data[1]))
	assert.NotContains(t, string(list.Data[0]), "softwareProductStatus", "the data holder has no status for it")
	assert.NotContains(t, string(list.Data[0]), "legal_entity", "an optional value that is empty is not given")
}

func TestTheDetailOfARegistrationGivesTheValuesAndTheTimeOfTheLastUpdate(t *testing.T) {
	r := newRecords(t)
	r.register(t, registration("client-1", "software-1"))

	updated := registration("", "software-1")
	updated.RedirectURIs = []string{"https://recipient.example/new"}
	updated.RecipientBaseURI = "https://recipient.example/new-cdr"
	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), r.events, "software-1", domain.UpdateClient{ClientID: "client-1", Registration: updated},
	))

	resp := get(t, r.ports, "/api/registrations/client-1")

	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var detail struct {
		registrationTimes
		Metadata struct {
			RedirectURIs     []string `json:"redirect_uris"`
			RecipientBaseURI string   `json:"recipient_base_uri"`
		} `json:"metadata"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &detail))
	assert.Equal(t, []string{"https://recipient.example/new"}, detail.Metadata.RedirectURIs)
	assert.Equal(t, "https://recipient.example/new-cdr", detail.Metadata.RecipientBaseURI)
	assert.True(t, detail.LastUpdated.After(detail.RegisteredAt), "the update changes the time of the last change")
}

// Verifies: DH-1622
func TestAnUnknownRegistrationIsNotFoundWithNoIdentifierInTheDetail(t *testing.T) {
	resp := get(t, newRecords(t).ports, "/api/registrations/client-zebracorn")

	require.Equal(t, http.StatusNotFound, resp.status)
	assert.NotContains(t, resp.body, "zebracorn")
}

func TestNoRegistrationsGiveAnEmptyList(t *testing.T) {
	resp := get(t, newRecords(t).ports, "/api/registrations")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"data": []}`, resp.body)
}

func (r records) arrangementOfClient(t *testing.T, id, clientID, customerNumber string, sharing time.Duration) {
	t.Helper()

	require.NoError(t, app.ArrangementType.Execute(t.Context(), r.events, id, domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: id, CustomerNumber: customerNumber, ClientID: clientID, SoftwareID: "software-1",
			ClientName: "Recipient Web", Scopes: []string{"openid"}, AccountIDs: []string{},
			ConfirmedAt: now.Add(-time.Hour), SharingDurationSeconds: int64(sharing.Seconds()),
		},
	}))
}

type registrationCounts struct {
	Data []struct {
		ArrangementCount        int `json:"arrangementCount"`
		CurrentArrangementCount int `json:"currentArrangementCount"`
		Metadata                struct {
			ClientID string `json:"client_id"`
		} `json:"metadata"`
	} `json:"data"`
}

// Verifies: DH-1631
func TestTheRegistrationsGiveTheNumbersOfTheArrangementsOfTheClientAndSortByEachKey(t *testing.T) {
	r := newRecords(t)
	r.register(t, registration("client-1", "software-1"))

	other := registration("client-2", "software-2")
	other.ClientName = "another web"
	r.register(t, other)

	r.arrangementOfClient(t, "arrangement-1", "client-1", "00000001", 24*time.Hour)
	// The sharing ends at the time of the request.
	r.arrangementOfClient(t, "arrangement-2", "client-1", "00000001", time.Hour)
	r.arrangementOfClient(t, "arrangement-3", "client-1", "00000002", time.Hour)

	list := func(target string) registrationCounts {
		t.Helper()

		resp := get(t, r.ports, target)
		require.Equal(t, http.StatusOK, resp.status, resp.body)

		var body registrationCounts
		require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

		return body
	}
	ids := func(target string) []string {
		t.Helper()

		var got []string
		for _, d := range list(target).Data {
			got = append(got, d.Metadata.ClientID)
		}

		return got
	}

	byName := list("/api/registrations")
	require.Len(t, byName.Data, 2)
	assert.Equal(t, "client-2", byName.Data[0].Metadata.ClientID, "the name sort has no difference for case")
	assert.Equal(t, 3, byName.Data[1].ArrangementCount)
	assert.Equal(t, 1, byName.Data[1].CurrentArrangementCount)

	assert.Equal(t, []string{"client-1", "client-2"}, ids("/api/registrations?sort=clientName&order=desc"))
	assert.Equal(t, []string{"client-1", "client-2"}, ids("/api/registrations?sort=registeredAt"))
	assert.Equal(t, []string{"client-2", "client-1"}, ids("/api/registrations?sort=registeredAt&order=desc"))
	assert.Equal(t, []string{"client-2", "client-1"}, ids("/api/registrations?sort=currentArrangementCount"))
	assert.Equal(t, []string{"client-1", "client-2"}, ids("/api/registrations?sort=arrangementCount&order=desc"))
	assert.Equal(t, []string{"client-2", "client-1"}, ids("/api/registrations?sort=&order=&status=ACTIVE"),
		"an empty or unknown parameter is ignored")
}

func TestTheRegistrationsRefuseASortOrAnOrderThatIsNotKnown(t *testing.T) {
	for _, target := range []string{"/api/registrations?sort=zebracorn", "/api/registrations?order=zebracorn"} {
		resp := get(t, newRecords(t).ports, target)

		assert.Equal(t, http.StatusBadRequest, resp.status, target)
	}
}

// Verifies: DH-1638
func TestTheDetailOfARegistrationGivesTheCountsAndNoArrangements(t *testing.T) {
	r := newRecords(t)
	r.register(t, registration("client-1", "software-1"))
	r.arrangementOfClient(t, "arrangement-1", "client-1", "00000001", 24*time.Hour)
	r.arrangementOfClient(t, "arrangement-2", "client-1", "00000002", time.Hour)

	resp := get(t, r.ports, "/api/registrations/client-1")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.NotContains(t, resp.body, "arrangement-1")

	var detail struct {
		ArrangementCount        int `json:"arrangementCount"`
		CurrentArrangementCount int `json:"currentArrangementCount"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &detail))
	assert.Equal(t, 2, detail.ArrangementCount)
	assert.Equal(t, 1, detail.CurrentArrangementCount)
}

// Verifies: DH-1636, DH-1637
func TestTheArrangementsOfARegistrationAreNewestFirstInPagesWithTheCustomerAndTheStatus(t *testing.T) {
	r := newRecords(t)
	r.register(t, registration("client-1", "software-1"))
	addCustomer(t, r.customers, "00000001", "Alex", "Example")
	addCustomer(t, r.customers, "00000002", "", "Sample")
	r.arrangementOfClient(t, "arrangement-1", "client-1", "00000001", 24*time.Hour)
	// The sharing ends at the time of the request.
	r.arrangementOfClient(t, "arrangement-2", "client-1", "00000002", time.Hour)
	r.arrangementOfClient(t, "arrangement-3", "client-1", "00000001", 24*time.Hour)
	r.arrangementOfClient(t, "arrangement-4", "client-2", "00000001", 24*time.Hour)

	resp := get(t, r.ports, "/api/registrations/client-1/arrangements?pageSize=2")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [
		{
			"arrangementId": "arrangement-1", "customerNumber": "00000001", "firstName": "Alex",
			"lastName": "Example", "status": "CURRENT",
			"confirmedAt": "2026-10-19T23:00:00Z", "sharingEndsAt": "2026-10-20T23:00:00Z"
		},
		{
			"arrangementId": "arrangement-2", "customerNumber": "00000002", "lastName": "Sample",
			"status": "ENDED", "confirmedAt": "2026-10-19T23:00:00Z", "sharingEndsAt": "2026-10-20T00:00:00Z"
		}
	], "totalRecords": 3}`, resp.body)

	resp = get(t, r.ports, "/api/registrations/client-1/arrangements?page=3&pageSize=2")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [], "totalRecords": 3}`, resp.body)
}

func TestARegistrationWithNoArrangementsGivesAnEmptyPage(t *testing.T) {
	r := newRecords(t)
	r.register(t, registration("client-1", "software-1"))

	resp := get(t, r.ports, "/api/registrations/client-1/arrangements")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [], "totalRecords": 0}`, resp.body)
}

// Verifies: DH-1622
func TestTheArrangementsOfAnUnknownRegistrationAreNotFoundWithNoIdentifierInTheDetail(t *testing.T) {
	resp := get(t, newRecords(t).ports, "/api/registrations/client-zebracorn/arrangements")

	require.Equal(t, http.StatusNotFound, resp.status)
	assert.NotContains(t, resp.body, "zebracorn")
}

func TestAPageOfTheArrangementsOfARegistrationThatIsNotValidGivesBadRequestWithNoValue(t *testing.T) {
	r := newRecords(t)
	r.register(t, registration("client-1", "software-1"))

	for _, target := range []string{"page=zebracorn", "page=0", "pageSize=101"} {
		resp := get(t, r.ports, "/api/registrations/client-1/arrangements?"+target)

		require.Equal(t, http.StatusBadRequest, resp.status, target)
		assert.NotContains(t, resp.body, "zebracorn", target)
	}
}
