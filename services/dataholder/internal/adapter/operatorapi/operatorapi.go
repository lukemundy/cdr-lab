// Package operatorapi serves the operator interface of the data holder.
package operatorapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// A customer, an account or a transaction is a few kilobytes at most.
const maxRequestBytes = 64 << 10

// Customers gives a domain.ValidationError for a record that is not valid, and the errors of app.CustomerStore.
type Customers interface {
	Create(ctx context.Context, customer domain.Customer) (domain.Customer, error)
	AddAccount(ctx context.Context, customerNumber string, account domain.Account) (domain.Account, error)
	CloseAccount(ctx context.Context, customerNumber, accountID string) (domain.Account, error)
	AddTransaction(
		ctx context.Context,
		customerNumber, accountID string,
		transaction domain.Transaction,
	) (domain.Transaction, error)
	AccountTransactions(ctx context.Context, customerNumber, accountID string) ([]domain.Transaction, error)
	Customer(ctx context.Context, customerNumber string) (app.CustomerRecord, error)
	Load(ctx context.Context, load domain.CustomerLoad) ([]app.LoadResult, error)
}

// The data holder creates the customer number, so a request does not have it.
type customerRequest struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
}

type customerJSON struct {
	domain.Customer

	Accounts []domain.Account `json:"accounts"`
}

func Register(mux *http.ServeMux, customers Customers, arrangements Arrangements, access operatorhttp.Config) {
	routes := httproute.New(mux)
	// The checks apply to each route, not to the mux, so that the telemetry of a refused request has its route.
	routes.Use(operatorhttp.Guard(access))

	routes.Handle("POST /v1/customers", createCustomer(customers))
	routes.Handle("GET /v1/customers/{customerNumber}", readCustomer(customers))
	routes.Handle("POST /v1/customers/{customerNumber}/accounts", createAccount(customers))
	routes.Handle("POST /v1/customers/{customerNumber}/accounts/{accountId}/close", closeAccount(customers))
	routes.Handle("POST /v1/customers/{customerNumber}/accounts/{accountId}/transactions", createTransaction(customers))
	routes.Handle("GET /v1/customers/{customerNumber}/accounts/{accountId}/transactions", readTransactions(customers))
	routes.Handle("GET /v1/customers/{customerNumber}/arrangements", readArrangements(arrangements))
	routes.Handle("POST /v1/customer-loads", loadCustomers(customers))
	routes.Handle(
		"POST /v1/customers/{customerNumber}/arrangements/{arrangementId}/withdraw",
		withdrawArrangement(arrangements),
	)
	routes.Handle("/v1/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		problem.Write(w, http.StatusNotFound, "The URL and method are not an operator action.")
	}))
}

func createCustomer(customers Customers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body customerRequest
		if !decode(w, r, &body) {
			return
		}

		created, err := customers.Create(r.Context(), domain.Customer{
			FirstName:    body.FirstName,
			LastName:     body.LastName,
			EmailAddress: body.EmailAddress,
		})
		if err != nil {
			writeError(w, r, err)
			return
		}

		jsonhttp.Write(w, http.StatusCreated, created)
	})
}

func createAccount(customers Customers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body domain.AccountInput
		if !decode(w, r, &body) {
			return
		}

		account, err := body.Account()
		if err != nil {
			writeError(w, r, err)
			return
		}

		created, err := customers.AddAccount(r.Context(), r.PathValue("customerNumber"), account)
		if err != nil {
			writeError(w, r, err)
			return
		}

		jsonhttp.Write(w, http.StatusCreated, created)
	})
}

func closeAccount(customers Customers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		closed, err := customers.CloseAccount(r.Context(), r.PathValue("customerNumber"), r.PathValue("accountId"))
		if err != nil {
			writeError(w, r, err)
			return
		}

		jsonhttp.Write(w, http.StatusOK, closed)
	})
}

func readCustomer(customers Customers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, err := customers.Customer(r.Context(), r.PathValue("customerNumber"))
		if err != nil {
			writeError(w, r, err)
			return
		}

		accounts := record.Accounts
		if accounts == nil {
			accounts = []domain.Account{}
		}

		jsonhttp.Write(w, http.StatusOK, customerJSON{Customer: record.Customer, Accounts: accounts})
	})
}

// decode writes the response when the body is not one JSON object with only the fields of into.
func decode(w http.ResponseWriter, r *http.Request, into any) bool {
	return decodeUpTo(w, r, into, maxRequestBytes)
}

func decodeUpTo(w http.ResponseWriter, r *http.Request, into any, limit int64) bool {
	err := jsonhttp.Decode(http.MaxBytesReader(w, r.Body, limit), into)
	if err == nil {
		return true
	}

	detail := "The request body must be one JSON object with only the fields of this action."

	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		// The other decoder errors name a field or a type, and have no value of the body.
		detail += " " + err.Error() + "."
	}

	problem.WriteBodyRefusal(w, err, detail)

	return false
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, detail := problemOf(err)
	if status == 0 {
		problem.WriteUnexpected(w, r, err, "The data holder could not complete the request.")
		return
	}

	if record, ok := errors.AsType[app.LoadRecordError](err); ok {
		detail = record.Position + ": " + detail
	}

	problem.Write(w, status, detail)
}

// problemOf gives the status 0 for an error that the operator cannot correct.
func problemOf(err error) (int, string) {
	var invalid domain.ValidationError

	switch {
	case errors.As(err, &invalid):
		return http.StatusBadRequest, invalid.Error()
	case errors.Is(err, app.ErrCustomerNotFound):
		return http.StatusNotFound, "The data holder has no customer with this customer number."
	case errors.Is(err, app.ErrEmailAddressUsed):
		return http.StatusConflict, "A customer has this emailAddress."
	case errors.Is(err, app.ErrAccountExists):
		return http.StatusConflict, "An account has this bsb and accountNumber."
	case errors.Is(err, app.ErrAccountNotFound):
		return http.StatusNotFound, "The customer has no account with this accountId."
	case errors.Is(err, domain.ErrAccountClosed):
		return http.StatusConflict, "The account is closed."
	case errors.Is(err, domain.ErrArrangementNotFound):
		return http.StatusNotFound, "The customer has no arrangement with this arrangementId."
	case errors.Is(err, domain.ErrArrangementEnded):
		return http.StatusConflict, "The arrangement ended."
	default:
		return 0, ""
	}
}
