package operatorapi_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	arrangementID     = "8f4d7a52-3c1e-4b6a-9d2f-1e0c5b7a9d31"
	arrangementsPath  = "/v1/customers/" + customerNumber + "/arrangements"
	withdrawPath      = arrangementsPath + "/" + arrangementID + "/withdraw"
	arrangementFields = `"arrangementId":"` + arrangementID + `","softwareId":"software-1",` +
		`"clientName":"Recipient Web","scopes":["openid","bank:accounts.basic:read"],"accountIds":["` + accountID +
		`"],"confirmedAt":"2026-10-17T01:00:00Z","sharingEndsAt":"2026-10-17T02:00:00Z"`
	activeJSON    = `{` + arrangementFields + `}`
	withdrawnJSON = `{` + arrangementFields + `,"endedAt":"2026-10-17T01:30:00Z","endReason":"WITHDRAWN_BY_CONSUMER",` +
		`"withdrawalMethod":"ALTERNATIVE_METHOD","noticeStatus":"PENDING"}`
)

var confirmedAt = time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)

func activeArrangement() domain.ArrangementState {
	return domain.ArrangementState{Created: true, Arrangement: domain.Arrangement{
		ArrangementID:          arrangementID,
		CustomerNumber:         customerNumber,
		ClientID:               "client-1",
		SoftwareID:             "software-1",
		ClientName:             "Recipient Web",
		Scopes:                 []string{"openid", "bank:accounts.basic:read"},
		AccountIDs:             []string{accountID},
		ConfirmedAt:            confirmedAt,
		SharingDurationSeconds: 3600,
	}}
}

func withdrawnArrangement() domain.ArrangementState {
	state := activeArrangement()
	state.EndedAt, state.EndReason = confirmedAt.Add(30*time.Minute), domain.EndReasonWithdrawnByConsumer

	return state
}

type fakeArrangements struct {
	states    []domain.ArrangementState
	err       error
	listed    string
	withdrawn [2]string
	method    domain.WithdrawalMethod
}

func (f *fakeArrangements) OfCustomer(_ context.Context, number string) ([]domain.ArrangementState, error) {
	f.listed = number

	return f.states, f.err
}

func (f *fakeArrangements) WithdrawByConsumer(
	_ context.Context,
	number, id string,
	method domain.WithdrawalMethod,
) (domain.ArrangementState, error) {
	f.withdrawn, f.method = [2]string{number, id}, method

	if f.err != nil {
		return domain.ArrangementState{}, f.err
	}

	return withdrawnArrangement(), nil
}

func serveArrangements(t *testing.T, arrangements *fakeArrangements, req request) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	operatorapi.Register(mux, &fakeCustomers{}, arrangements, access)

	httpReq := httptest.NewRequestWithContext(t.Context(), req.method, req.path, strings.NewReader(req.body))
	httpReq.Header = req.header

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httpReq)

	return rec
}

func readArrangements() request {
	return request{method: http.MethodGet, path: arrangementsPath, header: authorised()}
}

func withdraw() request {
	return request{method: http.MethodPost, path: withdrawPath, header: authorised()}
}

func TestTheArrangementsOfTheCustomerOfThePathAreRead(t *testing.T) {
	arrangements := &fakeArrangements{states: []domain.ArrangementState{activeArrangement(), withdrawnArrangement()}}

	rec := serveArrangements(t, arrangements, readArrangements())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, "["+activeJSON+","+withdrawnJSON+"]", rec.Body.String())
	assert.Equal(t, customerNumber, arrangements.listed)
}

func TestACustomerWithNoArrangementsHasAnEmptyList(t *testing.T) {
	rec := serveArrangements(t, &fakeArrangements{}, readArrangements())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, "[]", rec.Body.String())
}

func TestAWithdrawalGivesTheEndedArrangement(t *testing.T) {
	arrangements := &fakeArrangements{}

	rec := serveArrangements(t, arrangements, withdraw())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, withdrawnJSON, rec.Body.String())
	assert.Equal(t, [2]string{customerNumber, arrangementID}, arrangements.withdrawn)
	assert.Equal(t, domain.WithdrawalMethodAlternative, arrangements.method, "the operator acts for the consumer")
}

func TestArrangementErrors(t *testing.T) {
	noCustomer := errors.Join(errors.New("read the customer"), app.ErrCustomerNotFound)

	tests := map[string]struct {
		request request
		err     error
		status  int
		detail  string
	}{
		"read for no customer": {
			request: readArrangements(),
			err:     noCustomer,
			status:  http.StatusNotFound,
			detail:  "The data holder has no customer with this customer number.",
		},
		"withdrawal for no customer": {
			request: withdraw(),
			err:     noCustomer,
			status:  http.StatusNotFound,
			detail:  "The data holder has no customer with this customer number.",
		},
		"withdrawal of an arrangement that the customer does not have": {
			request: withdraw(),
			err:     domain.ErrArrangementNotFound,
			status:  http.StatusNotFound,
			detail:  "The customer has no arrangement with this arrangementId.",
		},
		"withdrawal of an arrangement that ended": {
			request: withdraw(),
			err:     fmt.Errorf("end the arrangement: %w", domain.ErrArrangementEnded),
			status:  http.StatusConflict,
			detail:  "The arrangement ended.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serveArrangements(t, &fakeArrangements{err: tt.err}, tt.request)

			assert.Equal(t, tt.detail, requireProblem(t, rec, tt.status))
		})
	}
}

func TestEachArrangementRouteNeedsTheOperatorToken(t *testing.T) {
	for name, req := range map[string]request{"read": readArrangements(), "withdraw": withdraw()} {
		t.Run(name, func(t *testing.T) {
			arrangements := &fakeArrangements{}
			req.header = http.Header{}

			requireProblem(t, serveArrangements(t, arrangements, req), http.StatusUnauthorized)
			assert.Empty(t, arrangements.listed)
			assert.Empty(t, arrangements.withdrawn)
		})
	}
}
