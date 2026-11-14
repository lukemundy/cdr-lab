package cdsrules_test

import (
	"testing"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/authserver/internal/cdsrules"
)

func TestAnAccessTokenOfAConsentHasNoLifetimeAfterTheEndOfTheSharing(t *testing.T) {
	now := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	grant := func(sharingDuration time.Duration, endsIn time.Duration) *goidc.Grant {
		return &goidc.Grant{Store: cdsrules.StoreOf(cdsrules.Consent{SharingEndsAt: now.Add(endsIn)}, sharingDuration)}
	}

	tests := map[string]struct {
		grant *goidc.Grant
		want  int
	}{
		"client credentials grant":            {&goidc.Grant{}, 300},
		"sharing ends in more than 5 minutes": {grant(time.Hour, 10*time.Minute), 300},
		"sharing ends in exactly 5 minutes":   {grant(time.Hour, 5*time.Minute), 300},
		"sharing ends in 2 minutes":           {grant(time.Hour, 2*time.Minute), 120},
		"sharing ends in 1.5 seconds":         {grant(time.Hour, 1500*time.Millisecond), 1},
		"sharing ended":                       {grant(time.Hour, -time.Minute), 0},
		"no sharing duration":                 {grant(0, 0), 300},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, cdsrules.AccessTokenLifetime(tc.grant, 300, now))
		})
	}
}
