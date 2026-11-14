package app

import (
	"context"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type CustomerSort string

const (
	SortByLastName         CustomerSort = "last-name"
	SortByCustomerNumber   CustomerSort = "customer-number"
	SortByLastUpdated      CustomerSort = "last-updated"
	SortByOpenAccountCount CustomerSort = "open-account-count"
	// The arrangements of a customer that did not end at CustomerPageQuery.At.
	SortByCurrentArrangementCount CustomerSort = "current-arrangement-count"
)

type CustomerPageQuery struct {
	// Contains matches a part of the customer number, the first name, the last name or the email address, with no
	// difference for case. It is empty for each customer.
	Contains   string
	Sort       CustomerSort
	Descending bool
	// Page starts at 1.
	Page     int
	PageSize int
	// At is the time at which an arrangement that ends at or before it has ended.
	At time.Time
}

type CustomerSummary struct {
	Customer            domain.Customer
	OpenAccounts        int
	CurrentArrangements int
}

// CustomerPageReader sorts by the customer number after the sort of the query, so that each page is stable.
type CustomerPageReader interface {
	CustomerPage(ctx context.Context, query CustomerPageQuery) (page []CustomerSummary, total int, err error)
}
