package consolehttp_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
)

func TestAPageHasDefaults(t *testing.T) {
	page, pageSize, err := consolehttp.ParsePage(url.Values{})

	require.NoError(t, err)
	assert.Equal(t, 1, page)
	assert.Equal(t, 25, pageSize)
}

func TestAPageIsInTheQuery(t *testing.T) {
	page, pageSize, err := consolehttp.ParsePage(url.Values{"page": {"3"}, "pageSize": {"100"}})

	require.NoError(t, err)
	assert.Equal(t, 3, page)
	assert.Equal(t, 100, pageSize)
}

func TestAPageThatIsNotValidIsAnError(t *testing.T) {
	tests := map[string]struct {
		query url.Values
		want  string
	}{
		"page not an integer":  {url.Values{"page": {"one"}}, "page must be an integer of 1 or more"},
		"page zero":            {url.Values{"page": {"0"}}, "page must be an integer of 1 or more"},
		"page size negative":   {url.Values{"pageSize": {"-1"}}, "pageSize must be an integer of 1 or more"},
		"page size too large":  {url.Values{"pageSize": {"101"}}, "pageSize must be 100 or less"},
		"page size not number": {url.Values{"pageSize": {"secret"}}, "pageSize must be an integer of 1 or more"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, err := consolehttp.ParsePage(tt.query)

			require.EqualError(t, err, tt.want)
		})
	}
}

type colour string

func TestAnEnumParameterIsOneOfTheAllowedValues(t *testing.T) {
	allowed := []colour{"RED", "BLUE"}

	none, err := consolehttp.ParseEnum(url.Values{}, "colour", allowed)
	require.NoError(t, err)
	assert.Nil(t, none)

	blue, err := consolehttp.ParseEnum(url.Values{"colour": {"BLUE"}}, "colour", allowed)
	require.NoError(t, err)
	assert.Equal(t, colour("BLUE"), *blue)

	_, err = consolehttp.ParseEnum(url.Values{"colour": {"GREEN"}}, "colour", allowed)
	require.EqualError(t, err, "colour must be one of [RED BLUE]")
}

func TestASortIsOneOfTheSortsOrTheFallback(t *testing.T) {
	sorts := map[string]int{"name": 1, "lastUpdated": 2}

	fallback, err := consolehttp.ParseSort(url.Values{}, sorts, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, fallback)

	sort, err := consolehttp.ParseSort(url.Values{"sort": {"lastUpdated"}}, sorts, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, sort)

	_, err = consolehttp.ParseSort(url.Values{"sort": {"secret"}}, sorts, 1)
	require.EqualError(t, err, "sort must be one of [lastUpdated name]")
}

func TestAnOrderIsAscendingOrDescending(t *testing.T) {
	tests := map[string]struct {
		order      string
		descending bool
	}{"default": {"", false}, "ascending": {"asc", false}, "descending": {"desc", true}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			descending, err := consolehttp.ParseOrder(url.Values{"order": {tt.order}})

			require.NoError(t, err)
			assert.Equal(t, tt.descending, descending)
		})
	}

	_, err := consolehttp.ParseOrder(url.Values{"order": {"up"}})
	require.EqualError(t, err, "order must be asc or desc")
}
