package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var ErrNotEmpty = errors.New("the Register already has participants")

type EventStore interface {
	eventsourcing.Store
	IsEmpty(ctx context.Context) (bool, error)
}

// Seed loads a full set of participants into an empty store, or makes no changes.
func Seed(ctx context.Context, store EventStore, p domain.Participants) error {
	errs := []error{p.Validate()}

	var appends []eventsourcing.Append

	for _, dh := range p.DataHolders {
		a, err := handleAll(
			DataHolderType,
			dh.LegalEntityID,
			DataHolderType.Initial,
			0,
			domain.DataHolderCommand(domain.ImportDataHolder{DataHolder: dh}),
		)
		errs = append(errs, err)
		appends = append(appends, a)
	}

	for _, dr := range p.DataRecipients {
		a, err := handleAll(
			DataRecipientType,
			dr.LegalEntityID,
			DataRecipientType.Initial,
			0,
			domain.DataRecipientCommand(domain.ImportDataRecipient{DataRecipient: dr}),
		)
		errs = append(errs, err)
		appends = append(appends, a)
	}

	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("participants are not valid:\n%w", err)
	}

	empty, err := store.IsEmpty(ctx)
	if err != nil {
		return err
	}

	if !empty {
		return ErrNotEmpty
	}

	return store.Append(ctx, appends...)
}

// handleAll handles each command on the state after the events of the commands before it, and gives one Append of
// all events at the version of the state.
func handleAll[S, C, E any](
	t eventsourcing.AggregateType[S, C, E],
	streamID string,
	state S,
	version int64,
	commands ...C,
) (eventsourcing.Append, error) {
	var events []E

	for _, c := range commands {
		handled, err := t.Handle(state, c)
		if err != nil {
			return eventsourcing.Append{}, err
		}

		for _, e := range handled {
			state = t.Apply(state, e)
		}

		events = append(events, handled...)
	}

	data, err := t.Encode(events)

	return eventsourcing.Append{
		StreamType: t.StreamType, StreamID: streamID, ExpectedVersion: version, Events: data,
	}, err
}
