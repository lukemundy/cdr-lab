package dataholder_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/dataholder"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/dataholder/dataholdertest"
)

const file = `{"customers": [{"lastName": "Lovelace", "emailAddress": "ada@mail.test"},
	{"lastName": "Hopper", "emailAddress": "grace@mail.test"}]}`

// Verifies: CLI-0180, CLI-0183, CLI-0186
func TestLoadCustomersSendsTheFileAndCountsTheResults(t *testing.T) {
	fake := dataholdertest.New(t)
	client := dataholder.New(fake.URL, dataholdertest.Token, fake.Client())

	first, err := client.LoadCustomers(t.Context(), []byte(file))
	require.NoError(t, err)

	second, err := client.LoadCustomers(t.Context(), []byte(file))
	require.NoError(t, err)

	assert.Equal(t, dataholder.CustomerLoad{Created: 2, Existed: 0}, first)
	assert.Equal(t, dataholder.CustomerLoad{Created: 0, Existed: 2}, second)
	assert.Equal(t, [][]byte{[]byte(file), []byte(file)}, fake.Loads())
}

// Verifies: CLI-0184
func TestARefusedCustomerLoadGivesTheStatusAndTheDetail(t *testing.T) {
	fake := dataholdertest.New(t)
	fake.Refuse(http.StatusConflict, "customers[1].accounts[0]: An account has this bsb and accountNumber.")

	_, err := dataholder.New(fake.URL, dataholdertest.Token, fake.Client()).LoadCustomers(t.Context(), []byte(file))

	require.EqualError(t, err, "load the customers into the data holder: the data holder gave 409 Conflict: "+
		"customers[1].accounts[0]: An account has this bsb and accountNumber.")
}
