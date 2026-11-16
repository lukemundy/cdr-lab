package cds_test

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

const (
	pageBaseURL = "https://api.example"
	pagePath    = "/cds-au/v1/things"
)

func parsePage(t *testing.T, query string) (cds.Page, bool, *httptest.ResponseRecorder) {
	t.Helper()

	rec := httptest.NewRecorder()
	page, ok := cds.ParsePage(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, pagePath+"?"+query, nil))

	return page, ok, rec
}

// Verifies: CDSL-0017, CDSL-0018, CDSL-0019, CDSL-0022
func TestParsePageGivesThePageOfTheRequestOrTheCDSDefaults(t *testing.T) {
	tests := map[string]struct {
		query string
		want  cds.Page
	}{
		"no parameters":         {"", cds.Page{Number: 1, Size: 25}},
		"empty parameters":      {"page=&page-size=", cds.Page{Number: 1, Size: 25}},
		"page and page-size":    {"page=3&page-size=10", cds.Page{Number: 3, Size: 10}},
		"the maximum page-size": {"page-size=1000", cds.Page{Number: 1, Size: 1000}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			page, ok, rec := parsePage(t, tt.query)

			require.True(t, ok, rec.Body.String())
			assert.Equal(t, tt.want, page)
			assert.Empty(t, rec.Body.String())
		})
	}
}

// Verifies: CDSL-0020, CDSL-0021
func TestParsePageRefusesAPageOrAPageSizeThatIsNotValid(t *testing.T) {
	const (
		invalid         = "urn:au-cds:error:cds-all:Field/Invalid"
		invalidPageSize = "urn:au-cds:error:cds-all:Field/InvalidPageSize"
	)

	tests := map[string]struct {
		query, code, title, detail string
	}{
		"page zero":                   {"page=0", invalid, "Invalid Field", "page"},
		"page not a number":           {"page=one", invalid, "Invalid Field", "page"},
		"page negative":               {"page=-1", invalid, "Invalid Field", "page"},
		"page-size zero":              {"page-size=0", invalidPageSize, "Invalid Page Size", "page-size"},
		"page-size not a number":      {"page-size=many", invalidPageSize, "Invalid Page Size", "page-size"},
		"page-size above the maximum": {"page-size=1001", invalidPageSize, "Invalid Page Size", "page-size"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, ok, rec := parsePage(t, tt.query)

			assert.False(t, ok)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.JSONEq(
				t,
				fmt.Sprintf(`{"errors": [{"code": %q, "title": %q, "detail": %q}]}`, tt.code, tt.title, tt.detail),
				rec.Body.String(),
			)
		})
	}
}

func TestReadPageGivesThePageOrTheParameterThatIsNotValid(t *testing.T) {
	page, err := cds.ReadPage(url.Values{"page": {"3"}, "page-size": {"1000"}})
	require.NoError(t, err)
	assert.Equal(t, cds.Page{Number: 3, Size: 1000}, page)

	page, err = cds.ReadPage(url.Values{"page": {"99999999999999999999"}})
	require.NoError(t, err)
	assert.Equal(t, math.MaxInt32, page.Number, "the limit for a row offset")

	tests := map[string]struct {
		query url.Values
		want  string
	}{
		"page":      {url.Values{"page": {"0"}}, "page must be an integer of 1 or more"},
		"page-size": {url.Values{"page-size": {"1001"}}, "page-size must be an integer from 1 to 1000"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := cds.ReadPage(tt.query)

			var pageErr cds.PageError
			require.ErrorAs(t, err, &pageErr)
			assert.Equal(t, name, pageErr.Parameter)
			assert.EqualError(t, err, tt.want)
		})
	}
}

func writePage(
	t *testing.T,
	page cds.Page,
	total int,
	filters url.Values,
	data any,
) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, pagePath+"?ignored=1", nil)
	cds.WritePage(rec, req, pageBaseURL, page, total, filters, data)

	return rec
}

// Verifies: CDSL-0024, CDSL-0025, CDSL-0026, CDSL-0027, CDSL-0028, CDSL-0031
func TestWritePageGivesTheLinksAndMetaOfTheCDSPaginationRules(t *testing.T) {
	link := func(page int) string {
		return fmt.Sprintf("%s%s?page=%d&page-size=2", pageBaseURL, pagePath, page)
	}

	tests := map[string]struct {
		page, total int
		wantLinks   map[string]string
		wantPages   int
	}{
		"first of three pages": {
			page:      1,
			total:     5,
			wantLinks: map[string]string{"self": link(1), "first": link(1), "next": link(2), "last": link(3)},
			wantPages: 3,
		},
		"middle of three pages": {
			page:  2,
			total: 5,
			wantLinks: map[string]string{
				"self":  link(2),
				"first": link(1),
				"prev":  link(1),
				"next":  link(3),
				"last":  link(3),
			},
			wantPages: 3,
		},
		"last of three pages": {
			page:      3,
			total:     5,
			wantLinks: map[string]string{"self": link(3), "first": link(1), "prev": link(2), "last": link(3)},
			wantPages: 3,
		},
		"total fills the pages": {
			page:      2,
			total:     4,
			wantLinks: map[string]string{"self": link(2), "first": link(1), "prev": link(1), "last": link(2)},
			wantPages: 2,
		},
		"no records": {page: 1, total: 0, wantLinks: map[string]string{"self": link(1)}, wantPages: 0},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := writePage(t, cds.Page{Number: tt.page, Size: 2}, tt.total, nil, []string{"a"})

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var body struct {
				Data  []string          `json:"data"`
				Links map[string]string `json:"links"`
				Meta  map[string]int    `json:"meta"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, []string{"a"}, body.Data)
			assert.Equal(t, tt.wantLinks, body.Links)
			assert.Equal(t, map[string]int{"totalRecords": tt.total, "totalPages": tt.wantPages}, body.Meta)
		})
	}
}

// Verifies: CDSL-0029
func TestWritePageGivesEachLinkTheFiltersOfTheRequest(t *testing.T) {
	filters := url.Values{"updated-since": {"2026-10-01T08:00:00+08:00"}, "brand": {"Mock Bank"}}

	rec := writePage(t, cds.Page{Number: 1, Size: 1}, 2, filters, []string{})

	var body struct {
		Links map[string]string `json:"links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	const query = "brand=Mock+Bank&page=%d&page-size=1&updated-since=2026-10-01T08%%3A00%%3A00%%2B08%%3A00"
	for name, page := range map[string]int{"self": 1, "first": 1, "next": 2, "last": 2} {
		assert.Equal(t, pageBaseURL+pagePath+"?"+fmt.Sprintf(query, page), body.Links[name], name)
	}

	assert.Equal(t, url.Values{"updated-since": {"2026-10-01T08:00:00+08:00"}, "brand": {"Mock Bank"}}, filters)
}

// Verifies: CDSL-0023
func TestWritePageRefusesAPageThatIsOutOfRange(t *testing.T) {
	tests := map[string]struct {
		page, total int
		detail      string
	}{
		"after the last page":              {page: 4, total: 5, detail: "3"},
		"page 2 of a list with no records": {page: 2, total: 0, detail: "0"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := writePage(t, cds.Page{Number: tt.page, Size: 2}, tt.total, nil, []string{})

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.JSONEq(t, fmt.Sprintf(`{"errors": [{
				"code": "urn:au-cds:error:cds-all:Field/InvalidPage",
				"title": "Invalid Page",
				"detail": %q
			}]}`, tt.detail), rec.Body.String())
		})
	}
}

func TestAPageNumberThatIsTooLargeForAnIntegerIsOutOfRange(t *testing.T) {
	page, ok, _ := parsePage(t, "page=99999999999999999999999")
	require.True(t, ok)

	rec := writePage(t, page, 1, nil, []string{})

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	// The row offset of the page must not be too large for the database.
	assert.LessOrEqual(t, page.Number, math.MaxInt32)
}

// Verifies: CDSL-0032
func TestWriteTransactionPageAddsThatEachQueryParameterIsSupported(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, pagePath, nil)

	cds.WriteTransactionPage(rec, req, pageBaseURL, cds.Page{Number: 1, Size: 25}, 1, nil, []string{"a"})

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Meta map[string]any `json:"meta"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, map[string]any{"totalRecords": 1.0, "totalPages": 1.0, "isQueryParamUnsupported": false}, body.Meta)

	outOfRange := httptest.NewRecorder()
	cds.WriteTransactionPage(outOfRange, req, pageBaseURL, cds.Page{Number: 2, Size: 25}, 1, nil, []string{})
	assert.Equal(t, http.StatusUnprocessableEntity, outOfRange.Code)
}
