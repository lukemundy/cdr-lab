// Package eventsourcingtest keeps the event golden file of a service: one encoded sample of each event type that the
// codecs of the service register (tools/release/DESIGN.md, Event golden files).
package eventsourcingtest

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
)

type Golden struct {
	path    string
	entries map[string]entry
	checks  map[string]func(eventsourcing.EventData) error
	errs    []error
}

type entry struct {
	SchemaVersion int             `json:"schemaVersion"`
	Payload       json.RawMessage `json:"payload"`
}

func NewGolden(path string) *Golden {
	return &Golden{path: path, entries: map[string]entry{}, checks: map[string]func(eventsourcing.EventData) error{}}
}

// Add needs exactly one sample of each event type that the codec of the aggregate type registers.
func Add[S, C, E any](g *Golden, aggregate eventsourcing.AggregateType[S, C, E], samples ...E) {
	lister, ok := aggregate.Codec.(interface{ EventTypes() []string })
	if !ok {
		g.errs = append(g.errs, fmt.Errorf("%s: the codec has no EventTypes method", aggregate.StreamType))
		return
	}

	for _, sample := range samples {
		data, err := aggregate.Codec.Encode(sample)
		if err != nil {
			g.errs = append(g.errs, fmt.Errorf("%s: %w", aggregate.StreamType, err))
			continue
		}

		key := aggregate.StreamType + "/" + data.EventType
		if _, ok := g.entries[key]; ok {
			g.errs = append(g.errs, fmt.Errorf("%s: more than one sample", key))
			continue
		}

		for _, field := range zeroFields(reflect.ValueOf(sample), "") {
			g.errs = append(g.errs, fmt.Errorf("%s: the field %s has the zero value", key, field))
		}

		g.entries[key] = entry{SchemaVersion: data.SchemaVersion, Payload: data.Payload}
		g.checks[key] = func(data eventsourcing.EventData) error {
			decoded, err := aggregate.Codec.Decode(data)
			if err != nil {
				return fmt.Errorf("%s: the codec cannot decode the entry: %w", key, err)
			}

			if !reflect.DeepEqual(any(decoded), any(sample)) {
				return fmt.Errorf("%s: the codec does not decode the entry to its sample", key)
			}

			return nil
		}
	}

	for _, eventType := range lister.EventTypes() {
		if key := aggregate.StreamType + "/" + eventType; g.entries[key].Payload == nil {
			g.errs = append(g.errs, fmt.Errorf("%s: the registered event type has no sample", key))
		}
	}
}

func (g *Golden) Check(t testing.TB, update bool) {
	t.Helper()
	require.NoError(t, g.Verify(update))
}

// Verify writes the file only when update is true and each sample is valid.
func (g *Golden) Verify(update bool) error {
	for _, key := range slices.Sorted(maps.Keys(g.checks)) {
		e := g.entries[key]
		_, eventType, _ := strings.Cut(key, "/")

		data := eventsourcing.EventData{EventType: eventType, SchemaVersion: e.SchemaVersion, Payload: e.Payload}
		if err := g.checks[key](data); err != nil {
			g.errs = append(g.errs, err)
		}
	}

	if len(g.errs) > 0 {
		return errors.Join(g.errs...)
	}

	want, err := json.MarshalIndent(g.entries, "", "  ")
	if err != nil {
		return err
	}

	want = append(want, '\n')

	if update {
		if err := os.MkdirAll(filepath.Dir(g.path), 0o750); err != nil {
			return err
		}

		return os.WriteFile(g.path, want, 0o600)
	}

	got, err := os.ReadFile(g.path)
	if err != nil {
		return fmt.Errorf("%w. %s", err, howToUpdate)
	}

	if bytes.Equal(got, want) {
		return nil
	}

	return g.difference(got)
}

const howToUpdate = "When the change of the stored events is intended, run the test again with the flag -update, " +
	"and review the diff of the event golden file: " +
	"a changed or removed entry makes the next release need a fresh deployment"

func (g *Golden) difference(got []byte) error {
	var file map[string]entry
	if err := json.Unmarshal(got, &file); err != nil {
		return fmt.Errorf("%s: %w. %s", g.path, err, howToUpdate)
	}

	var changed, missing, extra []string

	for key, e := range g.entries {
		f, ok := file[key]

		switch {
		case !ok:
			missing = append(missing, key)
		case f.SchemaVersion != e.SchemaVersion || !jsonEqual(f.Payload, e.Payload):
			changed = append(changed, key)
		}
	}

	for key := range file {
		if _, ok := g.entries[key]; !ok {
			extra = append(extra, key)
		}
	}

	var lines []string

	for _, group := range []struct {
		text string
		keys []string
	}{
		{"entries that are not the same as their sample", changed},
		{"samples that have no entry", missing},
		{"entries that have no registered event type", extra},
	} {
		if len(group.keys) > 0 {
			slices.Sort(group.keys)
			lines = append(lines, group.text+": "+strings.Join(group.keys, ", "))
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "the format of the file is not the expected format")
	}

	return fmt.Errorf("%s is not the same as the samples:\n%s\n%s", g.path, strings.Join(lines, "\n"), howToUpdate)
}

func jsonEqual(a, b json.RawMessage) bool {
	var x, y bytes.Buffer

	return json.Compact(&x, a) == nil && json.Compact(&y, b) == nil && bytes.Equal(x.Bytes(), y.Bytes())
}

var (
	jsonMarshaler = reflect.TypeFor[json.Marshaler]()
	textMarshaler = reflect.TypeFor[encoding.TextMarshaler]()
)

// zeroFields gives the JSON path of each value in v that has the zero value, because the encoding can leave out such
// a value, and then the golden file does not show its name and its type.
func zeroFields(v reflect.Value, path string) []string {
	if v.Type().Implements(jsonMarshaler) || v.Type().Implements(textMarshaler) {
		if v.IsZero() {
			return []string{path}
		}

		return nil
	}

	switch kind := v.Kind(); {
	case kind == reflect.Pointer || kind == reflect.Interface:
		if v.IsNil() {
			return []string{path}
		}

		return zeroFields(v.Elem(), path)
	case kind == reflect.Struct:
		return zeroStructFields(v, path)
	case kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map:
		return zeroItems(v, path)
	case v.IsZero():
		return []string{path}
	default:
		return nil
	}
}

func zeroStructFields(v reflect.Value, path string) []string {
	var fields []string

	for i := range v.NumField() {
		field := v.Type().Field(i)

		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" || (!field.IsExported() && !field.Anonymous) {
			continue
		}

		if field.Anonymous && name == "" {
			fields = append(fields, zeroFields(v.Field(i), path)...)
			continue
		}

		if name == "" {
			name = field.Name
		}

		fields = append(fields, zeroFields(v.Field(i), join(path, name))...)
	}

	return fields
}

func zeroItems(v reflect.Value, path string) []string {
	if v.Len() == 0 {
		return []string{path}
	}

	var fields []string

	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			fields = append(fields, zeroFields(v.MapIndex(key), fmt.Sprintf("%s[%v]", path, key))...)
		}

		return fields
	}

	for i := range v.Len() {
		fields = append(fields, zeroFields(v.Index(i), fmt.Sprintf("%s[%d]", path, i))...)
	}

	return fields
}

func join(path, name string) string {
	if path == "" {
		return name
	}

	return path + "." + name
}
