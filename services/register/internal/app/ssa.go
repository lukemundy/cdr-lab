package app

import (
	"context"
	"errors"
	"time"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var ErrNotBrandOfSoftwareProduct = errors.New("the data recipient brand is not the brand of the software product")

type SoftwareProductOwnerReader interface {
	// DataRecipientOfSoftwareProduct gives an error that wraps domain.ErrNotFound
	// when the software product does not exist.
	DataRecipientOfSoftwareProduct(ctx context.Context, softwareProductID string) (domain.DataRecipient, error)
}

type SSASigner interface {
	SignSSA(ctx context.Context, claims domain.SSAClaims) (string, error)
}

type SSAIssuer struct {
	Owners SoftwareProductOwnerReader
	Signer SSASigner
	Now    func() time.Time
	NewID  func() string
}

// Issue gives ErrNotBrandOfSoftwareProduct when dataRecipientBrandID is not the brand of the software product.
func (i SSAIssuer) Issue(ctx context.Context, softwareProductID, dataRecipientBrandID string) (string, error) {
	dr, err := i.Owners.DataRecipientOfSoftwareProduct(ctx, softwareProductID)
	if err != nil {
		return "", err
	}

	claims, err := dr.SSA(softwareProductID, i.NewID(), i.Now())
	if err != nil {
		return "", err
	}

	if claims.OrgID != dataRecipientBrandID {
		return "", ErrNotBrandOfSoftwareProduct
	}

	return i.Signer.SignSSA(ctx, claims)
}
