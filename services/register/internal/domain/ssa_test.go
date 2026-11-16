package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var ssaIssuedAt = time.Date(2026, 10, 3, 1, 2, 3, 0, time.UTC)

// Verifies: REG-0407, REG-0408, REG-0409, REG-0411
func TestTheSSAHasTheClaimsOfTheSoftwareProductItsBrandAndItsLegalEntity(t *testing.T) {
	dr := validDataRecipient()
	sp := &dr.Brands[0].SoftwareProducts[0]
	sp.SectorIdentifierURI = "https://recipient.example/sector"
	sp.TOSURI = "https://recipient.example/terms"
	sp.PolicyURI = "https://recipient.example/privacy"

	claims, err := dr.SSA("sp-1", "jti-1", ssaIssuedAt)

	require.NoError(t, err)
	data, err := json.Marshal(claims)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"iss": "cdr-register", "iat": 1790989323, "exp": 1790989923, "jti": "jti-1",
		"legal_entity_id": "dr-1", "legal_entity_name": "Recipient Pty Ltd",
		"org_id": "drb-1", "org_name": "Recipient App",
		"client_name": "Recipient App Web", "client_description": "A web app",
		"client_uri": "https://recipient.example",
		"redirect_uris": ["https://recipient.example/callback"],
		"sector_identifier_uri": "https://recipient.example/sector",
		"logo_uri": "https://recipient.example/product.png",
		"tos_uri": "https://recipient.example/terms", "policy_uri": "https://recipient.example/privacy",
		"jwks_uri": "https://recipient.example/jwks", "revocation_uri": "https://recipient.example/revoke",
		"recipient_base_uri": "https://cdr.recipient.example",
		"software_id": "sp-1", "software_roles": "data-recipient-software-product",
		"scope": "openid bank:accounts.basic:read"
	}`, string(data))
}

// Verifies: REG-0412
func TestTheSSAExpiresAfter10Minutes(t *testing.T) {
	claims, err := validDataRecipient().SSA("sp-1", "jti-1", ssaIssuedAt)

	require.NoError(t, err)
	assert.Equal(t, int64(600), claims.ExpiresAt-claims.IssuedAt)
}

// Verifies: REG-0409
func TestTheSSAOmitsTheOptionalClaimsThatTheSoftwareProductDoesNotHave(t *testing.T) {
	claims, err := validDataRecipient().SSA("sp-1", "jti-1", ssaIssuedAt)

	require.NoError(t, err)
	data, err := json.Marshal(claims)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "sector_identifier_uri")
	assert.NotContains(t, string(data), "tos_uri")
	assert.NotContains(t, string(data), "policy_uri")
}

// Verifies: REG-0402
func TestTheSSAIsOnlyForAnActiveSoftwareProductOfAnActiveDataRecipient(t *testing.T) {
	tests := map[string]struct {
		state   domain.DataRecipientState
		product string
		check   func(t *testing.T, err error)
	}{
		"inactive software product": {
			statusFixture(domain.DataRecipientActive),
			"sp-inactive",
			func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrNotActive)
			},
		},
		"removed software product": {
			statusFixture(domain.DataRecipientActive),
			"sp-removed",
			func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrNotActive)
			},
		},
		"suspended data recipient": {
			statusFixture(domain.DataRecipientSuspended),
			"sp-active",
			func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrNotActive)
			},
		},
		"unknown software product": {
			statusFixture(domain.DataRecipientActive),
			"sp-unknown",
			func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrNotFound)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := tt.state.DataRecipient.SSA(tt.product, "jti-1", ssaIssuedAt)

			require.Error(t, err)
			tt.check(t, err)
		})
	}
}

func TestTheSSAOfAnActiveSoftwareProductInAFixtureWithOtherStatuses(t *testing.T) {
	claims, err := statusFixture(domain.DataRecipientActive).DataRecipient.SSA("sp-active", "jti-1", ssaIssuedAt)

	require.NoError(t, err)
	assert.Equal(t, "sp-active", claims.SoftwareID)
	assert.Equal(t, "drb-1", claims.OrgID)
}
