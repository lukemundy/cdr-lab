package operatorapi

import (
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func createTransaction(customers Customers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body domain.TransactionInput
		if !decode(w, r, &body) {
			return
		}

		transaction, err := body.Transaction()
		if err != nil {
			writeError(w, r, err)
			return
		}

		created, err := customers.AddTransaction(
			r.Context(), r.PathValue("customerNumber"), r.PathValue("accountId"), transaction,
		)
		if err != nil {
			writeError(w, r, err)
			return
		}

		jsonhttp.Write(w, http.StatusCreated, created)
	})
}

func readTransactions(customers Customers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transactions, err := customers.AccountTransactions(
			r.Context(), r.PathValue("customerNumber"), r.PathValue("accountId"),
		)
		if err != nil {
			writeError(w, r, err)
			return
		}

		if transactions == nil {
			transactions = []domain.Transaction{}
		}

		jsonhttp.Write(w, http.StatusOK, transactions)
	})
}
