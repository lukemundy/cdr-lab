package app_test

import (
	"flag"
	"testing"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventsourcingtest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var update = flag.Bool("update", false, "write the event golden file in testdata")

// Verifies: RLS-0080, RLS-0081, RLS-0082, RLS-0083, RLS-0084, RLS-0085, RLS-0086, RLS-0087, RLS-0088
func TestTheEncodedSampleOfEachEventTypeIsTheEventGoldenFile(t *testing.T) {
	golden := eventsourcingtest.NewGolden("testdata/events.golden.json")
	eventsourcingtest.Add(golden, app.ArrangementType, arrangementSamples()...)
	eventsourcingtest.Add(golden, app.ClientRegistrationType, clientRegistrationSamples()...)
	golden.Check(t, *update)
}

var sampleTime = time.Date(2026, 11, 23, 4, 5, 6, 0, time.UTC)

func arrangementSamples() []domain.ArrangementEvent {
	attempt := domain.NoticeAttempt{Attempt: 1, At: sampleTime, Result: "HTTP 204"}

	return []domain.ArrangementEvent{
		domain.ArrangementCreated{Arrangement: domain.Arrangement{
			ArrangementID:          "arr-1",
			CustomerNumber:         "c-000001",
			ClientID:               "client-1",
			SoftwareID:             "sp-1",
			ClientName:             "Sample Budget",
			OrgName:                "Sample App",
			LegalEntityName:        "Sample Recipient Pty Ltd",
			RecipientBaseURI:       "https://api.app.example.com",
			Scopes:                 []string{"openid", "bank:accounts.basic:read"},
			AccountIDs:             []string{"acc-1"},
			ConfirmedAt:            sampleTime,
			SharingDurationSeconds: 7776000,
			OnceOffSharingSeconds:  600,
		}},
		domain.ArrangementRevoked{RevokedAt: sampleTime},
		domain.ArrangementWithdrawn{WithdrawnAt: sampleTime, Method: domain.WithdrawalMethodConsumerDashboard},
		domain.ArrangementEndedForIneligibility{EndedAt: sampleTime},
		domain.ArrangementEndedForRemoval{EndedAt: sampleTime},
		domain.ArrangementEndedForDeletion{EndedAt: sampleTime},
		domain.NoticeAttempted{NoticeAttempt: attempt},
		domain.NoticeRecorded{
			RecordedAt: sampleTime,
			Status:     domain.NoticeComplete,
			Attempts:   []domain.NoticeAttempt{attempt},
		},
	}
}

func clientRegistrationSamples() []domain.ClientRegistrationEvent {
	//nolint:gosec // G101 reads the protocol field token_endpoint_auth_method as a credential
	registration := domain.Registration{
		ClientID:                       "client-1",
		ClientIDIssuedAt:               1795406706,
		ClientName:                     "Sample Budget",
		ClientDescription:              "A sample budget product.",
		ClientURI:                      "https://app.example.com",
		LegalEntityID:                  "le-recipient",
		LegalEntityName:                "Sample Recipient Pty Ltd",
		OrgID:                          "drb-1",
		OrgName:                        "Sample App",
		RedirectURIs:                   []string{"https://app.example.com/callback"},
		SectorIdentifierURI:            "https://app.example.com/sector",
		LogoURI:                        "https://app.example.com/budget.png",
		TOSURI:                         "https://app.example.com/terms",
		PolicyURI:                      "https://app.example.com/policy",
		JWKSURI:                        "https://app.example.com/jwks",
		RevocationURI:                  "https://app.example.com/revoke",
		RecipientBaseURI:               "https://api.app.example.com",
		TokenEndpointAuthMethod:        "private_key_jwt",
		TokenEndpointAuthSigningAlg:    domain.PS256,
		GrantTypes:                     []domain.GrantType{domain.GrantAuthorizationCode, domain.GrantRefreshToken},
		ResponseTypes:                  []string{"code"},
		ApplicationType:                "web",
		IDTokenSignedResponseAlg:       domain.PS256,
		AuthorizationSignedResponseAlg: domain.PS256,
		RequestObjectSigningAlg:        domain.PS256,
		SoftwareStatement:              "sample.software.statement",
		SoftwareID:                     "sp-1",
		SoftwareRoles:                  "data-recipient-software-product",
		Scope:                          "openid bank:accounts.basic:read",
	}

	return []domain.ClientRegistrationEvent{
		domain.ClientRegistered{Registration: registration},
		domain.ClientUpdated{Registration: registration},
		domain.ClientDeleted{ClientID: "client-1"},
	}
}
