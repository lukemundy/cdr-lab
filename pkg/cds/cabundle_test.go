package cds_test

import (
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

func TestLoadCABundleGivesEachCertificateOfTheFile(t *testing.T) {
	first, _ := serverCertificate(t)
	second, _ := serverCertificate(t)
	file := filepath.Join(t.TempDir(), "ca-bundle.pem")

	var data []byte
	for _, cert := range [][]byte{first.Certificate[0], second.Certificate[0]} {
		data = append(data, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})...)
	}

	require.NoError(t, os.WriteFile(file, data, 0o600))

	certs, err := cds.LoadCABundle(file)

	require.NoError(t, err)
	require.Len(t, certs, 2)
	assert.Equal(t, first.Certificate[0], certs[0].Raw)
	assert.Equal(t, second.Certificate[0], certs[1].Raw)
}

func TestLoadCABundleRefusesAFileThatIsNotACABundle(t *testing.T) {
	dir := t.TempDir()
	tests := map[string][]byte{
		"no PEM certificate":         []byte("not PEM"),
		"a certificate with a fault": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not DER")}),
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			file := filepath.Join(dir, "bundle.pem")
			require.NoError(t, os.WriteFile(file, data, 0o600))

			_, err := cds.LoadCABundle(file)

			assert.ErrorContains(t, err, "load the CDR CA bundle")
		})
	}

	t.Run("no file", func(t *testing.T) {
		_, err := cds.LoadCABundle(filepath.Join(dir, "absent.pem"))

		assert.ErrorContains(t, err, "load the CDR CA bundle")
	})
}
