package banksite_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const fieldErrorsType = "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002"

// opener opens the account opened for the customer signedIn into the list of reader, and keeps the product ID of the
// request.
type opener struct {
	reader    *accounts
	opened    app.AccountDetail
	productID string
	input     domain.OpeningInput
	err       error
}

func (o *opener) OpenAccount(
	_ context.Context,
	customerNumber, productID string,
	input domain.OpeningInput,
) (domain.Account, error) {
	o.productID, o.input = productID, input

	if customerNumber != signedIn {
		return domain.Account{}, app.ErrCustomerNotFound
	}

	if o.err != nil {
		return domain.Account{}, o.err
	}

	o.reader.list = append(o.reader.list, o.opened)

	return o.opened.Account, nil
}

func opened() app.AccountDetail {
	detail := everyday()
	detail.Account.CreditLimit = 0
	detail.Balance = domain.Balance{}

	return detail
}

func openSite(t *testing.T, o *opener) *site {
	t.Helper()

	if o.reader == nil {
		o.reader = &accounts{}
	}

	return newSiteWith(t, banksite.Ports{
		Customers: newCustomers(domain.Customer{CustomerNumber: signedIn, LastName: "Quoll"}),
		Opener:    o, Accounts: o.reader,
	})
}

func signedInOpen(t *testing.T, o *opener, body string) response {
	t.Helper()

	s := openSite(t, o)
	cookie := sessionCookie(t, s.signIn(signedIn))

	return s.do(request{
		method: http.MethodPost, target: "/api/accounts", body: body,
		cookies: []*http.Cookie{requestCookie(cookieName, cookie.Value)},
	})
}

// Verifies: DH-1761
func TestAnOpenGivesTheNewAccountWithTheFieldsOfAnAccountOfTheList(t *testing.T) {
	o := &opener{opened: opened()}

	resp := signedInOpen(t, o, `{"productId": "saver"}`)

	require.Equal(t, http.StatusCreated, resp.status)
	assert.Equal(t, "saver", o.productID)
	assert.Empty(t, resp.header.Get("Location"))
	assert.JSONEq(t, `{
		"accountId": "account-1", "creationDate": "2026-10-01", "displayName": "Everyday",
		"maskedNumber": "xxx-xxx xxxx5678", "openStatus": "OPEN", "productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
		"productName": "Saver",
		"balance": {"currentBalance": "0.00", "availableBalance": "0.00", "creditLimit": "0.00"},
		"actions": [{"action": "close", "ends": []}]
	}`, resp.body)
}

func TestAFailureOfTheReadAfterAnOpenIsAServerError(t *testing.T) {
	resp := signedInOpen(t, &opener{opened: opened(), reader: &accounts{err: errors.New("no database")}},
		`{"productId": "saver"}`)

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "no database")
}

// Verifies: DH-1743
func TestAnOpenNeedsASession(t *testing.T) {
	o := &opener{opened: opened(), reader: &accounts{}}
	s := newSiteWith(t, banksite.Ports{Customers: newCustomers(), Opener: o, Accounts: o.reader})

	resp := s.do(request{method: http.MethodPost, target: "/api/accounts", body: `{"productId": "saver"}`})

	assert.Equal(t, http.StatusForbidden, resp.status)
	assert.Empty(t, o.productID, "no open")
}

func TestAnOpenGivesTheCreditLimitAndTheLoanAmountOfTheRequestToTheOpener(t *testing.T) {
	o := &opener{opened: opened()}

	resp := signedInOpen(t, o, `{"productId": "card", "creditLimit": "5000.00", "loanAmount": "1.00"}`)

	require.Equal(t, http.StatusCreated, resp.status)
	assert.Equal(t, domain.OpeningInput{CreditLimit: new("5000.00"), LoanAmount: new("1.00")}, o.input)
}

// Verifies: PROB-0002, DH-1759, DH-1760
func TestAnOpenThatTheDataHolderRefusesGivesAFieldErrorInPlainLanguage(t *testing.T) {
	tests := map[string]struct {
		err     error
		pointer string
		want    string
	}{
		"product that is not current or does not exist": {
			err: app.ErrProductNotFound, pointer: "#/productId",
			want: "The bank has no current product with this ID. Select a product from the list of products.",
		},
		"credit limit that is not valid": {
			err:     domain.ValidationError{Errors: []domain.FieldError{{Field: "creditLimit", Detail: "x"}}},
			pointer: "#/creditLimit", want: "Enter a credit limit of more than $0.00.",
		},
		"loan amount that is not valid": {
			err:     domain.ValidationError{Errors: []domain.FieldError{{Field: "loanAmount", Detail: "x"}}},
			pointer: "#/loanAmount", want: "Enter a loan amount of more than $0.00.",
		},
		"credit limit of a product with no credit limit": {
			err: domain.ValueNotForProductError{Field: "creditLimit"}, pointer: "#/creditLimit",
			want: "This product has no credit limit.",
		},
		"loan amount of a product that is not a loan": {
			err: domain.ValueNotForProductError{Field: "loanAmount"}, pointer: "#/loanAmount",
			want: "This product is not a loan, so it has no loan amount.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := signedInOpen(t, &opener{err: tt.err}, `{"productId": "loan"}`)

			require.Equal(t, http.StatusBadRequest, resp.status)
			assert.Contains(t, resp.body, `"type":"`+fieldErrorsType+`"`)
			assert.JSONEq(t, `[{"pointer": "`+tt.pointer+`", "detail": "`+tt.want+`"}]`, errorsOf(t, resp))
		})
	}
}

// Verifies: DH-1719
func TestAnOpenWithABodyThatIsNotValidGives400AndOpensNoAccount(t *testing.T) {
	for name, body := range map[string]string{
		"not JSON":      `productId=saver`,
		"null":          `null`,
		"unknown field": `{"productId": "saver", "displayName": "Mine"}`,
		"number amount": `{"productId": "card", "creditLimit": 5000}`,
	} {
		t.Run(name, func(t *testing.T) {
			o := &opener{opened: opened()}

			resp := signedInOpen(t, o, body)

			assert.Equal(t, http.StatusBadRequest, resp.status)
			assert.Empty(t, o.productID, "no open")
		})
	}
}

func TestAFailureOfAnOpenIsAServerError(t *testing.T) {
	resp := signedInOpen(t, &opener{err: errors.New("no database")}, `{"productId": "saver"}`)

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "no database")
}

// Verifies: DH-1744
func TestAnOpenForACustomerThatTheDataHolderNoLongerHasEndsTheSession(t *testing.T) {
	const gone = "00000002"

	o := &opener{opened: opened(), reader: &accounts{}}
	s := newSiteWith(t, banksite.Ports{
		Customers: newCustomers(domain.Customer{CustomerNumber: gone, LastName: "Quoll"}),
		Opener:    o, Accounts: o.reader,
	})
	cookie := requestCookie(cookieName, sessionCookie(t, s.signIn(gone)).Value)

	resp := s.do(request{
		method: http.MethodPost, target: "/api/accounts", body: `{"productId": "saver"}`,
		cookies: []*http.Cookie{cookie},
	})

	require.Equal(t, http.StatusForbidden, resp.status)
	assert.Contains(t, resp.body, `"type":"`+sessionEndedType+`"`)
	assertExpired(t, resp)
	assert.Equal(t, http.StatusForbidden, s.readSession(cookie).status, "the session ended")
}

// The UI gives the field of an open from these categories, so the document must give the rule of the data holder.
// Verifies: DH-1732, DH-2213, DH-2214
func TestTheDocumentGivesTheCategoriesOfACreditLimitAndOfALoanAmount(t *testing.T) {
	document, err := openapi3.NewLoader().LoadFromFile("openapi.yaml")
	require.NoError(t, err)

	for name, want := range map[string][]domain.ProductCategory{
		"RevolvingProductCategory": domain.RevolvingCategories(),
		"LoanProductCategory":      domain.LoanCategories(),
	} {
		schema := document.Components.Schemas[name]
		require.NotNil(t, schema, name)

		var got []domain.ProductCategory
		for _, value := range schema.Value.Enum {
			got = append(got, domain.ProductCategory(value.(string)))
		}

		assert.ElementsMatch(t, want, got, name)
	}
}
