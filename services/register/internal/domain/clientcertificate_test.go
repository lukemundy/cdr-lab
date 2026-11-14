package domain_test

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// Base64url SHA-256 values. The last character of a value has two bits that must be zero.
const (
	thumbprintA = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	thumbprintB = "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBA"
)

func softwareProductOf(t *testing.T, dr domain.DataRecipient, id string) domain.SoftwareProduct {
	t.Helper()

	for _, b := range dr.Brands {
		for _, sp := range b.SoftwareProducts {
			if sp.SoftwareProductID == id {
				return sp
			}
		}
	}

	t.Fatalf("software product %q is not in the data recipient", id)

	return domain.SoftwareProduct{}
}

func clientCertificates(t *testing.T, dr domain.DataRecipient, id string) []string {
	t.Helper()

	return softwareProductOf(t, dr, id).ClientCertificates
}

func TestRegisterClientCertificateAddsTheThumbprintToTheSoftwareProduct(t *testing.T) {
	before := statusFixture(domain.DataRecipientActive)

	events, state := execute(
		t,
		before,
		domain.RegisterClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	)
	_, state = execute(
		t,
		state,
		domain.RegisterClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintB},
	)

	assert.Equal(t, []domain.DataRecipientEvent{
		domain.ClientCertificateRegistered{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	}, events)
	assert.Equal(t, []string{thumbprintA, thumbprintB}, clientCertificates(t, state.DataRecipient, "sp-active"))
	assert.Empty(t, clientCertificates(t, state.DataRecipient, "sp-inactive"), "other software products do not change")
	assert.Empty(
		t,
		clientCertificates(t, before.DataRecipient, "sp-active"),
		"apply does not change the state of the caller",
	)
}

func TestRegisterClientCertificateGivesNoEventWhenTheCertificateIsRegistered(t *testing.T) {
	_, state := execute(
		t,
		statusFixture(domain.DataRecipientActive),
		domain.RegisterClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	)

	events, state := execute(
		t,
		state,
		domain.RegisterClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	)

	assert.Empty(t, events)
	assert.Equal(t, []string{thumbprintA}, clientCertificates(t, state.DataRecipient, "sp-active"))
}

func TestAnInactiveSoftwareProductCanGetAClientCertificate(t *testing.T) {
	_, state := execute(
		t,
		statusFixture(domain.DataRecipientActive),
		domain.RegisterClientCertificate{SoftwareProductID: "sp-inactive", Thumbprint: thumbprintA},
	)

	assert.Equal(t, []string{thumbprintA}, clientCertificates(t, state.DataRecipient, "sp-inactive"))
}

func TestRegisterClientCertificateErrors(t *testing.T) {
	tests := map[string]struct {
		command domain.RegisterClientCertificate
		check   func(t *testing.T, err error)
	}{
		"unknown software product": {
			domain.RegisterClientCertificate{SoftwareProductID: "sp-unknown", Thumbprint: thumbprintA},
			func(t *testing.T, err error) {
				require.ErrorIs(t, err, domain.ErrNotFound)
				assert.EqualError(t, err, `software product "sp-unknown" does not exist`)
			},
		},
		"removed software product": {
			domain.RegisterClientCertificate{SoftwareProductID: "sp-removed", Thumbprint: thumbprintA},
			func(t *testing.T, err error) {
				require.ErrorAs(t, err, &domain.StatusRuleError{})
				assert.EqualError(
					t,
					err,
					`software product "sp-removed" cannot get a client certificate: REMOVED is final`,
				)
			},
		},
		"thumbprint is not base64url SHA-256": {
			domain.RegisterClientCertificate{SoftwareProductID: "sp-active", Thumbprint: "not-a-thumbprint"},
			func(t *testing.T, err error) {
				require.ErrorIs(t, err, domain.ErrNotValid)
				assert.EqualError(
					t,
					err,
					`software product "sp-active": a client certificate thumbprint is not a base64url SHA-256 value`,
				)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events, err := domain.HandleDataRecipient(statusFixture(domain.DataRecipientActive), tt.command)

			require.Error(t, err)
			assert.Empty(t, events)
			tt.check(t, err)
		})
	}
}

func TestImportDataRecipientValidatesClientCertificates(t *testing.T) {
	tests := map[string]struct {
		thumbprints []string
		wantErr     string
	}{
		"valid": {[]string{thumbprintA, thumbprintB}, ""},
		"not SHA-256": {
			[]string{"AAAA"},
			`software product "sp-1": a client certificate thumbprint is not a base64url SHA-256 value`,
		},
		"padded": {
			[]string{thumbprintA + "="},
			`software product "sp-1": a client certificate thumbprint is not a base64url SHA-256 value`,
		},
		"duplicate": {
			[]string{thumbprintA, thumbprintA},
			`software product "sp-1": a client certificate thumbprint is not unique`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dr := validDataRecipient()
			dr.Brands[0].SoftwareProducts[0].ClientCertificates = tt.thumbprints

			_, err := importDataRecipient(dr)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func registered(t *testing.T, productIDs ...string) domain.DataRecipientState {
	t.Helper()

	state := statusFixture(domain.DataRecipientActive)
	for _, id := range productIDs {
		_, state = execute(t, state, domain.RegisterClientCertificate{SoftwareProductID: id, Thumbprint: thumbprintA})
		_, state = execute(t, state, domain.RegisterClientCertificate{SoftwareProductID: id, Thumbprint: thumbprintB})
	}

	return state
}

func TestRemoveClientCertificateRemovesOnlyThatThumbprintOfTheSoftwareProduct(t *testing.T) {
	before := registered(t, "sp-active", "sp-inactive")

	events, state := execute(
		t,
		before,
		domain.RemoveClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	)

	assert.Equal(t, []domain.DataRecipientEvent{
		domain.ClientCertificateRemoved{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	}, events)
	assert.Equal(t, []string{thumbprintB}, clientCertificates(t, state.DataRecipient, "sp-active"))
	assert.Equal(
		t,
		[]string{thumbprintA, thumbprintB},
		clientCertificates(t, state.DataRecipient, "sp-inactive"),
		"other software products do not change",
	)
	assert.Equal(
		t,
		[]string{thumbprintA, thumbprintB},
		clientCertificates(t, before.DataRecipient, "sp-active"),
		"apply does not change the state of the caller",
	)
}

func TestRemoveClientCertificateGivesNoEventWhenTheThumbprintIsNotRegistered(t *testing.T) {
	_, state := execute(
		t,
		registered(t, "sp-active"),
		domain.RemoveClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	)

	events, state := execute(
		t,
		state,
		domain.RemoveClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintA},
	)

	assert.Empty(t, events)
	assert.Equal(t, []string{thumbprintB}, clientCertificates(t, state.DataRecipient, "sp-active"))
}

func TestARemovedSoftwareProductCanLoseAClientCertificate(t *testing.T) {
	state := registered(t, "sp-active")
	_, state = execute(t, state, domain.RemoveSoftwareProduct{SoftwareProductID: "sp-active"})

	events, state := execute(
		t,
		state,
		domain.RemoveClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintB},
	)

	assert.Len(t, events, 1)
	assert.Equal(t, []string{thumbprintA}, clientCertificates(t, state.DataRecipient, "sp-active"))
}

func TestRemoveClientCertificateErrors(t *testing.T) {
	tests := map[string]struct {
		command domain.RemoveClientCertificate
		check   func(t *testing.T, err error)
	}{
		"unknown software product": {
			domain.RemoveClientCertificate{SoftwareProductID: "sp-unknown", Thumbprint: thumbprintA},
			func(t *testing.T, err error) {
				require.ErrorIs(t, err, domain.ErrNotFound)
				assert.EqualError(t, err, `software product "sp-unknown" does not exist`)
			},
		},
		"thumbprint is not base64url SHA-256": {
			domain.RemoveClientCertificate{SoftwareProductID: "sp-active", Thumbprint: thumbprintA + "="},
			func(t *testing.T, err error) {
				require.ErrorIs(t, err, domain.ErrNotValid)
				assert.EqualError(
					t,
					err,
					`software product "sp-active": a client certificate thumbprint is not a base64url SHA-256 value`,
				)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events, err := domain.HandleDataRecipient(registered(t, "sp-active"), tt.command)

			require.Error(t, err)
			assert.Empty(t, events)
			tt.check(t, err)
		})
	}
}

// zeroReader is the source of randomness of a certificate in a test: the domain gets no random values, and an Ed25519
// signature needs none.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

func certificatePEM(t *testing.T) (pemData []byte, der []byte) {
	t.Helper()

	key := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Mock Budget Web"}}
	der, err := x509.CreateCertificate(zeroReader{}, tmpl, tmpl, key.Public(), key)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), der
}

func TestClientCertificateThumbprintIsTheBase64urlSHA256OfTheDER(t *testing.T) {
	pemData, der := certificatePEM(t)
	sum := sha256.Sum256(der)

	thumbprint, err := domain.ClientCertificateThumbprint(append([]byte("\n"), pemData...))

	require.NoError(t, err)
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(sum[:]), thumbprint)
}

func TestClientCertificateThumbprintRefusesTextThatIsNotOnePEMCertificate(t *testing.T) {
	pemData, _ := certificatePEM(t)
	key := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("key")})

	for name, data := range map[string][]byte{
		"empty":             nil,
		"not PEM":           []byte("not a certificate"),
		"not a certificate": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not DER")}),
		"other block type":  key,
		"two certificates":  append(append([]byte{}, pemData...), pemData...),
		"text after":        append(append([]byte{}, pemData...), []byte("more")...),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := domain.ClientCertificateThumbprint(data)

			require.ErrorIs(t, err, domain.ErrNotValid)
		})
	}
}

func TestASoftwareProductThatIsNotRemovedPermitsTheRegistrationTheIssueAndTheRemovalOfACertificate(t *testing.T) {
	dr := statusFixture(domain.DataRecipientActive).DataRecipient
	all := []domain.ClientCertificateAction{
		domain.RegisterCertificate,
		domain.IssueCertificate,
		domain.RemoveCertificate,
	}

	tests := map[string][]domain.ClientCertificateAction{
		"sp-active":   all,
		"sp-inactive": all,
		"sp-removed":  {domain.RemoveCertificate},
	}
	for id, want := range tests {
		t.Run(id, func(t *testing.T) {
			assert.Equal(t, want, domain.PermittedClientCertificateActions(softwareProductOf(t, dr, id)))
		})
	}
}
