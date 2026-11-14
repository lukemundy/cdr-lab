package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type StatusResponse struct {
	// Statuses are by software product ID, as the Register gives them.
	Statuses    map[string]string
	ETag        string
	NotModified bool
}

type StatusSource interface {
	// SoftwareProductStatuses sends etag as If-None-Match when it is not empty.
	SoftwareProductStatuses(ctx context.Context, etag string) (StatusResponse, error)
}

type StatusStore interface {
	ETag(ctx context.Context) (string, error)
	// Save keeps the status of each software product that statuses does not have.
	Save(ctx context.Context, statuses map[string]domain.SoftwareProductStatus, etag string) error
	// SaveNotModified records the time of a response with no change. Before the first Save it does nothing.
	SaveNotModified(ctx context.Context) error
}

type Statuses struct {
	Source StatusSource
	Store  StatusStore
}

func (s Statuses) Refresh(ctx context.Context) error {
	etag, err := s.Store.ETag(ctx)
	if err != nil {
		return fmt.Errorf("read the ETag: %w", err)
	}

	response, err := s.Source.SoftwareProductStatuses(ctx, etag)
	if err != nil {
		return fmt.Errorf("get the statuses: %w", err)
	}

	if response.NotModified {
		if err := s.Store.SaveNotModified(ctx); err != nil {
			return fmt.Errorf("save the time of the response: %w", err)
		}

		return nil
	}

	known := make(map[string]domain.SoftwareProductStatus, len(response.Statuses))

	for id, text := range response.Statuses {
		if status, ok := domain.ParseSoftwareProductStatus(text); ok {
			known[id] = status
		}
	}

	if err := s.Store.Save(ctx, known, response.ETag); err != nil {
		return fmt.Errorf("save the statuses: %w", err)
	}

	return nil
}

type ProductStatus struct {
	SoftwareProductID string
	Status            domain.SoftwareProductStatus
	// ClientID and ClientName are empty when the software product has no registration.
	ClientID   string
	ClientName string
}

type StatusList struct {
	Products []ProductStatus
	// ReceivedAt is zero until the first Save, and SaveNotModified also changes it.
	ReceivedAt time.Time
}

type StatusSort string

const (
	// SortStatusesByName puts each software product with no registration last, also in the descending order.
	SortStatusesByName   StatusSort = "name"
	SortStatusesByStatus StatusSort = "status"
)

// StatusListQuery sorts by the software product ID after the sort key.
type StatusListQuery struct {
	// Registered is nil for no filter.
	Registered *bool
	Sort       StatusSort
	Descending bool
}

type StatusListReader interface {
	StatusList(ctx context.Context, query StatusListQuery) (StatusList, error)
}

type StatusReader interface {
	Status(ctx context.Context, softwareProductID string) (domain.SoftwareProductStatus, bool, error)
}

type ClientStatuses struct {
	Registrations RegistrationReader
	Statuses      StatusReader
}

func (c ClientStatuses) Status(ctx context.Context, clientID string) (domain.SoftwareProductStatus, error) {
	registration, err := c.Registrations.Registration(ctx, clientID)
	if errors.Is(err, ErrRegistrationNotFound) {
		return domain.SoftwareProductInactive, nil
	}

	if err != nil {
		return "", fmt.Errorf("read the registration: %w", err)
	}

	status, found, err := c.Statuses.Status(ctx, registration.SoftwareID)
	if err != nil {
		return "", fmt.Errorf("read the status: %w", err)
	}

	if !found {
		return domain.SoftwareProductInactive, nil
	}

	return status, nil
}

func (c ClientStatuses) Active(ctx context.Context, clientID string) (bool, error) {
	status, err := c.Status(ctx, clientID)
	if err != nil {
		return false, err
	}

	return status == domain.SoftwareProductActive, nil
}
