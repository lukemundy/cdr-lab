package app

import (
	"net/url"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// SectorIdentifier gives the value that selects the sub of a customer (OpenID Connect Core, section 8.1).
func SectorIdentifier(r domain.Registration) string {
	if host := hostOf(r.SectorIdentifierURI); host != "" {
		return host
	}

	var host string

	for _, redirectURI := range r.RedirectURIs {
		next := hostOf(redirectURI)
		if next == "" || (host != "" && next != host) {
			return r.SoftwareID
		}

		host = next
	}

	if host == "" {
		return r.SoftwareID
	}

	return host
}

func hostOf(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsed.Hostname()
}
