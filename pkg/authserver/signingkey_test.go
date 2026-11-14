package authserver_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

func TestANewSigningKeyIsAnRSAKeyOf2048BitsWithItsOwnID(t *testing.T) {
	first, err := authserver.NewSigningKey()
	require.NoError(t, err)
	second, err := authserver.NewSigningKey()
	require.NoError(t, err)

	assert.Equal(t, 2048, first.Key.N.BitLen())
	assert.NotEmpty(t, first.ID)
	assert.NotEqual(t, first.ID, second.ID)
	assert.False(t, first.Key.Equal(second.Key))
}

func TestASigningKeyIsTheSameAfterMarshalAndParse(t *testing.T) {
	key, err := authserver.NewSigningKey()
	require.NoError(t, err)

	der, err := key.MarshalPKCS8()
	require.NoError(t, err)
	parsed, err := authserver.ParseSigningKey(key.ID, der)

	require.NoError(t, err)
	assert.Equal(t, key.ID, parsed.ID)
	assert.True(t, key.Key.Equal(parsed.Key))
}

func TestParseSigningKeyRefusesDataThatIsNotAnRSAKey(t *testing.T) {
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ecDER, err := x509.MarshalPKCS8PrivateKey(ec)
	require.NoError(t, err)

	for name, der := range map[string][]byte{"not PKCS #8": []byte("key"), "an EC key": ecDER} {
		t.Run(name, func(t *testing.T) {
			_, err := authserver.ParseSigningKey("key-1", der)

			assert.ErrorContains(t, err, "signing key key-1")
		})
	}
}
