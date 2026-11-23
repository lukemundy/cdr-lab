package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// LoadParticipants adds the records of p that the Register does not have, in one append, or makes no changes. The
// error joins the errors of all legal entities.
func (o Operator) LoadParticipants(ctx context.Context, p domain.Participants) (domain.ParticipantLoad, error) {
	if err := p.Validate(); err != nil {
		return domain.ParticipantLoad{}, fmt.Errorf("participants are not valid:\n%w", err)
	}

	var (
		load    domain.ParticipantLoad
		appends []eventsourcing.Append
		errs    []error
	)

	add := func(a eventsourcing.Append, l domain.ParticipantLoad, err error) {
		load.Append(l)

		errs = append(errs, err)

		if err == nil && len(a.Events) > 0 {
			appends = append(appends, a)
		}
	}

	for _, dh := range p.DataHolders {
		add(loadAggregate(ctx, o.store, DataHolderType, dh.LegalEntityID,
			func(s domain.DataHolderState) ([]domain.DataHolderCommand, domain.ParticipantLoad) {
				return domain.DataHolderLoad(s, dh)
			}))
	}

	for _, dr := range p.DataRecipients {
		add(loadAggregate(ctx, o.store, DataRecipientType, dr.LegalEntityID,
			func(s domain.DataRecipientState) ([]domain.DataRecipientCommand, domain.ParticipantLoad) {
				return domain.DataRecipientLoad(s, dr)
			}))
	}

	if err := errors.Join(errs...); err != nil {
		return domain.ParticipantLoad{}, err
	}

	if len(appends) == 0 {
		return load, nil
	}

	return load, o.store.Append(ctx, appends...)
}

func loadAggregate[S, C, E any](
	ctx context.Context,
	store eventsourcing.Store,
	t eventsourcing.AggregateType[S, C, E],
	streamID string,
	commands func(S) ([]C, domain.ParticipantLoad),
) (eventsourcing.Append, domain.ParticipantLoad, error) {
	state, version, err := t.Load(ctx, store, streamID)
	if err != nil {
		return eventsourcing.Append{}, domain.ParticipantLoad{}, err
	}

	cs, load := commands(state)
	a, err := handleAll(t, streamID, state, version, cs...)

	return a, load, err
}
