package root

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

type composeService struct {
	Profiles    []string          `yaml:"profiles"`
	Ports       []string          `yaml:"ports"`
	Environment map[string]string `yaml:"environment"`
}

func sorrelService(t *testing.T) composeService {
	t.Helper()

	data, err := os.ReadFile("../../../../compose.yaml")
	require.NoError(t, err)

	var compose struct {
		Services map[string]composeService `yaml:"services"`
	}
	require.NoError(t, yaml.Unmarshal(data, &compose))

	service, ok := compose.Services["sorrel"]
	require.True(t, ok, "compose.yaml has no service sorrel")

	return service
}

// Verifies: DR-0600, DR-0601
func TestTheLabRunsTheDefaultPresetBrandWithNoProfile(t *testing.T) {
	service := sorrelService(t)

	assert.Empty(t, service.Profiles)
	assert.Equal(t, "sorrel", service.Environment["DATA_RECIPIENT_BRAND"])
	assert.Equal(t, "sorrel", service.Environment["PGDATABASE"])
}

// Verifies: DR-0602, DR-0603
func TestTheLabPublishesTheListenersOfSorrelInItsPortBlockOnTheLoopbackOnly(t *testing.T) {
	var want []string
	for _, port := range []int{internalPort, webAppPort} {
		want = append(want, fmt.Sprintf("127.0.0.1:%d:%d", 8580+port%10, port))
	}

	assert.ElementsMatch(t, want, sorrelService(t).Ports)
}

// Verifies: DR-0606, DR-0607
func TestTheLabGivesSorrelTheMailSinkAndTheBaseURIOfTheHost(t *testing.T) {
	env := sorrelService(t).Environment

	assert.Equal(t, "mail:1025", env["SMTP_ADDRESS"])
	assert.Equal(t, "no-reply@sorrel.test", env["MAIL_FROM"])
	assert.Equal(t, "https://localhost:8585", env["WEB_APP_BASE_URI"])
}
