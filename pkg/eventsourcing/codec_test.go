package eventsourcing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
)

func TestJSONCodecEncodesAndDecodesRegisteredEvents(t *testing.T) {
	codec := newCounterCodec()

	for _, event := range []counterEvent{incremented{By: 2}, reset{}} {
		data, err := codec.Encode(event)
		require.NoError(t, err)

		got, err := codec.Decode(data)
		require.NoError(t, err)
		assert.Equal(t, event, got)
	}
}

func TestJSONCodecGivesTheEventTypeAndSchemaVersion(t *testing.T) {
	data, err := newCounterCodec().Encode(incremented{By: 2})
	require.NoError(t, err)

	assert.Equal(t, "Incremented", data.EventType)
	assert.Equal(t, 1, data.SchemaVersion)
	assert.JSONEq(t, `{"by": 2}`, string(data.Payload))
}

func TestJSONCodecRejectsAnEventThatIsNotRegistered(t *testing.T) {
	codec := eventsourcing.NewJSONCodec[counterEvent]()

	_, err := codec.Encode(incremented{By: 1})

	assert.ErrorContains(t, err, "not registered")
}

func TestJSONCodecRejectsAnUnknownEventType(t *testing.T) {
	_, err := newCounterCodec().Decode(
		eventsourcing.EventData{EventType: "Unknown", SchemaVersion: 1, Payload: []byte(`{}`)},
	)

	assert.ErrorContains(t, err, `event type "Unknown"`)
}

func TestJSONCodecRejectsAnUnsupportedSchemaVersion(t *testing.T) {
	_, err := newCounterCodec().Decode(
		eventsourcing.EventData{EventType: "Incremented", SchemaVersion: 2, Payload: []byte(`{}`)},
	)

	assert.ErrorContains(t, err, "schema version 2")
}

func TestJSONCodecRejectsUnknownPayloadFields(t *testing.T) {
	_, err := newCounterCodec().Decode(
		eventsourcing.EventData{
			EventType:     "Incremented",
			SchemaVersion: 1,
			Payload:       []byte(`{"by": 1, "extra": true}`),
		},
	)

	assert.ErrorContains(t, err, "extra")
}

func TestRegisterJSONPanicsForADuplicateEventType(t *testing.T) {
	codec := newCounterCodec()

	assert.Panics(t, func() { eventsourcing.RegisterJSON[counterEvent, reset](codec, "Incremented") })
}

func TestRegisterJSONPanicsForADuplicateGoType(t *testing.T) {
	codec := newCounterCodec()

	assert.Panics(t, func() { eventsourcing.RegisterJSON[counterEvent, reset](codec, "ResetAgain") })
}

func TestRegisterJSONPanicsForATypeThatIsNotAnEvent(t *testing.T) {
	codec := eventsourcing.NewJSONCodec[counterEvent]()

	assert.Panics(t, func() { eventsourcing.RegisterJSON[counterEvent, string](codec, "String") })
}

func TestJSONCodecGivesEachRegisteredEventTypeInOrder(t *testing.T) {
	assert.Equal(t, []string{"Incremented", "Reset"}, newCounterCodec().EventTypes())
}
