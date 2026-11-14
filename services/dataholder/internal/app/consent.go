package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// ArrangementType has one stream for each CDR arrangement, whose stream ID is the cdr_arrangement_id.
var ArrangementType = eventsourcing.AggregateType[
	domain.ArrangementState, domain.ArrangementCommand, domain.ArrangementEvent,
]{
	StreamType: "arrangement",
	Handle:     domain.HandleArrangement,
	Apply:      domain.ApplyArrangement,
	Codec:      arrangementCodec(),
}

func arrangementCodec() eventsourcing.Codec[domain.ArrangementEvent] {
	c := eventsourcing.NewJSONCodec[domain.ArrangementEvent]()
	eventsourcing.RegisterJSON[domain.ArrangementEvent, domain.ArrangementCreated](c, "ArrangementCreated")
	eventsourcing.RegisterJSON[domain.ArrangementEvent, domain.ArrangementEnded](c, "ArrangementEnded")
	eventsourcing.RegisterJSON[domain.ArrangementEvent, domain.ArrangementWithdrawn](c, "ArrangementWithdrawn")
	eventsourcing.RegisterJSON[domain.ArrangementEvent, domain.NoticeAttempted](c, "NoticeAttempted")
	eventsourcing.RegisterJSON[domain.ArrangementEvent, domain.NoticeRecorded](c, "NoticeRecorded")

	return c
}

// SubjectStore keeps the pairwise sub of each customer for each sector identifier.
type SubjectStore interface {
	// Subject gives the sub of the pair. When the pair has none, it keeps candidate as the sub and gives it.
	Subject(ctx context.Context, customerNumber, sectorIdentifier, candidate string) (string, error)
}

type Consents struct {
	Customers CustomerStore
	Subjects  SubjectStore
	Store     eventsourcing.Store
	// NewID gives a random version 4 UUID, for a cdr_arrangement_id or a sub.
	NewID func() string
	Now   func() time.Time
}

type Confirmation struct {
	Arrangement domain.Arrangement
	// Subject is the sub of the customer for the sector identifier of the request.
	Subject string
}

// Accounts gives the accounts that the customer can select, or ErrCustomerNotFound.
func (c Consents) Accounts(ctx context.Context, customerNumber string) ([]domain.Account, error) {
	record, err := c.Customers.Customer(ctx, customerNumber)
	if err != nil {
		return nil, fmt.Errorf("read the customer: %w", err)
	}

	return domain.OpenAccounts(record.Accounts), nil
}

// Confirm creates the arrangement, or gives the errors of domain.NewArrangement and of Accounts.
func (c Consents) Confirm(
	ctx context.Context,
	request domain.ConsentRequest,
	selected []string,
) (Confirmation, error) {
	accounts, err := c.Accounts(ctx, request.CustomerNumber)
	if err != nil {
		return Confirmation{}, err
	}

	arrangement, err := domain.NewArrangement(c.NewID(), request, accounts, selected, c.Now())
	if err != nil {
		return Confirmation{}, err
	}

	subject, err := c.Subjects.Subject(ctx, request.CustomerNumber, request.SectorIdentifier, c.NewID())
	if err != nil {
		return Confirmation{}, fmt.Errorf("get the subject identifier: %w", err)
	}

	err = ArrangementType.Execute(
		ctx, c.Store, arrangement.ArrangementID, domain.CreateArrangement{Arrangement: arrangement},
	)
	if err != nil {
		return Confirmation{}, fmt.Errorf("create the arrangement: %w", err)
	}

	return Confirmation{Arrangement: arrangement, Subject: subject}, nil
}
