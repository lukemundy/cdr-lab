package root

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

// Verifies: DH-1310, DH-1311
func TestEachBankInstanceOfTheLabPublishesItsListenersInItsPortBlock(t *testing.T) {
	data, err := os.ReadFile("../../../../compose.yaml")
	require.NoError(t, err)

	var compose struct {
		Services map[string]struct {
			Ports []string `yaml:"ports"`
		} `yaml:"services"`
	}
	require.NoError(t, yaml.Unmarshal(data, &compose))

	blocks := map[string]int{"wattle-bank": 8180, "redgum-bank": 8280, "harbour-bank": 8380, "kestrel-bank": 8480}
	listeners := []int{internalPort, apiPort, securePort, bankSitePort}

	for name, block := range blocks {
		t.Run(name, func(t *testing.T) {
			service, ok := compose.Services[name]
			require.True(t, ok, "compose.yaml has no service %s", name)

			var want []string
			for _, port := range listeners {
				want = append(want, fmt.Sprintf("127.0.0.1:%d:%d", block+port%10, port))
			}

			assert.ElementsMatch(t, want, service.Ports)
		})
	}
}
