package domain_test

import (
	"crypto/x509/pkix"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// The names of sp-active and its brand in statusFixture.
func clientSubject() pkix.Name {
	return pkix.Name{
		CommonName:         "Recipient App Web",
		Organization:       []string{"Recipient App"},
		OrganizationalUnit: []string{"Consumer Data Right"},
		Country:            []string{"AU"},
	}
}

// Verifies: REG-0610
func TestIssueClientCertificateRegistersTheIssuedCertificate(t *testing.T) {
	events, state := execute(
		t,
		statusFixture(domain.DataRecipientActive),
		domain.IssueClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	)

	assert.Equal(t, []domain.DataRecipientEvent{
		domain.ClientCertificateIssued{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	}, events)
	assert.Equal(t, []string{thumbprintA}, clientCertificates(t, state.DataRecipient, "sp-active"))
}

// Verifies: REG-0609
func TestIssueClientCertificateIsPermittedInEachStatusExceptRemoved(t *testing.T) {
	for _, id := range []string{"sp-active", "sp-inactive", "sp-under-inactive-brand"} {
		events, _ := execute(
			t,
			statusFixture(domain.DataRecipientSuspended),
			domain.IssueClientCertificate{SoftwareProductID: id, Thumbprint: thumbprintA},
		)
		assert.Len(t, events, 1, id)
	}
}

func TestIssueClientCertificateIsRefusedForARemovedOrUnknownProductOrABadThumbprint(t *testing.T) {
	state := statusFixture(domain.DataRecipientActive)

	_, err := domain.HandleDataRecipient(
		state,
		domain.IssueClientCertificate{SoftwareProductID: "sp-removed", Thumbprint: thumbprintA},
	)
	require.ErrorAs(t, err, new(domain.StatusRuleError))

	_, err = domain.HandleDataRecipient(
		state,
		domain.IssueClientCertificate{SoftwareProductID: "sp-unknown", Thumbprint: thumbprintA},
	)
	require.ErrorIs(t, err, domain.ErrNotFound)

	_, err = domain.HandleDataRecipient(
		state,
		domain.IssueClientCertificate{SoftwareProductID: "sp-active", Thumbprint: "not-a-thumbprint"},
	)
	require.ErrorIs(t, err, domain.ErrNotValid)
}

func TestAClientCSRSubjectWithTheNamesOfTheRecordBreaksNoRule(t *testing.T) {
	rules, err := domain.ClientCertificateSubjectRules(
		statusFixture(domain.DataRecipientActive),
		"sp-active",
		clientSubject(),
	)

	require.NoError(t, err)
	assert.Empty(t, rules)
}

// Verifies: REG-0606, REG-0607
func TestAClientCSRSubjectMustHaveTheNamesOfTheRecord(t *testing.T) {
	tests := map[string]struct {
		change func(*pkix.Name)
		want   []string
	}{
		"other CN": {
			func(n *pkix.Name) { n.CommonName = "Recipient App Mobile" },
			[]string{`subject CN must be the software product name "Recipient App Web"`},
		},
		"CN in other case": {
			func(n *pkix.Name) { n.CommonName = "recipient app web" },
			[]string{`subject CN must be the software product name "Recipient App Web"`},
		},
		"other O": {
			func(n *pkix.Name) { n.Organization = []string{"Recipient Pty Ltd"} },
			[]string{`subject O must be the brand name "Recipient App"`},
		},
		"two O values": {
			func(n *pkix.Name) { n.Organization = []string{"Recipient App", "Recipient App"} },
			[]string{`subject O must be the brand name "Recipient App"`},
		},
		"both": {
			func(n *pkix.Name) { n.CommonName, n.Organization = "", nil },
			[]string{
				`subject CN must be the software product name "Recipient App Web"`,
				`subject O must be the brand name "Recipient App"`,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subject := clientSubject()
			tt.change(&subject)

			rules, err := domain.ClientCertificateSubjectRules(
				statusFixture(domain.DataRecipientActive),
				"sp-active",
				subject,
			)

			require.NoError(t, err)
			assert.Equal(t, tt.want, rules)
		})
	}
}

// Verifies: REG-0608
func TestANameOfMoreThan64CharactersCannotBeInAClientCertificate(t *testing.T) {
	state := statusFixture(domain.DataRecipientActive)
	productName, brandName := strings.Repeat("é", 65), strings.Repeat("b", 65)
	state.DataRecipient.Brands[0].BrandName = brandName
	state.DataRecipient.Brands[0].SoftwareProducts[0].SoftwareProductName = productName

	subject := clientSubject()
	subject.CommonName, subject.Organization = productName, []string{brandName}
	rules, err := domain.ClientCertificateSubjectRules(state, "sp-active", subject)

	require.NoError(t, err)
	assert.Equal(t, []string{
		"the software product name has more than 64 characters, and a certificate CN can have 64 or fewer",
		"the brand name has more than 64 characters, and a certificate O can have 64 or fewer",
	}, rules)

	state.DataRecipient.Brands[0].SoftwareProducts[0].SoftwareProductName = strings.Repeat("é", 64)
	subject.CommonName = strings.Repeat("é", 64)
	rules, err = domain.ClientCertificateSubjectRules(state, "sp-active", subject)

	require.NoError(t, err)
	assert.Len(t, rules, 1, "a name of 64 characters can be in a certificate")
}

func TestTheSubjectRulesRefuseARemovedOrUnknownProductFirst(t *testing.T) {
	state := statusFixture(domain.DataRecipientActive)

	_, err := domain.ClientCertificateSubjectRules(state, "sp-removed", pkix.Name{})
	require.ErrorAs(t, err, new(domain.StatusRuleError))

	_, err = domain.ClientCertificateSubjectRules(state, "sp-unknown", pkix.Name{})
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestACSRRulesErrorIsNotValidAndNamesEachRule(t *testing.T) {
	err := domain.CSRRulesError{Rules: []string{"subject OU must be \"Consumer Data Right\"", "subject C is required"}}

	require.ErrorIs(t, err, domain.ErrNotValid)
	assert.EqualError(
		t,
		err,
		`The CSR breaks these rules: subject OU must be "Consumer Data Right"; subject C is required.`,
	)
}
