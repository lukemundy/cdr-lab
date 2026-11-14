package cds

import (
	"errors"
	"maps"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
)

const (
	defaultPageSize = 25
	maxPageSize     = 1000
)

type Page struct {
	// Number starts at 1.
	Number int
	Size   int
}

type PageError struct {
	Parameter string
	error     Error
	text      string
}

func (e PageError) Error() string { return e.text }

func ReadPage(params url.Values) (Page, error) {
	page := Page{Number: 1, Size: defaultPageSize}

	if v := params.Get("page"); v != "" {
		number, valid := PositiveInteger(v)
		if !valid {
			return Page{}, PageError{
				Parameter: "page", error: FieldInvalid, text: "page must be an integer of 1 or more",
			}
		}
		// A larger page is also out of range, and its row offset would be too large for the database.
		page.Number = min(number, math.MaxInt32)
	}

	if v := params.Get("page-size"); v != "" {
		size, valid := PositiveInteger(v)
		if !valid || size > maxPageSize {
			return Page{}, PageError{
				Parameter: "page-size", error: InvalidPageSize,
				text: "page-size must be an integer from 1 to " + strconv.Itoa(maxPageSize),
			}
		}

		page.Size = size
	}

	return page, nil
}

// ParsePage writes the error response and gives false when page or page-size is not valid.
func ParsePage(w http.ResponseWriter, r *http.Request) (Page, bool) {
	page, err := ReadPage(r.URL.Query())
	if pageErr, ok := errors.AsType[PageError](err); ok {
		WriteError(w, http.StatusBadRequest, pageErr.error, pageErr.Parameter)
		return Page{}, false
	}

	return page, true
}

type paginatedJSON[T, M any] struct {
	Data  T                  `json:"data"`
	Links linksPaginatedJSON `json:"links"`
	Meta  M                  `json:"meta"`
}

type linksPaginatedJSON struct {
	Self  string `json:"self"`
	First string `json:"first,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`
}

type metaPaginatedJSON struct {
	TotalRecords int `json:"totalRecords"`
	TotalPages   int `json:"totalPages"`
}

type metaPaginatedTransactionJSON struct {
	metaPaginatedJSON

	IsQueryParamUnsupported bool `json:"isQueryParamUnsupported"`
}

// WritePage writes the error response when page is out of range, and each link has filters.
func WritePage[T any](
	w http.ResponseWriter,
	r *http.Request,
	baseURL string,
	page Page,
	total int,
	filters url.Values,
	data T,
) {
	writePage(w, r, baseURL, page, total, filters, data, func(meta metaPaginatedJSON) metaPaginatedJSON {
		return meta
	})
}

// WriteTransactionPage is WritePage with the MetaPaginatedTransaction of the CDS.
func WriteTransactionPage[T any](
	w http.ResponseWriter,
	r *http.Request,
	baseURL string,
	page Page,
	total int,
	filters url.Values,
	data T,
) {
	writePage(w, r, baseURL, page, total, filters, data, func(meta metaPaginatedJSON) metaPaginatedTransactionJSON {
		return metaPaginatedTransactionJSON{metaPaginatedJSON: meta}
	})
}

func writePage[T, M any](
	w http.ResponseWriter,
	r *http.Request,
	baseURL string,
	page Page,
	total int,
	filters url.Values,
	data T,
	meta func(metaPaginatedJSON) M,
) {
	totalPages := (total + page.Size - 1) / page.Size
	// Page 1 of an empty list is in range.
	if page.Number > max(totalPages, 1) {
		WriteError(w, http.StatusUnprocessableEntity, InvalidPage, strconv.Itoa(totalPages))
		return
	}

	link := func(number int) string {
		q := maps.Clone(filters)
		if q == nil {
			q = url.Values{}
		}

		q.Set("page", strconv.Itoa(number))
		q.Set("page-size", strconv.Itoa(page.Size))

		return baseURL + r.URL.EscapedPath() + "?" + q.Encode()
	}

	links := linksPaginatedJSON{Self: link(page.Number)}
	if totalPages > 0 {
		links.First, links.Last = link(1), link(totalPages)
	}

	if page.Number > 1 {
		links.Prev = link(page.Number - 1)
	}

	if page.Number < totalPages {
		links.Next = link(page.Number + 1)
	}

	jsonhttp.Write(w, http.StatusOK, paginatedJSON[T, M]{
		Data:  data,
		Links: links,
		Meta:  meta(metaPaginatedJSON{TotalRecords: total, TotalPages: totalPages}),
	})
}
