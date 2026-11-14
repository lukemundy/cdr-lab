package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const maxBalancesRequestBytes = 64 << 10

type BalanceReader interface {
	// Balance gives app.ErrAccountNotFound when the arrangement does not give the account.
	Balance(ctx context.Context, arrangementID, id string) (app.ConsentBalance, error)
	// Balances gives an app.UnknownAccountsError when query.IDs has an account that the arrangement does not give.
	Balances(ctx context.Context, arrangementID string, query app.AccountQuery) ([]app.ConsentBalance, int, error)
}

// balanceJSON is a BankingBalance.
type balanceJSON struct {
	AccountID        string        `json:"accountId"`
	CurrentBalance   domain.Amount `json:"currentBalance"`
	AvailableBalance domain.Amount `json:"availableBalance"`
	CreditLimit      domain.Amount `json:"creditLimit"`
}

type balanceListJSON struct {
	Balances []balanceJSON `json:"balances"`
}

// accountIDsRequest has pointers, so that an absent field is different from an empty field.
type accountIDsRequest struct {
	Data *struct {
		AccountIDs *[]string `json:"accountIds"`
	} `json:"data"`
}

func newBalanceJSON(b app.ConsentBalance) balanceJSON {
	return balanceJSON{
		AccountID:        b.ID,
		CurrentBalance:   b.Balance.Current,
		AvailableBalance: b.Balance.Available,
		CreditLimit:      b.Balance.CreditLimit,
	}
}

func newBalanceListJSON(balances []app.ConsentBalance) balanceListJSON {
	list := balanceListJSON{Balances: make([]balanceJSON, 0, len(balances))}
	for _, b := range balances {
		list.Balances = append(list.Balances, newBalanceJSON(b))
	}

	return list
}

func accountBalance(baseURL string, reader BalanceReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("accountId")

		balance, err := reader.Balance(r.Context(), consentTokenOf(r).ArrangementID, id)
		if errors.Is(err, app.ErrAccountNotFound) {
			cds.WriteError(w, http.StatusNotFound, cds.InvalidBankingAccount, id)
			return
		}

		if err != nil {
			cds.WriteUnexpected(w, r, err, "The data holder cannot read the balance.")
			return
		}

		jsonhttp.Write(w, http.StatusOK, cds.NewResponse(r, baseURL, newBalanceJSON(balance)))
	})
}

func bulkBalances(baseURL string, reader BalanceReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, filters, ok := accountQuery(w, r)
		if !ok {
			return
		}

		writeBalancePage(w, r, baseURL, reader, query, filters)
	})
}

func specificBalances(baseURL string, reader BalanceReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			cds.WriteError(
				w, http.StatusUnsupportedMediaType, cds.Expected, "The Content-Type must be application/json.",
			)

			return
		}

		ids, ok := decodeAccountIDs(w, r)
		if !ok {
			return
		}

		writeBalancePage(w, r, baseURL, reader, app.AccountQuery{IDs: ids}, nil)
	})
}

func writeBalancePage(
	w http.ResponseWriter,
	r *http.Request,
	baseURL string,
	reader BalanceReader,
	query app.AccountQuery,
	filters url.Values,
) {
	page, ok := cds.ParsePage(w, r)
	if !ok {
		return
	}

	query.Page, query.PageSize = page.Number, page.Size

	found, total, err := reader.Balances(r.Context(), consentTokenOf(r).ArrangementID, query)

	if unknown, ok := errors.AsType[app.UnknownAccountsError](err); ok {
		cds.WriteErrors(w, http.StatusUnprocessableEntity, cds.InvalidBankingAccount, unknown.IDs...)
		return
	}

	if err != nil {
		cds.WriteUnexpected(w, r, err, "The data holder cannot read the balances.")
		return
	}

	cds.WritePage(w, r, baseURL, page, total, filters, newBalanceListJSON(found))
}

// decodeAccountIDs writes the error response and gives false when the body is not a RequestAccountIds.
func decodeAccountIDs(w http.ResponseWriter, r *http.Request) ([]string, bool) {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBalancesRequestBytes))

	var body accountIDsRequest

	err := decoder.Decode(&body)
	if err == nil && decoder.More() {
		err = errors.New("more than one JSON value")
	}

	var (
		tooLarge  *http.MaxBytesError
		wrongType *json.UnmarshalTypeError
	)

	switch {
	case errors.As(err, &tooLarge):
		cds.WriteError(
			w, http.StatusRequestEntityTooLarge, cds.Expected, "The request body must be 64 KiB or less.",
		)
	case errors.As(err, &wrongType) && wrongType.Field != "":
		cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, wrongType.Field)
	case err != nil:
		cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "data")
	case body.Data == nil:
		cds.WriteError(w, http.StatusBadRequest, cds.FieldMissing, "data")
	case body.Data.AccountIDs == nil:
		cds.WriteError(w, http.StatusBadRequest, cds.FieldMissing, "data.accountIds")
	default:
		return *body.Data.AccountIDs, true
	}

	return nil, false
}
