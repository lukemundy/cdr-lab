package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type owners struct {
	dr  domain.DataRecipient
	err error
}

func (o owners) DataRecipientOfSoftwareProduct(context.Context, string) (domain.DataRecipient, error) {
	return o.dr, o.err
}

type signer struct {
	claims []domain.SSAClaims
	err    error
}

func (s *signer) SignSSA(_ context.Context, claims domain.SSAClaims) (string, error) {
	s.claims = append(s.claims, claims)
	return "signed-ssa", s.err
}

func ssaRecipient(status domain.Status) domain.DataRecipient {
	return domain.DataRecipient{
		LegalEntityID: "dr-1", LegalEntityName: "Recipient Pty Ltd", Status: domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{
			{
				DataRecipientBrandID: "drb-1",
				BrandName:            "Recipient App",
				Status:               domain.Active,
				SoftwareProducts: []domain.SoftwareProduct{
					{
						SoftwareProductID:   "sp-1",
						SoftwareProductName: "Recipient App Web",
						Status:              status,
						Scopes:              []string{"openid"},
					},
				},
			},
		},
	}
}

func issuer(o owners, s *signer) app.SSAIssuer {
	return app.SSAIssuer{
		Owners: o, Signer: s,
		Now:   func() time.Time { return time.Unix(1_000_000, 0) },
		NewID: func() string { return "jti-1" },
	}
}

func TestIssueSignsTheClaimsOfTheSoftwareProduct(t *testing.T) {
	s := &signer{}

	ssa, err := issuer(owners{dr: ssaRecipient(domain.Active)}, s).Issue(t.Context(), "sp-1", "drb-1")

	require.NoError(t, err)
	assert.Equal(t, "signed-ssa", ssa)
	require.Len(t, s.claims, 1)
	assert.Equal(t, "sp-1", s.claims[0].SoftwareID)
	assert.Equal(t, "jti-1", s.claims[0].ID)
	assert.Equal(t, int64(1_000_000), s.claims[0].IssuedAt)
}

func TestIssueErrorsSignNothing(t *testing.T) {
	failure := errors.New("database failure")

	tests := map[string]struct {
		owners  owners
		brand   string
		wantErr error
	}{
		"another brand": {
			owners{dr: ssaRecipient(domain.Active)},
			"drb-other",
			app.ErrNotBrandOfSoftwareProduct,
		},
		"inactive software product": {owners{dr: ssaRecipient(domain.Inactive)}, "drb-1", domain.ErrNotActive},
		"unknown software product": {
			owners{err: fmt.Errorf("software product %w", domain.ErrNotFound)},
			"drb-1",
			domain.ErrNotFound,
		},
		"read failure": {owners{err: failure}, "drb-1", failure},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &signer{}

			_, err := issuer(tt.owners, s).Issue(t.Context(), "sp-1", tt.brand)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Empty(t, s.claims)
		})
	}
}

func TestIssueGivesTheSignerError(t *testing.T) {
	failure := errors.New("no signing key")

	_, err := issuer(owners{dr: ssaRecipient(domain.Active)}, &signer{err: failure}).Issue(t.Context(), "sp-1", "drb-1")

	assert.ErrorIs(t, err, failure)
}
