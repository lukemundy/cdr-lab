package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const scopeTransactions = "bank:transactions:read"

type TransactionReader interface {
	// Transactions and Transaction give app.ErrAccountNotFound when the arrangement does not give the account.
	Transactions(
		ctx context.Context,
		arrangementID, accountID string,
		filter app.TransactionFilter,
	) ([]app.ConsentTransaction, int, error)
	// Transaction gives app.ErrTransactionNotFound when the account has no transaction with the ID that has detail.
	Transaction(ctx context.Context, arrangementID, accountID, id string) (app.ConsentTransaction, error)
}

// transactionJSON is a BankingTransactionV2.
type transactionJSON struct {
	AccountID            string                   `json:"accountId"`
	TransactionID        string                   `json:"transactionId"`
	IsDetailAvailable    bool                     `json:"isDetailAvailable"`
	Type                 domain.TransactionType   `json:"type"`
	Status               domain.TransactionStatus `json:"status"`
	Description          string                   `json:"description"`
	PostingDateTime      string                   `json:"postingDateTime,omitempty"`
	ValueDateTime        string                   `json:"valueDateTime,omitempty"`
	ExecutionDateTime    string                   `json:"executionDateTime,omitempty"`
	Amount               domain.Amount            `json:"amount"`
	Reference            string                   `json:"reference"`
	MerchantName         string                   `json:"merchantName,omitempty"`
	MerchantCategoryCode string                   `json:"merchantCategoryCode,omitempty"`
	BillerCode           string                   `json:"billerCode,omitempty"`
	BillerName           string                   `json:"billerName,omitempty"`
	CRN                  string                   `json:"crn,omitempty"`
	APCANumber           string                   `json:"apcaNumber,omitempty"`
}

type transactionDetailJSON struct {
	transactionJSON

	ExtendedData *domain.ExtendedData `json:"extendedData"`
}

type transactionListJSON struct {
	Transactions []transactionJSON `json:"transactions"`
}

func dateTime(at *time.Time) string {
	if at == nil {
		return ""
	}

	return at.UTC().Format(time.RFC3339)
}

func newTransactionJSON(accountID string, ct app.ConsentTransaction) transactionJSON {
	t := ct.Transaction

	return transactionJSON{
		AccountID:            accountID,
		TransactionID:        ct.ID,
		IsDetailAvailable:    t.IsDetailAvailable(),
		Type:                 t.Type,
		Status:               t.Status,
		Description:          t.Description,
		PostingDateTime:      dateTime(t.PostingDateTime),
		ValueDateTime:        dateTime(t.ValueDateTime),
		ExecutionDateTime:    dateTime(t.ExecutionDateTime),
		Amount:               t.Amount,
		Reference:            t.Reference,
		MerchantName:         t.MerchantName,
		MerchantCategoryCode: t.MerchantCategoryCode,
		BillerCode:           t.BillerCode,
		BillerName:           t.BillerName,
		CRN:                  t.CRN,
		APCANumber:           t.APCANumber,
	}
}

func transactions(baseURL string, reader TransactionReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter, filters, ok := transactionFilter(w, r)
		if !ok {
			return
		}

		page, ok := cds.ParsePage(w, r)
		if !ok {
			return
		}

		filter.Page, filter.PageSize = page.Number, page.Size
		accountID := r.PathValue("accountId")

		found, total, err := reader.Transactions(r.Context(), consentTokenOf(r).ArrangementID, accountID, filter)
		if errors.Is(err, app.ErrAccountNotFound) {
			cds.WriteError(w, http.StatusNotFound, cds.InvalidBankingAccount, accountID)
			return
		}

		if err != nil {
			cds.WriteUnexpected(w, r, err, "The data holder cannot read the transactions.")
			return
		}

		list := transactionListJSON{Transactions: make([]transactionJSON, 0, len(found))}
		for _, transaction := range found {
			list.Transactions = append(list.Transactions, newTransactionJSON(accountID, transaction))
		}

		cds.WriteTransactionPage(w, r, baseURL, page, total, filters, list)
	})
}

// transactionFilter writes the error response and gives false when a filter is not valid.
func transactionFilter(w http.ResponseWriter, r *http.Request) (app.TransactionFilter, url.Values, bool) {
	params := r.URL.Query()
	filters := url.Values{}

	for _, name := range []string{"oldest-time", "newest-time", "min-amount", "max-amount", "text"} {
		if v := params.Get(name); v != "" {
			filters.Set(name, v)
		}
	}

	filter := app.TransactionFilter{Text: filters.Get("text")}

	for _, limit := range []struct {
		name string
		at   **time.Time
	}{{"oldest-time", &filter.Oldest}, {"newest-time", &filter.Newest}} {
		if v := filters.Get(limit.name); v != "" {
			at, err := time.Parse(time.RFC3339, v)
			if err != nil {
				cds.WriteError(w, http.StatusBadRequest, cds.InvalidDateTime, limit.name)
				return app.TransactionFilter{}, nil, false
			}

			*limit.at = &at
		}
	}

	if v := filters.Get("min-amount"); v != "" {
		atLeast, _, err := domain.AmountBounds(v)
		if err != nil {
			cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "min-amount")
			return app.TransactionFilter{}, nil, false
		}

		filter.MinAmount = &atLeast
	}

	if v := filters.Get("max-amount"); v != "" {
		_, atMost, err := domain.AmountBounds(v)
		if err != nil {
			cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "max-amount")
			return app.TransactionFilter{}, nil, false
		}

		filter.MaxAmount = &atMost
	}

	return filter, filters, true
}

func transactionDetail(baseURL string, reader TransactionReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, id := r.PathValue("accountId"), r.PathValue("transactionId")

		found, err := reader.Transaction(r.Context(), consentTokenOf(r).ArrangementID, accountID, id)

		switch {
		case errors.Is(err, app.ErrAccountNotFound):
			cds.WriteError(w, http.StatusNotFound, cds.InvalidBankingAccount, accountID)
		case errors.Is(err, app.ErrTransactionNotFound):
			cds.WriteError(w, http.StatusNotFound, cds.ResourceInvalid, id)
		case err != nil:
			cds.WriteUnexpected(w, r, err, "The data holder cannot read the transaction.")
		default:
			jsonhttp.Write(w, http.StatusOK, cds.NewResponse(r, baseURL, transactionDetailJSON{
				transactionJSON: newTransactionJSON(accountID, found),
				ExtendedData:    found.Transaction.ExtendedData,
			}))
		}
	})
}
