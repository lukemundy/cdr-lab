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

const (
	scopeAccountsBasic  = "bank:accounts.basic:read"
	scopeAccountsDetail = "bank:accounts.detail:read"
)

type AccountReader interface {
	Accounts(ctx context.Context, arrangementID string, query app.AccountQuery) ([]app.ConsentAccount, int, error)
	// Account gives app.ErrAccountNotFound when the arrangement does not give the account.
	Account(ctx context.Context, arrangementID, id string) (app.ConsentAccount, error)
}

// accountJSON is a BankingAccountV3, and each account has its customer as its only owner.
type accountJSON struct {
	AccountID        string                 `json:"accountId"`
	CreationDate     string                 `json:"creationDate"`
	DisplayName      string                 `json:"displayName"`
	MaskedNumber     string                 `json:"maskedNumber"`
	OpenStatus       domain.OpenStatus      `json:"openStatus"`
	IsOwned          bool                   `json:"isOwned"`
	AccountOwnership string                 `json:"accountOwnership"`
	ProductCategory  domain.ProductCategory `json:"productCategory"`
	ProductName      string                 `json:"productName"`
}

type accountDetailJSON struct {
	accountJSON

	BSB           string `json:"bsb"`
	AccountNumber string `json:"accountNumber"`
}

type accountListJSON struct {
	Accounts []accountJSON `json:"accounts"`
}

func newAccountJSON(a app.ConsentAccount) accountJSON {
	return accountJSON{
		AccountID:        a.ID,
		CreationDate:     a.Account.CreatedAt.UTC().Format(time.DateOnly),
		DisplayName:      a.Account.DisplayName,
		MaskedNumber:     a.Account.MaskedNumber(),
		OpenStatus:       a.Account.OpenStatus,
		IsOwned:          true,
		AccountOwnership: "ONE_PARTY",
		ProductCategory:  a.Account.ProductCategory,
		ProductName:      a.Account.ProductName,
	}
}

func accounts(baseURL string, reader AccountReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, filters, ok := accountQuery(w, r)
		if !ok {
			return
		}

		page, ok := cds.ParsePage(w, r)
		if !ok {
			return
		}

		query.Page, query.PageSize = page.Number, page.Size

		found, total, err := reader.Accounts(r.Context(), consentTokenOf(r).ArrangementID, query)
		if err != nil {
			cds.WriteUnexpected(w, r, err, "The data holder cannot read the accounts.")
			return
		}

		list := accountListJSON{Accounts: make([]accountJSON, 0, len(found))}
		for _, account := range found {
			list.Accounts = append(list.Accounts, newAccountJSON(account))
		}

		cds.WritePage(w, r, baseURL, page, total, filters, list)
	})
}

// accountQuery writes the error response and gives false when a filter is not valid.
func accountQuery(w http.ResponseWriter, r *http.Request) (app.AccountQuery, url.Values, bool) {
	params := r.URL.Query()
	filters := url.Values{}

	for _, name := range []string{"product-category", "open-status", "is-owned"} {
		if v := params.Get(name); v != "" {
			filters.Set(name, v)
		}
	}

	var query app.AccountQuery

	invalid := func(name string) (app.AccountQuery, url.Values, bool) {
		cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, name)
		return app.AccountQuery{}, nil, false
	}

	if v := filters.Get("product-category"); v != "" {
		category, valid := domain.ParseProductCategory(v)
		if !valid {
			return invalid("product-category")
		}

		query.Category = category
	}

	switch v := filters.Get("open-status"); v {
	case "", "ALL":
	case string(domain.OpenStatusOpen), string(domain.OpenStatusClosed):
		query.OpenStatus = domain.OpenStatus(v)
	default:
		return invalid("open-status")
	}

	switch filters.Get("is-owned") {
	case "":
	case "true":
		query.IsOwned = new(true)
	case "false":
		query.IsOwned = new(false)
	default:
		return invalid("is-owned")
	}

	return query, filters, true
}

func accountDetail(baseURL string, reader AccountReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("accountId")

		account, err := reader.Account(r.Context(), consentTokenOf(r).ArrangementID, id)
		if errors.Is(err, app.ErrAccountNotFound) {
			cds.WriteError(w, http.StatusNotFound, cds.InvalidBankingAccount, id)
			return
		}

		if err != nil {
			cds.WriteUnexpected(w, r, err, "The data holder cannot read the account.")
			return
		}

		jsonhttp.Write(w, http.StatusOK, cds.NewResponse(r, baseURL, accountDetailJSON{
			accountJSON:   newAccountJSON(account),
			BSB:           account.Account.BSB,
			AccountNumber: account.Account.AccountNumber,
		}))
	})
}
