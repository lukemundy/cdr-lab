// Package eventsourcing contains the generic types and functions of event sourcing.
package eventsourcing

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrVersionConflict = errors.New("stream version is not the expected version")

type AggregateType[S, C, E any] struct {
	StreamType string
	Initial    S
	// Handle and Apply must be pure.
	Handle func(S, C) ([]E, error)
	Apply  func(S, E) S
	Codec  Codec[E]
}

type Codec[E any] interface {
	Encode(E) (EventData, error)
	Decode(EventData) (E, error)
}

type EventData struct {
	EventType     string
	SchemaVersion int
	Payload       []byte
}

type RecordedEvent struct {
	EventData
	StreamVersion  int64
	GlobalPosition int64
	RecordedAt     time.Time
}

type Append struct {
	StreamType string
	StreamID   string
	// ExpectedVersion 0 means that the stream must not exist.
	ExpectedVersion int64
	Events          []EventData
}

type Store interface {
	// Read returns no events and no error for a stream that does not exist.
	Read(ctx context.Context, streamType, streamID string) ([]RecordedEvent, error)
	// Append is atomic: if one stream is not at its expected version, it adds no events and returns ErrVersionConflict.
	Append(ctx context.Context, appends ...Append) error
}

func (t AggregateType[S, C, E]) Load(
	ctx context.Context,
	store Store,
	streamID string,
) (state S, version int64, err error) {
	recorded, err := store.Read(ctx, t.StreamType, streamID)
	if err != nil {
		return t.Initial, 0, err
	}

	return t.Fold(streamID, recorded)
}

func (t AggregateType[S, C, E]) Fold(streamID string, recorded []RecordedEvent) (state S, version int64, err error) {
	state = t.Initial
	for _, r := range recorded {
		event, err := t.Codec.Decode(r.EventData)
		if err != nil {
			return t.Initial, 0, fmt.Errorf("%s %s version %d: %w", t.StreamType, streamID, r.StreamVersion, err)
		}

		state = t.Apply(state, event)
		version = r.StreamVersion
	}

	return state, version, nil
}

func (t AggregateType[S, C, E]) Execute(ctx context.Context, store Store, streamID string, command C) error {
	state, version, err := t.Load(ctx, store, streamID)
	if err != nil {
		return err
	}

	return t.handleAndAppend(ctx, store, streamID, state, version, command)
}

// ExecuteAt refuses a stream at a different version before Handle, and lets Handle decide for a new stream.
func (t AggregateType[S, C, E]) ExecuteAt(
	ctx context.Context,
	store Store,
	streamID string,
	expectedVersion int64,
	command C,
) error {
	state, version, err := t.Load(ctx, store, streamID)
	if err != nil {
		return err
	}

	if version != 0 && version != expectedVersion {
		return ErrVersionConflict
	}

	return t.handleAndAppend(ctx, store, streamID, state, version, command)
}

func (t AggregateType[S, C, E]) handleAndAppend(
	ctx context.Context,
	store Store,
	streamID string,
	state S,
	version int64,
	command C,
) error {
	events, err := t.Handle(state, command)
	if err != nil || len(events) == 0 {
		return err
	}

	data, err := t.Encode(events)
	if err != nil {
		return err
	}

	return store.Append(
		ctx,
		Append{StreamType: t.StreamType, StreamID: streamID, ExpectedVersion: version, Events: data},
	)
}

func (t AggregateType[S, C, E]) Encode(events []E) ([]EventData, error) {
	data := make([]EventData, 0, len(events))
	for _, e := range events {
		d, err := t.Codec.Encode(e)
		if err != nil {
			return nil, err
		}

		data = append(data, d)
	}

	return data, nil
}
