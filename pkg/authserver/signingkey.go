package authserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/google/uuid"
)

// The CDS requires RSA keys of 2048 bits or longer (FAPI 1.0 Advanced, section 8.6).
const signingKeyBits = 2048

func NewSigningKey() (SigningKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, signingKeyBits)
	if err != nil {
		return SigningKey{}, err
	}

	return SigningKey{ID: uuid.NewString(), Key: key}, nil
}

func (k SigningKey) MarshalPKCS8() ([]byte, error) {
	return x509.MarshalPKCS8PrivateKey(k.Key)
}

func ParseSigningKey(id string, der []byte) (SigningKey, error) {
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return SigningKey{}, fmt.Errorf("signing key %s: %w", id, err)
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return SigningKey{}, fmt.Errorf("signing key %s is not an RSA key", id)
	}

	return SigningKey{ID: id, Key: key}, nil
}
