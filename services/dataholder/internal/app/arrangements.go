package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var ErrInvalidArrangement = errors.New("the arrangement does not exist, is of a different client or ended")

type GrantRevoker interface {
	RevokeArrangement(ctx context.Context, arrangementID string) error
}

type ArrangementReader interface {
	// ArrangementsOf gives the arrangements of the customer in the sequence of their confirmation.
	ArrangementsOf(ctx context.Context, customerNumber string) ([]domain.ArrangementState, error)
	ArrangementsOfSoftwareProduct(ctx context.Context, softwareID string) ([]domain.ArrangementState, error)
}

type Arrangements struct {
	Store     eventsourcing.Store
	Grants    GrantRevoker
	Customers CustomerStore
	Reader    ArrangementReader
	Now       func() time.Time
}

func (a Arrangements) RevokeByDataRecipient(ctx context.Context, clientID, arrangementID string) error {
	now := a.Now()

	state, err := a.load(ctx, arrangementID)
	if err != nil {
		return err
	}

	if !state.Created || state.Arrangement.ClientID != clientID || state.Ended(now) {
		return ErrInvalidArrangement
	}

	_, err = a.end(ctx, state, domain.EndArrangement{EndedAt: now, Reason: domain.EndReasonRevokedByDataRecipient})
	if errors.Is(err, domain.ErrArrangementEnded) {
		return ErrInvalidArrangement
	}

	return err
}

// EndOneTimeCollection ends, as a revocation by the data recipient, a current arrangement of the client for a
// disclosure on one occasion, because the client revoked its refresh token after the collection. It changes nothing
// for each other arrangement.
func (a Arrangements) EndOneTimeCollection(ctx context.Context, clientID, arrangementID string) error {
	now := a.Now()

	state, err := a.load(ctx, arrangementID)
	if err != nil {
		return err
	}

	sharing := time.Duration(state.Arrangement.SharingDurationSeconds) * time.Second
	if !state.Created || state.Arrangement.ClientID != clientID || state.Ended(now) || !domain.IsOneOccasion(sharing) {
		return nil
	}

	_, err = a.end(ctx, state, domain.EndArrangement{EndedAt: now, Reason: domain.EndReasonRevokedByDataRecipient})
	if errors.Is(err, domain.ErrArrangementEnded) {
		return nil
	}

	return err
}

// WithdrawByConsumer gives ErrCustomerNotFound, domain.ErrArrangementNotFound for an arrangement that the customer
// does not have, and domain.ErrArrangementEnded.
func (a Arrangements) WithdrawByConsumer(
	ctx context.Context,
	customerNumber, arrangementID string,
	method domain.WithdrawalMethod,
) (domain.ArrangementState, error) {
	now := a.Now()

	if _, err := a.Customers.Customer(ctx, customerNumber); err != nil {
		return domain.ArrangementState{}, fmt.Errorf("read the customer: %w", err)
	}

	state, err := a.load(ctx, arrangementID)
	if err != nil {
		return domain.ArrangementState{}, err
	}

	if !state.Created || state.Arrangement.CustomerNumber != customerNumber {
		return domain.ArrangementState{}, domain.ErrArrangementNotFound
	}

	if state.Ended(now) {
		return domain.ArrangementState{}, domain.ErrArrangementEnded
	}

	return a.end(ctx, state, domain.WithdrawArrangement{WithdrawnAt: now, Method: method})
}

func (a Arrangements) IneligibilityEnds(
	ctx context.Context,
	customerNumber string,
	at time.Time,
) ([]domain.ArrangementState, error) {
	states, err := a.Reader.ArrangementsOf(ctx, customerNumber)
	if err != nil {
		return nil, fmt.Errorf("read the arrangements: %w", err)
	}

	return domain.IneligibilityEnds(states, at), nil
}

// EndForIneligibleCustomer ends at the time of the close each arrangement of the customer that did not end. It is
// safe to repeat, so that a job can end the arrangements after a stop between the close and the end.
func (a Arrangements) EndForIneligibleCustomer(ctx context.Context, customerNumber string, closedAt time.Time) error {
	ends, err := a.IneligibilityEnds(ctx, customerNumber, closedAt)
	if err != nil {
		return err
	}

	var errs []error

	for _, state := range ends {
		_, err := a.end(
			ctx,
			state,
			domain.EndArrangement{EndedAt: closedAt, Reason: domain.EndReasonConsumerNotEligible},
		)
		if err != nil && !errors.Is(err, domain.ErrArrangementEnded) {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// RemoveSoftwareProduct is safe to repeat.
func (a Arrangements) RemoveSoftwareProduct(ctx context.Context, softwareID string, removedAt time.Time) error {
	states, err := a.Reader.ArrangementsOfSoftwareProduct(ctx, softwareID)
	if err != nil {
		return fmt.Errorf("read the arrangements: %w", err)
	}

	for _, state := range states {
		if state.Ended(removedAt) {
			continue
		}

		_, err := a.end(
			ctx,
			state,
			domain.EndArrangement{EndedAt: removedAt, Reason: domain.EndReasonSoftwareProductRemoved},
		)
		if err != nil && !errors.Is(err, domain.ErrArrangementEnded) {
			return err
		}
	}

	registration, _, err := ClientRegistrationType.Load(ctx, a.Store, softwareID)
	if err != nil {
		return fmt.Errorf("read the registration: %w", err)
	}

	if !registration.Registered {
		return nil
	}

	err = ClientRegistrationType.Execute(
		ctx, a.Store, softwareID, domain.DeleteClient{ClientID: registration.Registration.ClientID},
	)
	if err != nil && !errors.Is(err, domain.ErrNoRegistration) {
		return fmt.Errorf("delete the registration: %w", err)
	}

	return nil
}

// EndForDeletedRegistration ends each arrangement of the client as a revocation by the data recipient, because the
// software product asked for the deletion.
func (a Arrangements) EndForDeletedRegistration(
	ctx context.Context,
	softwareID, clientID string,
	deletedAt time.Time,
) error {
	states, err := a.Reader.ArrangementsOfSoftwareProduct(ctx, softwareID)
	if err != nil {
		return fmt.Errorf("read the arrangements: %w", err)
	}

	for _, state := range states {
		if state.Arrangement.ClientID != clientID || state.Ended(deletedAt) {
			continue
		}

		_, err := a.end(
			ctx,
			state,
			domain.EndArrangement{EndedAt: deletedAt, Reason: domain.EndReasonRevokedByDataRecipient},
		)
		if err != nil && !errors.Is(err, domain.ErrArrangementEnded) {
			return err
		}
	}

	return nil
}

// OfCustomer gives ErrCustomerNotFound.
func (a Arrangements) OfCustomer(ctx context.Context, customerNumber string) ([]domain.ArrangementState, error) {
	if _, err := a.Customers.Customer(ctx, customerNumber); err != nil {
		return nil, fmt.Errorf("read the customer: %w", err)
	}

	arrangements, err := a.Reader.ArrangementsOf(ctx, customerNumber)
	if err != nil {
		return nil, fmt.Errorf("read the arrangements: %w", err)
	}

	return arrangements, nil
}

func (a Arrangements) load(ctx context.Context, arrangementID string) (domain.ArrangementState, error) {
	state, _, err := ArrangementType.Load(ctx, a.Store, arrangementID)
	if err != nil {
		return domain.ArrangementState{}, fmt.Errorf("read the arrangement: %w", err)
	}

	return state, nil
}

// end revokes the grant before it records the end, so that a request that fails between the two leaves no valid
// token, and a new request can end the arrangement.
func (a Arrangements) end(
	ctx context.Context,
	state domain.ArrangementState,
	command domain.ArrangementCommand,
) (domain.ArrangementState, error) {
	arrangementID := state.Arrangement.ArrangementID
	if err := a.Grants.RevokeArrangement(ctx, arrangementID); err != nil {
		return domain.ArrangementState{}, fmt.Errorf("revoke the grant: %w", err)
	}

	if err := ArrangementType.Execute(ctx, a.Store, arrangementID, command); err != nil {
		return domain.ArrangementState{}, fmt.Errorf("end the arrangement: %w", err)
	}

	return a.load(ctx, arrangementID)
}
