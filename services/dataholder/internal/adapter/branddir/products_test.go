package branddir_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/branddir"
)

func TestReadProductsGivesEachProductAsJSON(t *testing.T) {
	products, err := branddir.ReadProducts(
		strings.NewReader(`{"products": [{"productId": "a"}, {"productId": "b", "x": 1}]}`),
	)

	require.NoError(t, err)
	require.Len(t, products, 2)
	assert.JSONEq(t, `{"productId": "a"}`, string(products[0]))
	assert.JSONEq(t, `{"productId": "b", "x": 1}`, string(products[1]))
}

func TestReadProductsRefusesAFileThatIsNotAProductSet(t *testing.T) {
	tests := map[string]string{
		"unknown top-level field": `{"product": []}`,
		"data after the products": `{"products": []} {}`,
		"not JSON":                `products:`,
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := branddir.ReadProducts(strings.NewReader(content))

			assert.Error(t, err)
		})
	}
}
