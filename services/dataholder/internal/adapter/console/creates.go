package console

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// Creator gives a domain.ValidationError for a record that is not valid. Create also gives app.ErrEmailAddressUsed.
// AddAccount also gives
// app.ErrCustomerNotFound and app.ErrAccountExists. AddTransaction also gives app.ErrCustomerNotFound,
// app.ErrAccountNotFound and domain.ErrAccountClosed.
type Creator interface {
	Create(ctx context.Context, customer domain.Customer) (domain.Customer, error)
	AddAccount(ctx context.Context, customerNumber string, account domain.Account) (domain.Account, error)
	AddTransaction(
		ctx context.Context, customerNumber, accountID string, transaction domain.Transaction,
	) (domain.Transaction, error)
}

const (
	fieldErrorsType = "https://github.com/lukemundy/cdr-lab/blob/main/services/dataholder/DESIGN.md#console-api"
	// A customer, an account or a transaction is a few kilobytes at most.
	maxCreateBodyBytes = 64 << 10
	badCustomerBody    = "The request body must be a JSON object with only the fields of the customer."
	badAccountBody     = "The request body must be a JSON object with only the fields of the account."
)

// The data holder creates the customer number, so a request does not have it.
type customerCreate struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
}

func createCustomer(creator Creator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A pointer is nil after the JSON null, which decodes into a struct with no error.
		var request *customerCreate
		if jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxCreateBodyBytes), &request) != nil || request == nil {
			problem.Write(w, http.StatusBadRequest, badCustomerBody)
			return
		}

		created, err := creator.Create(r.Context(), domain.Customer{
			FirstName: request.FirstName, LastName: request.LastName, EmailAddress: request.EmailAddress,
		})
		if err != nil {
			writeCreateError(w, r, err)
			return
		}

		w.Header().Set("Location", "/api/customers/"+url.PathEscape(created.CustomerNumber))
		w.WriteHeader(http.StatusCreated)
	})
}

func createAccount(creator Creator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request *domain.AccountInput
		if jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxCreateBodyBytes), &request) != nil || request == nil {
			problem.Write(w, http.StatusBadRequest, badAccountBody)
			return
		}

		account, err := request.Account()
		if err != nil {
			writeCreateError(w, r, err)
			return
		}

		customerNumber := r.PathValue("customerNumber")

		created, err := creator.AddAccount(r.Context(), customerNumber, account)
		if err != nil {
			writeCreateError(w, r, err)
			return
		}

		w.Header().Set("Location",
			"/api/customers/"+url.PathEscape(customerNumber)+"/accounts/"+url.PathEscape(created.AccountID))
		w.WriteHeader(http.StatusCreated)
	})
}

func writeCreateError(w http.ResponseWriter, r *http.Request, err error) {
	var conflict problem.FieldError

	switch {
	case errors.Is(err, app.ErrAccountExists):
		conflict = problem.FieldError{
			Pointer: problem.FieldPointer("accountNumber"), Detail: "An account has this bsb and accountNumber.",
		}
	case errors.Is(err, app.ErrEmailAddressUsed):
		conflict = problem.FieldError{
			Pointer: problem.FieldPointer("emailAddress"),
			Detail:  "A customer has this emailAddress.",
		}
	}

	if conflict.Pointer != "" {
		problem.WriteFieldErrors(w, fieldErrorsType, http.StatusConflict, "", []problem.FieldError{conflict})
		return
	}

	invalid, ok := errors.AsType[domain.ValidationError](err)
	if !ok {
		writeDetailError(w, r, err)
		return
	}

	errs := make([]problem.FieldError, 0, len(invalid.Errors))
	for _, f := range invalid.Errors {
		errs = append(errs, problem.FieldError{Pointer: problem.FieldPointer(f.Field), Detail: f.Detail})
	}

	problem.WriteFieldErrors(w, fieldErrorsType, http.StatusBadRequest, "", errs)
}
