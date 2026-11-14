package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
)

type disclosure struct {
	arrangementID string
	scope         string
}

// disclosureRecords keeps each record, or gives err and keeps none.
type disclosureRecords struct {
	records []disclosure
	err     error
}

func (d *disclosureRecords) RecordDisclosure(_ context.Context, arrangementID, scope string) error {
	if d.err != nil {
		return d.err
	}

	d.records = append(d.records, disclosure{arrangementID: arrangementID, scope: scope})

	return nil
}

func disclosureSecure() (httpapi.Secure, *disclosureRecords) {
	records := &disclosureRecords{}
	s := consentSecure()
	s.Balances = &arrangementBalances{}
	s.Disclosures = records

	return s, records
}

func TestEachSuccessfulResponseOfAConsentRecordsADisclosureOfTheScopeOfTheAPI(t *testing.T) {
	const basic = "bank:accounts.basic:read"

	tests := map[string]struct {
		request func(t *testing.T) *http.Request
		scope   string
	}{
		"Get Customer": {
			request: func(t *testing.T) *http.Request {
				return resourceRequest(t, http.MethodGet, customerPath, "full")
			},
			scope: "common:customer.basic:read",
		},
		"Get Accounts": {
			request: func(t *testing.T) *http.Request { return accountsRequest(t, accountsPath, "bank-only") },
			scope:   basic,
		},
		"Get Account Detail": {
			request: func(t *testing.T) *http.Request { return detailRequest(t, accountID, "account-detail") },
			scope:   "bank:accounts.detail:read",
		},
		"Get Account Balance": {
			request: func(t *testing.T) *http.Request { return balanceRequest(t, balancePath, "bank-only") },
			scope:   basic,
		},
		"Get Bulk Balances": {
			request: func(t *testing.T) *http.Request {
				req := balanceRequest(t, balancesPath, "bank-only")
				req.Header.Set("x-v", "2")

				return req
			},
			scope: basic,
		},
		"Get Balances For Specific Accounts": {
			request: func(t *testing.T) *http.Request {
				return specificBalancesRequest(t, balancesPath, `{"data": {"accountIds": ["`+accountID+`"]}}`)
			},
			scope: basic,
		},
		"Get Transactions For Account": {
			request: func(t *testing.T) *http.Request {
				return transactionsRequest(t, transactionsPath, "transactions")
			},
			scope: "bank:transactions:read",
		},
		"Get Transaction Detail": {
			request: func(t *testing.T) *http.Request {
				return transactionDetailRequest(t, transactionsPath+"/"+transactionID)
			},
			scope: "bank:transactions:read",
		},
		"UserInfo with the scope profile": {
			request: func(t *testing.T) *http.Request {
				return resourceRequest(t, http.MethodGet, userInfoPath, "full")
			},
			scope: "profile",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s, records := disclosureSecure()

			rec := httptest.NewRecorder()
			secureMux(s).ServeHTTP(rec, tt.request(t))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, []disclosure{{arrangementID: "arrangement-1", scope: tt.scope}}, records.records)
		})
	}
}

func TestAResponseThatDisclosesNoDataClusterRecordsNoDisclosure(t *testing.T) {
	tests := map[string]struct {
		request func(t *testing.T) *http.Request
		status  int
	}{
		"an account that the arrangement does not have": {
			request: func(t *testing.T) *http.Request { return detailRequest(t, "unknown", "account-detail") },
			status:  http.StatusNotFound,
		},
		"a token with no scope of the API": {
			request: func(t *testing.T) *http.Request { return accountsRequest(t, accountsPath, "transactions") },
			status:  http.StatusForbidden,
		},
		"a token that is not valid": {
			request: func(t *testing.T) *http.Request { return accountsRequest(t, accountsPath, "unknown") },
			status:  http.StatusUnauthorized,
		},
		"a version that the API does not have": {
			request: func(t *testing.T) *http.Request {
				req := accountsRequest(t, accountsPath, "bank-only")
				req.Header.Set("x-v", "99")

				return req
			},
			status: http.StatusNotAcceptable,
		},
		"a HEAD request, whose response has no body": {
			request: func(t *testing.T) *http.Request {
				return resourceRequest(t, http.MethodHead, customerPath, "full")
			},
			status: http.StatusOK,
		},
		"UserInfo with no scope profile": {
			request: func(t *testing.T) *http.Request {
				return resourceRequest(t, http.MethodGet, userInfoPath, "no-profile")
			},
			status: http.StatusOK,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s, records := disclosureSecure()

			rec := httptest.NewRecorder()
			secureMux(s).ServeHTTP(rec, tt.request(t))

			require.Equal(t, tt.status, rec.Code, rec.Body.String())
			assert.Empty(t, records.records)
		})
	}
}

func TestAResponseWhoseDisclosureTheDataHolderCannotRecordSendsNoData(t *testing.T) {
	tests := map[string]struct {
		target string
		// version is the x-v of the response of the API, and empty for UserInfo, which has no version.
		version string
	}{
		"Get Customer": {target: customerPath, version: "1"},
		"UserInfo":     {target: userInfoPath},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s, records := disclosureSecure()
			records.err = errors.New("database failure")

			rec := httptest.NewRecorder()
			secureMux(s).ServeHTTP(rec, resourceRequest(t, http.MethodGet, tt.target, "full"))

			requireCDSError(
				t, rec, http.StatusInternalServerError, cds.Unexpected,
				"The data holder cannot record the disclosure.",
			)
			assert.Equal(t, tt.version, rec.Header().Get("x-v"), "the headers of the CDS stay")
			assert.NotContains(t, rec.Body.String(), "Example")
			assert.NotContains(t, rec.Body.String(), "subject-1")
		})
	}
}
