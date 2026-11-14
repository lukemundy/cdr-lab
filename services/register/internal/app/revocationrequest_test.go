package app_test

import (
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

func pemOf(t *testing.T, raw []byte) string {
	t.Helper()

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw}))
}

func TestARevocationRequestGivesTheFirstCertificateAndTheReason(t *testing.T) {
	cert := issuedClient(t)
	chain := pemOf(t, cert.Raw) + pemOf(t, testLab().CDRIntermediate.Certificate.Raw)
	reason := "cessationOfOperation"

	got, gotReason, errs := app.ParseRevocationRequest(&chain, &reason)

	require.Empty(t, errs)
	assert.Equal(t, cert.Raw, got.Raw)
	assert.Equal(t, ca.CessationOfOperation, gotReason)
}

func TestARevocationRequestNamesEachFieldThatIsNotValid(t *testing.T) {
	text, reason := "not a certificate", "unknown"
	csr := "-----BEGIN CERTIFICATE REQUEST-----\nAAAA\n-----END CERTIFICATE REQUEST-----\n"

	tests := map[string]struct {
		certificate, reason *string
		want                []app.RevocationFieldError
	}{
		"no fields": {want: []app.RevocationFieldError{
			{Field: "certificate", Detail: "The certificate is required."},
			{Field: "reason", Detail: "The reason is required."},
		}},
		"not PEM and an unknown reason": {certificate: &text, reason: &reason, want: []app.RevocationFieldError{
			{Field: "certificate", Detail: app.NotACertificate},
			{Field: "reason", Detail: "The reason must be unspecified, keyCompromise, affiliationChanged, " +
				"superseded, cessationOfOperation or privilegeWithdrawn."},
		}},
		"a CSR": {certificate: &csr, reason: new("superseded"), want: []app.RevocationFieldError{
			{Field: "certificate", Detail: app.NotACertificate},
		}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cert, _, errs := app.ParseRevocationRequest(tt.certificate, tt.reason)

			assert.Nil(t, cert)
			assert.Equal(t, tt.want, errs)
		})
	}
}
