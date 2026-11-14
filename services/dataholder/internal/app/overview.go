package app

import (
	"context"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const NewestChanges = 10

type ChangeKind string

const (
	ChangeCustomerCreated      ChangeKind = "CUSTOMER_CREATED"
	ChangeAccountCreated       ChangeKind = "ACCOUNT_CREATED"
	ChangeAccountClosed        ChangeKind = "ACCOUNT_CLOSED"
	ChangeArrangementConfirmed ChangeKind = "ARRANGEMENT_CONFIRMED"
	// ChangeArrangementEnded is also the end of the sharing, which has no EndReason.
	ChangeArrangementEnded ChangeKind = "ARRANGEMENT_ENDED"
	ChangeClientRegistered ChangeKind = "CLIENT_REGISTERED"
	ChangeClientUpdated    ChangeKind = "CLIENT_UPDATED"
	ChangeClientDeleted    ChangeKind = "CLIENT_DELETED"
)

// Change has no name, email address, display name or account number of a customer, so that the overview shows no
// personal information.
type Change struct {
	Time time.Time
	Kind ChangeKind
	// CustomerNumber is empty for a change of a client registration.
	CustomerNumber string
	// ClientID and ClientName are empty for a change of a customer or an account.
	ClientID   string
	ClientName string
	// ClientRegistered is true when a registration has the client ID now.
	ClientRegistered bool
	// ProductName is set only for a change of an account.
	ProductName string
	// SharingEndsAt is set only for ChangeArrangementConfirmed.
	SharingEndsAt    time.Time
	EndReason        domain.EndReason
	WithdrawalMethod domain.WithdrawalMethod
}

type Overview struct {
	Customers           int
	OpenAccounts        int
	ClosedAccounts      int
	CurrentArrangements int
	EndedArrangements   int
	Registrations       int
	// SoftwareProductsNotActive are the registered software products with a Register status that is not ACTIVE.
	SoftwareProductsNotActive int
	// Changes are newest first.
	Changes []Change
}

// OverviewReader reads the overview in one snapshot, so that the counts and the changes agree. An arrangement that
// ends at or before at has ended.
type OverviewReader interface {
	Overview(ctx context.Context, at time.Time, changes int) (Overview, error)
}
