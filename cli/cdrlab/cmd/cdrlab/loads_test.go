package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/dataholder/dataholdertest"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register/registertest"
)

const customerFile = `{"customers": [
	{"firstName": "Ada", "lastName": "Lovelace", "emailAddress": "ada@mail.test",
		"accounts": [{"productId": "saver", "accountNumber": "11112222"}]},
	{"firstName": "Grace", "lastName": "Hopper", "emailAddress": "grace@mail.test"}]}`

// customerValues are the values of customerFile, and the customer numbers of the fake.
var customerValues = []string{
	"Ada", "Lovelace", "ada@mail.test", "saver", "11112222", "Grace", "Hopper", "grace@mail.test", "00000001",
	"00000002",
}

func writeFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "load.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func (h *harness) participantsAdd(file string) []string {
	return []string{"participants", "add", "-register-url", h.register.URL, "-file", file}
}

func (h *harness) customersAdd(file string) []string {
	return []string{"customers", "add", "-dataholder-url", h.holder.URL, "-file", file}
}

// Verifies: CLI-0177, CLI-0179, CLI-0181, CLI-0183, CLI-0185, CLI-0189
func TestParticipantsAddSendsTheFileAndGivesEachIdentifier(t *testing.T) {
	h := newHarness(t)
	h.register.SetParticipantLoad([]string{"dr-1", "brand-1"}, []string{"dh-1"})

	file := `{"dataRecipients": [{"legalEntityId": "dr-1"}]}`

	code, stdout, stderr := h.run(h.participantsAdd(writeFile(t, file))...)

	require.Equal(t, 0, code, stderr)
	assert.Equal(t, "added dr-1\nadded brand-1\nexisting dh-1\n", stdout)
	assert.Equal(t, [][]byte{[]byte(file)}, h.register.ParticipantLoads())
	assert.Zero(t, h.kubeCalls)
}

// Verifies: CLI-0062, CLI-0178, CLI-0180, CLI-0181, CLI-0183, CLI-0186, CLI-0189
func TestCustomersAddGivesTheCountsAndASecondRunChangesNothing(t *testing.T) {
	h := newHarness(t)
	path := writeFile(t, customerFile)

	code, first, stderr := h.run(h.customersAdd(path)...)
	require.Equal(t, 0, code, stderr)

	code, second, stderr := h.run(h.customersAdd(path)...)
	require.Equal(t, 0, code, stderr)

	assert.Equal(t, "2 customers created, 0 existed\n", first)
	assert.Equal(t, "0 customers created, 2 existed\n", second)
	assert.Equal(t, [][]byte{[]byte(customerFile), []byte(customerFile)}, h.holder.Loads())
	assert.Zero(t, h.kubeCalls)
}

func TestCustomersAddGivesOneCustomerInTheSingular(t *testing.T) {
	h := newHarness(t)

	code, stdout, stderr := h.run(h.customersAdd(writeFile(t, `{"customers": [{"emailAddress": "a@mail.test"}]}`))...)

	require.Equal(t, 0, code, stderr)
	assert.Equal(t, "1 customer created, 0 existed\n", stdout)
}

// Verifies: CLI-0186
func TestCustomersAddGivesNoValueOfACustomer(t *testing.T) {
	h := newHarness(t)
	path := writeFile(t, customerFile)

	_, out, errOut := h.run(h.customersAdd(path)...)

	// The data holder names a customer in a refusal by its position only.
	h.holder.Refuse(http.StatusBadRequest,
		"customers[0].accounts[0]: productId: The data holder has no product with this productId.")
	code, refusedOut, refusedErr := h.run(h.customersAdd(path)...)

	require.Equal(t, 1, code)
	assert.Contains(t, refusedErr, "customers[0].accounts[0]")

	for _, text := range []string{out, errOut, refusedOut, refusedErr} {
		for _, value := range customerValues {
			assert.NotContains(t, text, value)
		}
	}
}

// Verifies: CLI-0003, CLI-0184
func TestALoadThatTheServiceRefusesFailsWithTheStatusAndTheDetail(t *testing.T) {
	h := newHarness(t)
	h.register.Refuse(http.StatusConflict, "dr-1: the identifier is in use by a different record.")
	h.holder.Refuse(http.StatusRequestEntityTooLarge, "The request body is too large.")

	code, stdout, stderr := h.run(h.participantsAdd(writeFile(t, "{}"))...)
	assert.Equal(t, 1, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "409 Conflict: dr-1: the identifier is in use by a different record.")

	code, stdout, stderr = h.run(h.customersAdd(writeFile(t, "{}"))...)
	assert.Equal(t, 1, code)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "413 Request Entity Too Large: The request body is too large.")
}

// Verifies: CLI-0182
func TestALoadCommandNeedsEachInput(t *testing.T) {
	h := newHarness(t)
	file := writeFile(t, "{}")

	for name, tc := range map[string]struct {
		args  []string
		unset string
		want  string
	}{
		"no Register URL":     {args: []string{"participants", "add", "-file", file}, want: "-register-url"},
		"no participant file": {args: []string{"participants", "add", "-register-url", "http://r"}, want: "-file"},
		"no Register token": {
			args: h.participantsAdd(file), unset: "REGISTER_OPERATOR_TOKEN", want: "REGISTER_OPERATOR_TOKEN",
		},
		"no data holder URL": {args: []string{"customers", "add", "-file", file}, want: "-dataholder-url"},
		"no customer file":   {args: []string{"customers", "add", "-dataholder-url", "http://d"}, want: "-file"},
		"no data holder token": {
			args: h.customersAdd(file), unset: "DATAHOLDER_OPERATOR_TOKEN", want: "DATAHOLDER_OPERATOR_TOKEN",
		},
		"an argument that is not a flag": {args: append(h.customersAdd(file), "extra"), want: "extra"},
	} {
		t.Run(name, func(t *testing.T) {
			token := h.env[tc.unset]
			delete(h.env, tc.unset)
			t.Cleanup(func() {
				if tc.unset != "" {
					h.env[tc.unset] = token
				}
			})

			code, _, stderr := h.run(tc.args...)

			assert.Equal(t, 2, code)
			assert.Contains(t, stderr, tc.want)
		})
	}

	assert.Empty(t, h.register.ParticipantLoads())
	assert.Empty(t, h.holder.Loads())
}

func TestALoadOfAnAbsentFileSendsNoRequest(t *testing.T) {
	h := newHarness(t)
	absent := filepath.Join(t.TempDir(), "absent.json")

	participantsCode, _, participantsErr := h.run(h.participantsAdd(absent)...)
	customersCode, _, customersErr := h.run(h.customersAdd(absent)...)

	assert.Equal(t, 1, participantsCode)
	assert.Equal(t, 1, customersCode)
	assert.Contains(t, participantsErr, "absent.json")
	assert.Contains(t, customersErr, "absent.json")
	assert.Empty(t, h.register.ParticipantLoads())
	assert.Empty(t, h.holder.Loads())
}

// Verifies: CLI-0060
func TestALoadCommandGivesNoToken(t *testing.T) {
	h := newHarness(t)
	path := writeFile(t, customerFile)

	var texts []string

	collect := func(args []string) {
		_, out, errOut := h.run(args...)
		texts = append(texts, out, errOut)
	}

	collect(h.participantsAdd(path))
	collect(h.customersAdd(path))

	h.env["REGISTER_OPERATOR_TOKEN"], h.env["DATAHOLDER_OPERATOR_TOKEN"] = "not-the-token", "not-the-token"
	collect(h.participantsAdd(path))
	collect(h.customersAdd(path))
	collect([]string{"participants", "add", "-register-url", "http://127.0.0.1:1", "-file", path})
	collect([]string{"customers", "add", "-dataholder-url", "http://127.0.0.1:1", "-file", path})

	for _, text := range texts {
		assert.NotContains(t, text, registertest.Token)
		assert.NotContains(t, text, dataholdertest.Token)
		assert.NotContains(t, text, "not-the-token")
	}
}

// Verifies: CLI-0190, CLI-0191
func TestAParticipantAddAndACertificateIssueOnboardAParticipant(t *testing.T) {
	for name, tc := range map[string]struct {
		cert  []string
		issue string
	}{
		"a software product": {
			cert:  []string{"-kind", "client", "-legal-entity-id", "dr-1", "-software-product-id", "sp-1"},
			issue: "POST /v1/data-recipients/dr-1/software-products/sp-1/certificates/issue",
		},
		"a data holder brand": {
			cert:  []string{"-kind", "server", "-legal-entity-id", "dh-1", "-brand-id", "brand-1", "-dns", "bank"},
			issue: "POST /v1/data-holders/dh-1/brands/brand-1/certificates/issue",
		},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			dir := t.TempDir()

			code, _, stderr := h.run(h.participantsAdd(writeFile(t, "{}"))...)
			require.Equal(t, 0, code, stderr)

			code, _, stderr = h.run(
				append(
					append([]string{"cert", "issue", "-register-url", h.register.URL}, tc.cert...),
					"-cn",
					"cn",
					"-o",
					"o",
					"-cert",
					filepath.Join(dir, "c.crt"),
					"-key",
					filepath.Join(dir, "c.key"),
				)...)
			require.Equal(t, 0, code, stderr)

			assert.Len(t, h.register.ParticipantLoads(), 1)
			assert.Equal(t, []string{tc.issue}, h.register.Issues())
		})
	}
}
