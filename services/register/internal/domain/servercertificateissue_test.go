package domain_test

import (
	"crypto/x509/pkix"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// serverCertificateFixture has dhb-1 (ACTIVE), dhb-2 (interim ID only), dhb-inactive and dhb-removed.
func serverCertificateFixture() domain.DataHolderState {
	dh := validDataHolder()
	dh.Brands[0].InterimID = "interim-1"

	inactive, removed := dh.Brands[0], dh.Brands[0]
	inactive.DataHolderBrandID, inactive.InterimID, inactive.Status = "dhb-inactive", "", domain.Inactive
	removed.DataHolderBrandID, removed.InterimID, removed.Status = "dhb-removed", "", domain.Removed
	dh.Brands = append(dh.Brands, secondDataHolderBrand(), inactive, removed)

	return dataHolderState(dh)
}

func serverSubject(o string) pkix.Name {
	return pkix.Name{
		CommonName:         "mtls.holder.example",
		Organization:       []string{o},
		OrganizationalUnit: []string{"Consumer Data Right"},
		Country:            []string{"AU"},
	}
}

// Verifies: REG-0630
func TestIssueServerCertificateRecordsTheIssueForTheBrandID(t *testing.T) {
	for _, id := range []string{"dhb-1", "interim-1"} {
		t.Run(id, func(t *testing.T) {
			state := serverCertificateFixture()

			after, events := handleAndApply(
				t,
				state,
				domain.IssueServerCertificate{BrandID: id, Thumbprint: thumbprintA},
			)

			assert.Equal(t, []domain.DataHolderEvent{
				domain.ServerCertificateIssued{BrandID: "dhb-1", Thumbprint: thumbprintA},
			}, events)
			assert.Equal(t, state.DataHolder, after.DataHolder, "the Register does not keep a server certificate")
		})
	}
}

// Verifies: REG-0627
func TestIssueServerCertificateIsPermittedForAnInactiveBrand(t *testing.T) {
	_, events := handleAndApply(
		t,
		serverCertificateFixture(),
		domain.IssueServerCertificate{BrandID: "dhb-inactive", Thumbprint: thumbprintA},
	)

	assert.Len(t, events, 1)
}

func TestIssueServerCertificateIsRefusedForAnUnknownBrandOrABadThumbprint(t *testing.T) {
	tests := map[string]struct {
		state   domain.DataHolderState
		command domain.IssueServerCertificate
		want    error
	}{
		"unknown data holder": {
			state:   domain.DataHolderState{},
			command: domain.IssueServerCertificate{BrandID: "dhb-1", Thumbprint: thumbprintA},
			want:    domain.ErrNotFound,
		},
		"unknown brand": {
			state:   serverCertificateFixture(),
			command: domain.IssueServerCertificate{BrandID: "dhb-9", Thumbprint: thumbprintA},
			want:    domain.ErrNotFound,
		},
		"bad thumbprint": {
			state:   serverCertificateFixture(),
			command: domain.IssueServerCertificate{BrandID: "dhb-1", Thumbprint: "not-a-thumbprint"},
			want:    domain.ErrNotValid,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataHolder(tt.state, tt.command)

			assert.ErrorIs(t, err, tt.want)
		})
	}
}

// Verifies: REG-0626, REG-0628
func TestIssueServerCertificateIsRefusedForABrandThatCannotUseIt(t *testing.T) {
	tests := map[string]struct {
		brandID    string
		wantDetail string
	}{
		"removed brand": {
			brandID:    "dhb-removed",
			wantDetail: `data holder brand "dhb-removed" cannot get a server certificate: REMOVED is final`,
		},
		"brand with no dataHolderBrandId": {
			brandID: "dhb-2",
			wantDetail: `data holder brand "dhb-2" cannot get a server certificate: it has no dataHolderBrandId, ` +
				"so it shares only product data",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataHolder(
				serverCertificateFixture(),
				domain.IssueServerCertificate{BrandID: tt.brandID, Thumbprint: thumbprintA},
			)

			require.ErrorAs(t, err, new(domain.StatusRuleError))
			assert.EqualError(t, err, tt.wantDetail)
		})
	}
}

func TestOnlyABrandThatCanGetAServerCertificatePermitsTheIssue(t *testing.T) {
	issue := []domain.ServerCertificateAction{domain.IssueServerCertificateAction}
	want := map[string][]domain.ServerCertificateAction{
		"dhb-1":        issue,
		"dhb-2":        {},
		"dhb-inactive": issue,
		"dhb-removed":  {},
	}

	brands := serverCertificateFixture().DataHolder.Brands
	require.Len(t, brands, len(want))

	for _, b := range brands {
		t.Run(b.ID(), func(t *testing.T) {
			assert.Equal(t, want[b.ID()], domain.PermittedServerCertificateActions(b))
		})
	}
}

func TestAServerCSRSubjectWithTheBrandNameOrTheLegalEntityNameBreaksNoRule(t *testing.T) {
	for _, o := range []string{"Holder Bank", "Holder Pty Ltd"} {
		rules, err := domain.ServerCertificateSubjectRules(serverCertificateFixture(), "interim-1", serverSubject(o))

		require.NoError(t, err)
		assert.Empty(t, rules, o)
	}
}

// Verifies: REG-0623
func TestAServerCSRSubjectMustHaveTheBrandNameOrTheLegalEntityName(t *testing.T) {
	want := []string{`subject O must be the brand name "Holder Bank" or the legal entity name "Holder Pty Ltd"`}

	for name, organization := range map[string][]string{
		"other O":         {"Holder Group"},
		"O in other case": {"holder bank"},
		"no O":            nil,
		"two O values":    {"Holder Bank", "Holder Bank"},
	} {
		t.Run(name, func(t *testing.T) {
			subject := serverSubject("")
			subject.Organization = organization

			rules, err := domain.ServerCertificateSubjectRules(serverCertificateFixture(), "dhb-1", subject)

			require.NoError(t, err)
			assert.Equal(t, want, rules)
		})
	}
}

// Verifies: REG-0624
func TestABrandNameOfMoreThan64CharactersPermitsOnlyTheLegalEntityName(t *testing.T) {
	state := serverCertificateFixture()
	brandName := strings.Repeat("é", 65)
	state.DataHolder.Brands[0].BrandName = brandName

	rules, err := domain.ServerCertificateSubjectRules(state, "dhb-1", serverSubject(brandName))

	require.NoError(t, err)
	assert.Equal(t, []string{
		`subject O must be the legal entity name "Holder Pty Ltd": the brand name has more than 64 characters, ` +
			"and a certificate O can have 64 or fewer",
	}, rules)

	rules, err = domain.ServerCertificateSubjectRules(state, "dhb-1", serverSubject("Holder Pty Ltd"))

	require.NoError(t, err)
	assert.Empty(t, rules)
}

func TestTheServerSubjectRulesRefuseABrandThatCannotGetACertificateFirst(t *testing.T) {
	state := serverCertificateFixture()

	_, err := domain.ServerCertificateSubjectRules(domain.DataHolderState{}, "dhb-1", pkix.Name{})
	require.ErrorIs(t, err, domain.ErrNotFound)

	_, err = domain.ServerCertificateSubjectRules(state, "dhb-9", pkix.Name{})
	require.ErrorIs(t, err, domain.ErrNotFound)

	for _, id := range []string{"dhb-removed", "dhb-2"} {
		_, err = domain.ServerCertificateSubjectRules(state, id, pkix.Name{})
		require.ErrorAs(t, err, new(domain.StatusRuleError), id)
	}
}
