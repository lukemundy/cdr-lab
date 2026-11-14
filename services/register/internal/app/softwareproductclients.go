package app

import (
	"context"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type SoftwareProductClient struct {
	SoftwareProductID   string
	Status              domain.Status
	DataRecipientStatus domain.DataRecipientStatus
	JWKSURI             string
	ClientCertificates  []string
}

type SoftwareProductClientReader interface {
	// SoftwareProductClient gives an error that wraps domain.ErrNotFound when the software product does not exist.
	SoftwareProductClient(ctx context.Context, softwareProductID string) (SoftwareProductClient, error)
}
