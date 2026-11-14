package eventsourcing_test

import (
	"context"
	"errors"
	"slices"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
)

const counterLimit = 10

var errLimit = errors.New("counter limit exceeded")

type counterEvent interface{ isCounterEvent() }

type incremented struct {
	By int `json:"by"`
}

type reset struct{}

func (incremented) isCounterEvent() {}
func (reset) isCounterEvent()       {}

type counterCommand interface{ isCounterCommand() }

type increment struct{ By int }

type resetCounter struct{}

func (increment) isCounterCommand()    {}
func (resetCounter) isCounterCommand() {}

func handleCounter(state int, command counterCommand) ([]counterEvent, error) {
	switch c := command.(type) {
	case increment:
		if state+c.By > counterLimit {
			return nil, errLimit
		}

		return []counterEvent{incremented(c)}, nil
	case resetCounter:
		if state == 0 {
			return nil, nil
		}

		return []counterEvent{reset{}}, nil
	}

	return nil, errors.New("unknown command")
}

func applyCounter(state int, event counterEvent) int {
	switch e := event.(type) {
	case incremented:
		return state + e.By
	case reset:
		return 0
	}

	return state
}

func newCounterCodec() *eventsourcing.JSONCodec[counterEvent] {
	codec := eventsourcing.NewJSONCodec[counterEvent]()
	eventsourcing.RegisterJSON[counterEvent, incremented](codec, "Incremented")
	eventsourcing.RegisterJSON[counterEvent, reset](codec, "Reset")

	return codec
}

func newCounterType() eventsourcing.AggregateType[int, counterCommand, counterEvent] {
	return eventsourcing.AggregateType[int, counterCommand, counterEvent]{
		StreamType: "counter",
		Handle:     handleCounter,
		Apply:      applyCounter,
		Codec:      newCounterCodec(),
	}
}

type memoryStore struct {
	streams map[string][]eventsourcing.RecordedEvent
	appends [][]eventsourcing.Append
	err     error
}

func newMemoryStore() *memoryStore {
	return &memoryStore{streams: map[string][]eventsourcing.RecordedEvent{}}
}

func (s *memoryStore) Read(_ context.Context, streamType, streamID string) ([]eventsourcing.RecordedEvent, error) {
	return slices.Clone(s.streams[streamType+"/"+streamID]), nil
}

func (s *memoryStore) Append(_ context.Context, appends ...eventsourcing.Append) error {
	if s.err != nil {
		return s.err
	}

	for _, a := range appends {
		if int64(len(s.streams[a.StreamType+"/"+a.StreamID])) != a.ExpectedVersion {
			return eventsourcing.ErrVersionConflict
		}
	}

	s.appends = append(s.appends, appends)
	for _, a := range appends {
		key := a.StreamType + "/" + a.StreamID
		for _, data := range a.Events {
			s.streams[key] = append(s.streams[key], eventsourcing.RecordedEvent{
				EventData:     data,
				StreamVersion: int64(len(s.streams[key]) + 1),
			})
		}
	}

	return nil
}
