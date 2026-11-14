package eventsourcing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
)

const jsonSchemaVersion = 1

type JSONCodec[E any] struct {
	goTypes    map[string]reflect.Type
	eventTypes map[reflect.Type]string
}

func NewJSONCodec[E any]() *JSONCodec[E] {
	return &JSONCodec[E]{goTypes: map[string]reflect.Type{}, eventTypes: map[reflect.Type]string{}}
}

// RegisterJSON panics if T does not implement E, or if T or eventType is already registered.
func RegisterJSON[E, T any](codec *JSONCodec[E], eventType string) {
	goType := reflect.TypeFor[T]()
	if _, ok := any(*new(T)).(E); !ok {
		panic(fmt.Sprintf("eventsourcing: %s does not implement %s", goType, reflect.TypeFor[E]()))
	}

	if _, ok := codec.goTypes[eventType]; ok {
		panic(fmt.Sprintf("eventsourcing: event type %q is already registered", eventType))
	}

	if _, ok := codec.eventTypes[goType]; ok {
		panic(fmt.Sprintf("eventsourcing: %s is already registered", goType))
	}

	codec.goTypes[eventType] = goType
	codec.eventTypes[goType] = eventType
}

func (c *JSONCodec[E]) EventTypes() []string {
	return slices.Sorted(maps.Keys(c.goTypes))
}

func (c *JSONCodec[E]) Encode(event E) (EventData, error) {
	eventType, ok := c.eventTypes[reflect.TypeOf(event)]
	if !ok {
		return EventData{}, fmt.Errorf("%T is not registered", event)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return EventData{}, err
	}

	return EventData{EventType: eventType, SchemaVersion: jsonSchemaVersion, Payload: payload}, nil
}

// Decode rejects payload fields that the Go type does not have.
func (c *JSONCodec[E]) Decode(data EventData) (E, error) {
	var zero E

	goType, ok := c.goTypes[data.EventType]
	if !ok {
		return zero, fmt.Errorf("event type %q is not registered", data.EventType)
	}

	if data.SchemaVersion != jsonSchemaVersion {
		return zero, fmt.Errorf("event type %q: schema version %d is not supported", data.EventType, data.SchemaVersion)
	}

	event := reflect.New(goType)
	decoder := json.NewDecoder(bytes.NewReader(data.Payload))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(event.Interface()); err != nil {
		return zero, fmt.Errorf("event type %q: %w", data.EventType, err)
	}

	e, _ := reflect.TypeAssert[E](event.Elem()) // RegisterJSON checked that goType implements E.

	return e, nil
}
