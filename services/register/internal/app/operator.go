package app

import (
	"context"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type Operator struct {
	store eventsourcing.Store
}

func NewOperator(store eventsourcing.Store) Operator {
	return Operator{store: store}
}

func (o Operator) Execute(ctx context.Context, legalEntityID string, command domain.DataRecipientCommand) error {
	return DataRecipientType.Execute(ctx, o.store, legalEntityID, command)
}

func (o Operator) ExecuteDataHolder(
	ctx context.Context,
	legalEntityID string,
	command domain.DataHolderCommand,
) error {
	return DataHolderType.Execute(ctx, o.store, legalEntityID, command)
}

func (o Operator) ExecuteAt(
	ctx context.Context,
	legalEntityID string,
	version int64,
	command domain.DataRecipientCommand,
) error {
	return DataRecipientType.ExecuteAt(ctx, o.store, legalEntityID, version, command)
}

func (o Operator) ExecuteDataHolderAt(
	ctx context.Context,
	legalEntityID string,
	version int64,
	command domain.DataHolderCommand,
) error {
	return DataHolderType.ExecuteAt(ctx, o.store, legalEntityID, version, command)
}
