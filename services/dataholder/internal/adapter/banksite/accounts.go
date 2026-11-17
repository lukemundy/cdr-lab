package banksite

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// AccountReader gives the account data of one customer with the functions of the APIs of a consent.
type AccountReader interface {
	Accounts(ctx context.Context, customerNumber string) (app.CustomerAccounts, error)
	// Balance and Transactions give app.ErrAccountNotFound when the account is not of the customer.
	Balance(ctx context.Context, customerNumber, accountID string) (domain.Balance, error)
	Transactions(
		ctx context.Context,
		customerNumber, accountID string,
		page, pageSize int,
	) ([]domain.Transaction, int, error)
}

// AccountOpener gives app.ErrProductNotFound for a product that is not current, domain.ValueNotForProductError and
// domain.ValidationError for the input, and app.ErrCustomerNotFound.
type AccountOpener interface {
	OpenAccount(
		ctx context.Context,
		customerNumber, productID string,
		input domain.OpeningInput,
	) (domain.Account, error)
}

// The JSON of the account data has the names and the values of the CDS schemas, so that a test compares the site
// with the APIs of a consent with no conversion. The identifiers are those of the data holder.

// balanceJSON is a BankingBalance with no accountId.
type balanceJSON struct {
	CurrentBalance   domain.Amount `json:"currentBalance"`
	AvailableBalance domain.Amount `json:"availableBalance"`
	CreditLimit      domain.Amount `json:"creditLimit"`
}

// accountJSON has the fields of BankingAccountV3 that the site shows, and the balance of the account.
type accountJSON struct {
	AccountID       string                 `json:"accountId"`
	CreationDate    string                 `json:"creationDate"`
	DisplayName     string                 `json:"displayName"`
	MaskedNumber    string                 `json:"maskedNumber"`
	OpenStatus      domain.OpenStatus      `json:"openStatus"`
	ProductCategory domain.ProductCategory `json:"productCategory"`
	ProductName     string                 `json:"productName"`
	Balance         balanceJSON            `json:"balance"`
	Actions         []actionJSON           `json:"actions"`
}

// The site gives only the close: the other actions of an account are for the operator.
type actionJSON struct {
	Action domain.AccountAction `json:"action"`
	Ends   []endJSON            `json:"ends"`
}

type endJSON struct {
	ArrangementID string `json:"arrangementId"`
	ClientName    string `json:"clientName"`
}

type accountListJSON struct {
	Accounts []accountJSON `json:"accounts"`
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

type transactionListJSON struct {
	Transactions []transactionJSON `json:"transactions"`
	TotalRecords int               `json:"totalRecords"`
	TotalPages   int               `json:"totalPages"`
}

func newBalanceJSON(b domain.Balance) balanceJSON {
	return balanceJSON{CurrentBalance: b.Current, AvailableBalance: b.Available, CreditLimit: b.CreditLimit}
}

func closeActionsOf(permitted []domain.Action[domain.AccountAction]) []actionJSON {
	actions := []actionJSON{}

	for _, action := range permitted {
		if action.Action != domain.AccountActionClose {
			continue
		}

		ends := make([]endJSON, 0, len(action.Ends))
		for _, e := range action.Ends {
			ends = append(
				ends,
				endJSON{ArrangementID: e.Arrangement.ArrangementID, ClientName: e.Arrangement.ClientName},
			)
		}

		actions = append(actions, actionJSON{Action: action.Action, Ends: ends})
	}

	return actions
}

func newAccountJSON(detail app.AccountDetail, actions []actionJSON) accountJSON {
	a := detail.Account

	return accountJSON{
		AccountID:       a.AccountID,
		CreationDate:    a.CreatedAt.UTC().Format(time.DateOnly),
		DisplayName:     a.DisplayName,
		MaskedNumber:    a.MaskedNumber(),
		OpenStatus:      a.OpenStatus,
		ProductCategory: a.ProductCategory,
		ProductName:     a.ProductName,
		Balance:         newBalanceJSON(detail.Balance),
		Actions:         actions,
	}
}

func dateTime(at *time.Time) string {
	if at == nil {
		return ""
	}

	return at.UTC().Format(time.RFC3339)
}

func newTransactionJSON(accountID string, t domain.Transaction) transactionJSON {
	return transactionJSON{
		AccountID:            accountID,
		TransactionID:        t.TransactionID,
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

// writeAccountError gives no identifier in the detail, because a path value does not go into a response. A session
// of a customer that the data holder no longer has ends, as a session that ended.
func writeAccountError(w http.ResponseWriter, r *http.Request, s sessions, err error) {
	switch {
	case errors.Is(err, app.ErrAccountNotFound):
		problem.Write(w, http.StatusNotFound, "You have no account with this ID.")
	case errors.Is(err, app.ErrCustomerNotFound):
		s.endOf(r)
		s.refuse(w, r)
	default:
		problem.WriteUnexpected(w, r, err, notCompleted)
	}
}

func listAccounts(reader AccountReader, s sessions, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerNumber, ok := s.customerOf(r)
		if !ok {
			s.refuse(w, r)
			return
		}

		found, err := reader.Accounts(r.Context(), customerNumber)
		if err != nil {
			writeAccountError(w, r, s, err)
			return
		}

		jsonhttp.Write(w, http.StatusOK, accountListOf(found, now()))
	})
}

func accountListOf(found app.CustomerAccounts, now time.Time) accountListJSON {
	accounts := make([]domain.Account, 0, len(found.Accounts))
	for _, detail := range found.Accounts {
		accounts = append(accounts, detail.Account)
	}

	permitted := domain.PermittedAccountActions(accounts, found.Arrangements, now)

	list := accountListJSON{Accounts: make([]accountJSON, 0, len(found.Accounts))}
	for _, detail := range found.Accounts {
		list.Accounts = append(
			list.Accounts, newAccountJSON(detail, closeActionsOf(permitted[detail.Account.AccountID])),
		)
	}

	return list
}

func readBalance(reader AccountReader, s sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerNumber, ok := s.customerOf(r)
		if !ok {
			s.refuse(w, r)
			return
		}

		balance, err := reader.Balance(r.Context(), customerNumber, r.PathValue("accountId"))
		if err != nil {
			writeAccountError(w, r, s, err)
			return
		}

		jsonhttp.Write(w, http.StatusOK, newBalanceJSON(balance))
	})
}

func listTransactions(reader AccountReader, s sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerNumber, ok := s.customerOf(r)
		if !ok {
			s.refuse(w, r)
			return
		}

		page, err := cds.ReadPage(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		accountID := r.PathValue("accountId")

		found, total, err := reader.Transactions(r.Context(), customerNumber, accountID, page.Number, page.Size)
		if err != nil {
			writeAccountError(w, r, s, err)
			return
		}

		body := transactionListJSON{
			Transactions: make([]transactionJSON, 0, len(found)),
			TotalRecords: total,
			TotalPages:   (total + page.Size - 1) / page.Size,
		}
		for _, t := range found {
			body.Transactions = append(body.Transactions, newTransactionJSON(accountID, t))
		}

		jsonhttp.Write(w, http.StatusOK, body)
	})
}

const (
	// A customer has few arrangements, and each ID is short.
	maxCloseBodyBytes = 1 << 14
	badCloseBody      = "The request body must be a JSON object with only ends, the list of the IDs of the " +
		"arrangements that the confirmation of the close named."
	endsChanged = "Your accounts or your consents changed after you opened this confirmation. We closed nothing. " +
		"Check the account and try again."
)

func closeAccount(customers Customers, s sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerNumber, ok := s.customerOf(r)
		if !ok {
			s.refuse(w, r)
			return
		}

		ends, err := decodeEnds(http.MaxBytesReader(w, r.Body, maxCloseBodyBytes))
		if err != nil {
			problem.WriteBodyRefusal(w, err, badCloseBody)
			return
		}

		_, err = customers.CloseAccountThatEnds(r.Context(), customerNumber, r.PathValue("accountId"), ends)

		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, app.ErrEndsChanged):
			problem.Write(w, http.StatusConflict, endsChanged)
		case errors.Is(err, domain.ErrAccountClosed):
			problem.Write(w, http.StatusConflict, "This account is already closed.")
		default:
			writeAccountError(w, r, s, err)
		}
	})
}

var errNoEnds = errors.New("the body has no ends")

func decodeEnds(body io.Reader) ([]string, error) {
	var request *struct {
		Ends *[]string `json:"ends"`
	}

	if err := jsonhttp.Decode(body, &request); err != nil {
		return nil, err
	}

	if request == nil || request.Ends == nil {
		return nil, errNoEnds
	}

	return *request.Ends, nil
}

const badOpenBody = "The request body must be a JSON object with productId, and optionally creditLimit or " +
	"loanAmount as strings."

var errOpenedNotListed = errors.New("the accounts of the customer do not have the opened account")

type openRequest struct {
	ProductID string `json:"productId"`
	domain.OpeningInput
}

// openAccount gives the new account as the list of the accounts gives it, also with its actions.
func openAccount(opener AccountOpener, reader AccountReader, s sessions, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerNumber, ok := s.customerOf(r)
		if !ok {
			s.refuse(w, r)
			return
		}

		var request *openRequest

		err := jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxBodyBytes), &request)
		if err != nil || request == nil {
			problem.WriteBodyRefusal(w, err, badOpenBody)

			return
		}

		opened, err := opener.OpenAccount(r.Context(), customerNumber, request.ProductID, request.OpeningInput)
		if err != nil {
			writeOpenError(w, r, s, err)
			return
		}

		found, err := reader.Accounts(r.Context(), customerNumber)
		if err != nil {
			writeAccountError(w, r, s, err)
			return
		}

		for _, account := range accountListOf(found, now()).Accounts {
			if account.AccountID == opened.AccountID {
				jsonhttp.Write(w, http.StatusCreated, account)
				return
			}
		}

		problem.WriteUnexpected(w, r, errOpenedNotListed, notCompleted)
	})
}

// The texts name the fields with the labels of the open page.
var (
	openValueTexts = map[string]string{
		"creditLimit": "Enter a credit limit of more than $0.00.",
		"loanAmount":  "Enter a loan amount of more than $0.00.",
	}
	openValueNotForProductTexts = map[string]string{
		"creditLimit": "This product has no credit limit.",
		"loanAmount":  "This product is not a loan, so it has no loan amount.",
	}
)

// writeOpenError gives texts for a consumer, who selects a product from a page and does not type its ID.
func writeOpenError(w http.ResponseWriter, r *http.Request, s sessions, err error) {
	var (
		invalid       domain.ValidationError
		notForProduct domain.ValueNotForProductError
		errs          []problem.FieldError
	)

	switch {
	case errors.Is(err, app.ErrProductNotFound):
		errs = []problem.FieldError{{
			Pointer: problem.FieldPointer("productId"),
			Detail:  "The bank has no current product with this ID. Select a product from the list of products.",
		}}
	case errors.As(err, &notForProduct):
		field := notForProduct.Field
		errs = []problem.FieldError{{
			Pointer: problem.FieldPointer(field), Detail: openValueNotForProductTexts[field],
		}}
	case errors.As(err, &invalid) && onlyConsumerValues(invalid):
		for _, f := range invalid.Errors {
			errs = append(
				errs,
				problem.FieldError{Pointer: problem.FieldPointer(f.Field), Detail: openValueTexts[f.Field]},
			)
		}
	default:
		writeAccountError(w, r, s, err)
		return
	}

	problem.WriteFieldErrors(w, http.StatusBadRequest, "", errs)
}

// onlyConsumerValues is false for an error of a value that the data holder takes from the product.
func onlyConsumerValues(invalid domain.ValidationError) bool {
	for _, f := range invalid.Errors {
		if _, ok := openValueTexts[f.Field]; !ok {
			return false
		}
	}

	return true
}
