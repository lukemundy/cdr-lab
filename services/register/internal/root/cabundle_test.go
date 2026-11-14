package root

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func TestTheOperatorGetsTheCABundleThatTheMTLSListenersTrust(t *testing.T) {
	handlers := lab(t)

	rec := serveOperatorHost(t, handlers, operatorHost, http.MethodGet, "/v1/cdr-ca-bundle", authorisedOperator(), nil)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/pem-certificate-chain", rec.Header().Get("Content-Type"))

	got, err := ca.ReadCertificates(writeTemp(t, rec.Body.Bytes()))
	require.NoError(t, err)

	want, err := ca.ReadCertificates(os.Getenv("CDR_CA_BUNDLE_FILE"))
	require.NoError(t, err)
	require.Len(t, got, len(want))

	for i := range want {
		assert.Equal(t, want[i].Raw, got[i].Raw)
	}
}
