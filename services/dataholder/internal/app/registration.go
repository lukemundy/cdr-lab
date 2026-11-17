package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var (
	// ErrInvalidJWT means that the JWT is not a JWT, or that its signature does not verify with the keys of its signer.
	ErrInvalidJWT = errors.New("the JWT is not valid")
	// ErrKeysNotAvailable means that the data holder cannot get the keys of a software product.
	ErrKeysNotAvailable     = errors.New("the keys of the software product are not available")
	ErrJTIUsed              = errors.New("the JWT ID is used again")
	ErrRegistrationNotFound = errors.New("the data holder has no registration with this client ID")
)

var ClientRegistrationType = eventsourcing.AggregateType[
	domain.ClientRegistrationState, domain.ClientRegistrationCommand, domain.ClientRegistrationEvent,
]{
	StreamType: "client-registration",
	Handle:     domain.HandleClientRegistration,
	Apply:      domain.ApplyClientRegistration,
	Codec:      clientRegistrationCodec(),
}

func clientRegistrationCodec() eventsourcing.Codec[domain.ClientRegistrationEvent] {
	c := eventsourcing.NewJSONCodec[domain.ClientRegistrationEvent]()
	eventsourcing.RegisterJSON[domain.ClientRegistrationEvent, domain.ClientRegistered](c, "ClientRegistered")
	eventsourcing.RegisterJSON[domain.ClientRegistrationEvent, domain.ClientUpdated](c, "ClientUpdated")
	eventsourcing.RegisterJSON[domain.ClientRegistrationEvent, domain.ClientDeleted](c, "ClientDeleted")

	return c
}

// RegistrationJWTs reads the two JWTs of a registration request. A method gives ErrInvalidJWT for a fault of the JWT.
type RegistrationJWTs interface {
	// SoftwareStatement gives the SSA of a request JWT, and does not verify the request JWT.
	SoftwareStatement(requestJWT string) (string, error)
	// VerifySSA verifies the SSA with the keys of the Register.
	VerifySSA(ctx context.Context, ssa string) (domain.SSA, error)
	// VerifyRequest verifies the request JWT with the keys at jwksURI, and gives ErrKeysNotAvailable when it cannot get
	// them.
	VerifyRequest(ctx context.Context, requestJWT, jwksURI string) (domain.RegistrationRequest, error)
}

type RegistrationReader interface {
	// Registration gives ErrRegistrationNotFound when no registration has the client ID.
	Registration(ctx context.Context, clientID string) (domain.Registration, error)
}

type RegistrationRecord struct {
	Registration domain.Registration
	// RegisteredAt is the time of the last registration of the software product, and UpdatedAt the time of the last
	// change of that registration.
	RegisteredAt time.Time
	UpdatedAt    time.Time
	// Status is empty when the data holder has no status for the software product.
	Status domain.SoftwareProductStatus
	// Arrangements counts each arrangement of the client ID, and CurrentArrangements only those that did not end at
	// the time of the query.
	Arrangements        int
	CurrentArrangements int
}

type RegistrationSort string

const (
	SortRegistrationsByClientName          RegistrationSort = "client-name"
	SortRegistrationsByRegisteredAt        RegistrationSort = "registered-at"
	SortRegistrationsByCurrentArrangements RegistrationSort = "current-arrangements"
	SortRegistrationsByArrangements        RegistrationSort = "arrangements"
)

type RegistrationListQuery struct {
	Sort       RegistrationSort
	Descending bool
	// At is the time at which an arrangement that ends at or before it has ended.
	At time.Time
}

type ClientArrangement struct {
	State     domain.ArrangementState
	FirstName string
	LastName  string
}

type RegistrationRecords interface {
	// RegistrationList sorts by the client ID after the sort of the query, and the client_name sort has no difference
	// for case.
	RegistrationList(ctx context.Context, query RegistrationListQuery) ([]RegistrationRecord, error)
	// RegistrationRecord gives ErrRegistrationNotFound when no registration has the client ID, and at is as in
	// RegistrationListQuery.
	RegistrationRecord(ctx context.Context, clientID string, at time.Time) (RegistrationRecord, error)
}

type RegistrationArrangements interface {
	// EndForDeletedRegistration ends each arrangement of the client that did not end, and is safe to repeat.
	EndForDeletedRegistration(ctx context.Context, softwareID, clientID string, deletedAt time.Time) error
}

type JTIStore interface {
	// Consume gives ErrJTIUsed when the JWT ID was consumed before.
	Consume(ctx context.Context, jti string) error
	// ConsumeSSA gives ErrJTIUsed when the JWT ID of an SSA was consumed before. It keeps the JWT ID until expiresAt.
	ConsumeSSA(ctx context.Context, jti string, expiresAt time.Time) error
}

// maxAttempts is the number of times that an update or a deletion reads the stream again after a version conflict.
const maxAttempts = 3

type Registrar struct {
	// Issuer is the issuer of the data holder, which a request JWT has as its audience.
	Issuer string
	JWTs   RegistrationJWTs
	JTIs   JTIStore
	Store  eventsourcing.Store
	// Registrations gives the software product of a client ID, which is the stream of its registration.
	Registrations RegistrationReader
	Statuses      StatusReader
	Arrangements  RegistrationArrangements
	Now           func() time.Time
	NewID         func() string
}

func refused(code domain.RegistrationErrorCode, description string) error {
	return domain.RegistrationError{Code: code, Description: description}
}

// Register gives a domain.RegistrationError when it refuses the request.
func (r Registrar) Register(ctx context.Context, requestJWT string) (domain.Registration, error) {
	now := r.Now()

	registration, jtis, err := r.verifiedRegistration(ctx, requestJWT, now)
	if err != nil {
		return domain.Registration{}, err
	}

	if err := r.consume(ctx, jtis); err != nil {
		return domain.Registration{}, err
	}

	registration.ClientID = r.NewID()
	registration.ClientIDIssuedAt = now.Unix()

	err = ClientRegistrationType.Execute(
		ctx,
		r.Store,
		registration.SoftwareID,
		domain.RegisterClient{Registration: registration},
	)
	// A different request registered the software product after this request read the stream.
	if errors.Is(err, eventsourcing.ErrVersionConflict) {
		return domain.Registration{}, refused(
			domain.InvalidSoftwareStatement,
			"the software_id of the SSA has a registration",
		)
	}

	if err != nil {
		return domain.Registration{}, fmt.Errorf("save the registration: %w", err)
	}

	return registration, nil
}

// Update gives a domain.RegistrationError when it refuses the request, and ErrRegistrationNotFound when no
// registration has the client ID.
func (r Registrar) Update(ctx context.Context, clientID, requestJWT string) (domain.Registration, error) {
	existing, err := r.Registrations.Registration(ctx, clientID)
	if err != nil {
		return domain.Registration{}, err
	}

	replacement, jtis, err := r.verifiedRegistration(ctx, requestJWT, r.Now())
	if err != nil {
		return domain.Registration{}, err
	}
	// The aggregate has the same rule. This check is before the jti is used, as each other check of the request is.
	if replacement.SoftwareID != existing.SoftwareID {
		return domain.Registration{}, refused(
			domain.InvalidSoftwareStatement,
			"the software_id of the SSA is not the software_id of the registration",
		)
	}

	if err := r.consume(ctx, jtis); err != nil {
		return domain.Registration{}, err
	}

	command := domain.UpdateClient{ClientID: clientID, Registration: replacement}
	if err := r.execute(ctx, existing.SoftwareID, command); err != nil {
		return domain.Registration{}, err
	}

	replacement.ClientID = existing.ClientID
	replacement.ClientIDIssuedAt = existing.ClientIDIssuedAt

	return replacement, nil
}

// Delete gives ErrRegistrationNotFound when no registration has the client ID. It ends the arrangements before it
// deletes the registration, so that a new request can end the arrangements after a failure.
func (r Registrar) Delete(ctx context.Context, clientID string) error {
	existing, err := r.Registrations.Registration(ctx, clientID)
	if err != nil {
		return err
	}

	if err := r.Arrangements.EndForDeletedRegistration(ctx, existing.SoftwareID, clientID, r.Now()); err != nil {
		return fmt.Errorf("end the arrangements of the registration: %w", err)
	}

	return r.execute(ctx, existing.SoftwareID, domain.DeleteClient{ClientID: clientID})
}

// execute reads the stream again after a version conflict, because the command applies to each state of a
// registration that has the client ID.
func (r Registrar) execute(ctx context.Context, softwareID string, command domain.ClientRegistrationCommand) error {
	var err error
	for range maxAttempts {
		err = ClientRegistrationType.Execute(ctx, r.Store, softwareID, command)
		if !errors.Is(err, eventsourcing.ErrVersionConflict) {
			break
		}
	}

	if errors.Is(err, domain.ErrNoRegistration) {
		return ErrRegistrationNotFound
	}

	if err != nil {
		return fmt.Errorf("change the registration: %w", err)
	}

	return nil
}

// verifiedRegistration gives the registration that a request asks for, with no client ID, and the JWT IDs that the
// data holder consumes when it accepts the request.
func (r Registrar) verifiedRegistration(
	ctx context.Context,
	requestJWT string,
	now time.Time,
) (domain.Registration, requestJTIs, error) {
	ssa, err := r.verifiedSSA(ctx, requestJWT, now)
	if err != nil {
		return domain.Registration{}, requestJTIs{}, err
	}

	if err := r.refuseRemoved(ctx, ssa.SoftwareID); err != nil {
		return domain.Registration{}, requestJTIs{}, err
	}

	// The SSA is verified first, so that the data holder connects only to a JWKS URI that the Register signed.
	request, err := r.JWTs.VerifyRequest(ctx, requestJWT, ssa.JWKSURI)
	switch {
	case errors.Is(err, ErrInvalidJWT):
		return domain.Registration{}, requestJTIs{}, refused(
			domain.InvalidClientMetadata,
			"the signature of the request JWT does not verify with a key at the jwks_uri of the SSA",
		)
	case errors.Is(err, ErrKeysNotAvailable):
		return domain.Registration{}, requestJTIs{}, refused(
			domain.InvalidClientMetadata,
			"the data holder cannot get the keys at the jwks_uri of the SSA",
		)
	case err != nil:
		return domain.Registration{}, requestJTIs{}, fmt.Errorf("verify the request JWT: %w", err)
	}

	registration, err := domain.NewRegistration(ssa, request, r.Issuer, now)
	jtis := requestJTIs{request: request.ID, ssa: ssa.ID, ssaExpiresAt: time.Unix(ssa.ExpiresAt, 0)}

	return registration, jtis, err
}

// refuseRemoved permits a software product with no status, because only REMOVED is final.
func (r Registrar) refuseRemoved(ctx context.Context, softwareID string) error {
	status, _, err := r.Statuses.Status(ctx, softwareID)
	if err != nil {
		return fmt.Errorf("read the status of the software product: %w", err)
	}

	if status == domain.SoftwareProductRemoved {
		return refused(domain.UnapprovedSoftwareStatement, "the software product is REMOVED")
	}

	return nil
}

type requestJTIs struct {
	request      string
	ssa          string
	ssaExpiresAt time.Time
}

// consume uses the jti of the request JWT first, so that a second request with the same request JWT gets the error of
// the request JWT.
func (r Registrar) consume(ctx context.Context, jtis requestJTIs) error {
	err := r.JTIs.Consume(ctx, jtis.request)
	if errors.Is(err, ErrJTIUsed) {
		return refused(domain.InvalidClientMetadata, "the jti of the request JWT was used before")
	}

	if err != nil {
		return fmt.Errorf("consume the jti of the request JWT: %w", err)
	}

	err = r.JTIs.ConsumeSSA(ctx, jtis.ssa, jtis.ssaExpiresAt)
	if errors.Is(err, ErrJTIUsed) {
		return refused(domain.InvalidSoftwareStatement, "the jti of the SSA was used before")
	}

	if err != nil {
		return fmt.Errorf("consume the jti of the SSA: %w", err)
	}

	return nil
}

func (r Registrar) verifiedSSA(ctx context.Context, requestJWT string, now time.Time) (domain.SSA, error) {
	statement, err := r.JWTs.SoftwareStatement(requestJWT)
	if errors.Is(err, ErrInvalidJWT) {
		return domain.SSA{}, refused(domain.InvalidClientMetadata, "the request is not a JWT")
	}

	if err != nil {
		return domain.SSA{}, fmt.Errorf("read the request JWT: %w", err)
	}

	if statement == "" {
		return domain.SSA{}, refused(domain.InvalidSoftwareStatement, "the request has no software_statement")
	}

	ssa, err := r.JWTs.VerifySSA(ctx, statement)
	if errors.Is(err, ErrInvalidJWT) {
		return domain.SSA{}, refused(
			domain.InvalidSoftwareStatement,
			"the SSA is not a JWT, or its signature does not verify with a key of the Register",
		)
	}

	if err != nil {
		return domain.SSA{}, fmt.Errorf("verify the SSA: %w", err)
	}

	return ssa, ssa.Validate(now)
}
