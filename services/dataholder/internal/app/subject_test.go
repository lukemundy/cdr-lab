package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// Verifies: DH-0510, DH-0511, DH-0512
func TestTheSectorIdentifierOfARegistration(t *testing.T) {
	tests := map[string]struct {
		registration domain.Registration
		want         string
	}{
		"host of the sector identifier URI": {
			registration: domain.Registration{
				SectorIdentifierURI: "https://sector.example:8443/redirects.json",
				RedirectURIs:        []string{"https://app.example/callback"},
				SoftwareID:          "software-1",
			},
			want: "sector.example",
		},
		"host of the redirect URIs": {
			registration: domain.Registration{
				RedirectURIs: []string{"https://app.example/callback", "https://app.example:8443/other"},
				SoftwareID:   "software-1",
			},
			want: "app.example",
		},
		"software ID for redirect URIs with two hosts": {
			registration: domain.Registration{
				RedirectURIs: []string{"https://app.example/callback", "https://other.example/callback"},
				SoftwareID:   "software-1",
			},
			want: "software-1",
		},
		"software ID for a redirect URI with no host": {
			registration: domain.Registration{RedirectURIs: []string{"not a url"}, SoftwareID: "software-1"},
			want:         "software-1",
		},
		"software ID for no redirect URI": {
			registration: domain.Registration{SoftwareID: "software-1"},
			want:         "software-1",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, app.SectorIdentifier(tt.registration))
		})
	}
}
