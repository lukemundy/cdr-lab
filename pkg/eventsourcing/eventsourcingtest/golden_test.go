package eventsourcingtest_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventsourcingtest"
)

const (
	write   = true
	compare = false
)

type event interface{ isEvent() }

type opened struct {
	Name    string            `json:"name"`
	Note    string            `json:"note,omitempty"`
	At      time.Time         `json:"at"`
	Tags    []string          `json:"tags"`
	Labels  map[string]string `json:"labels"`
	Owner   *owner            `json:"owner"`
	Active  bool              `json:"active"`
	Ignored string            `json:"-"`
}

type owner struct {
	ID string `json:"id"`
}

type closed struct {
	Reason
}

type Reason struct {
	Why string `json:"why"`
}

// halfCodec encodes a closed event but cannot decode it.
type halfCodec struct{ eventsourcing.Codec[event] }

func (c halfCodec) EventTypes() []string { return []string{"Closed", "Opened"} }

func (c halfCodec) Decode(data eventsourcing.EventData) (event, error) {
	e, err := c.Codec.Decode(data)
	if _, ok := e.(closed); ok {
		return closed{}, nil
	}

	return e, err
}

type plainCodec struct{ eventsourcing.Codec[event] }

func (opened) isEvent() {}
func (closed) isEvent() {}

func newCodec() *eventsourcing.JSONCodec[event] {
	c := eventsourcing.NewJSONCodec[event]()
	eventsourcing.RegisterJSON[event, opened](c, "Opened")
	eventsourcing.RegisterJSON[event, closed](c, "Closed")

	return c
}

func aggregate(codec eventsourcing.Codec[event]) eventsourcing.AggregateType[int, any, event] {
	return eventsourcing.AggregateType[int, any, event]{StreamType: "door", Codec: codec}
}

func fullOpened() opened {
	return opened{
		Name:   "front",
		Note:   "a note",
		At:     time.Date(2026, 11, 23, 4, 5, 6, 0, time.UTC),
		Tags:   []string{"a"},
		Labels: map[string]string{"k": "v"},
		Owner:  &owner{ID: "o-1"},
		Active: true,
	}
}

func fullClosed() closed { return closed{Reason{Why: "night"}} }

func goldenPath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), "testdata", "events.golden.json")
}

func verify(path string, update bool, codec eventsourcing.Codec[event], samples ...event) error {
	golden := eventsourcingtest.NewGolden(path)
	eventsourcingtest.Add(golden, aggregate(codec), samples...)

	return golden.Verify(update)
}

func TestUpdateWritesOneIndentedEntryForEachSampleWithSortedKeys(t *testing.T) {
	path := goldenPath(t)

	require.NoError(t, verify(path, write, newCodec(), fullOpened(), fullClosed()))

	got, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	//nolint:testifylint // JSONEq ignores the indentation and the key order, and the file format includes them.
	assert.Equal(t, `{
  "door/Closed": {
    "schemaVersion": 1,
    "payload": {
      "why": "night"
    }
  },
  "door/Opened": {
    "schemaVersion": 1,
    "payload": {
      "name": "front",
      "note": "a note",
      "at": "2026-11-23T04:05:06Z",
      "tags": [
        "a"
      ],
      "labels": {
        "k": "v"
      },
      "owner": {
        "id": "o-1"
      },
      "active": true
    }
  }
}
`, string(got))
	require.NoError(t, verify(path, compare, newCodec(), fullOpened(), fullClosed()))
}

func TestEntriesOfTwoAggregatesAreInOneFile(t *testing.T) {
	path := goldenPath(t)
	samples := []event{fullOpened(), fullClosed()}
	golden := eventsourcingtest.NewGolden(path)
	eventsourcingtest.Add(golden, aggregate(newCodec()), samples...)
	other := aggregate(newCodec())
	other.StreamType = "window"
	eventsourcingtest.Add(golden, other, samples...)

	require.NoError(t, golden.Verify(write))

	var entries map[string]json.RawMessage

	data, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &entries))
	assert.Len(t, entries, 4)
	assert.Contains(t, entries, "window/Opened")
}

// Verifies: RLS-0082
func TestARegisteredEventTypeWithNoSampleFails(t *testing.T) {
	err := verify(goldenPath(t), write, newCodec(), fullOpened())

	require.Error(t, err)
	assert.Contains(t, err.Error(), `door/Closed`)
	assert.Contains(t, err.Error(), "no sample")
}

func TestTwoSamplesOfOneEventTypeFail(t *testing.T) {
	err := verify(goldenPath(t), write, newCodec(), fullOpened(), fullOpened(), fullClosed())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "door/Opened")
}

func TestACodecThatDoesNotListItsEventTypesFails(t *testing.T) {
	err := verify(goldenPath(t), write, plainCodec{newCodec()}, fullOpened(), fullClosed())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "EventTypes")
}

// Verifies: RLS-0085
func TestAFieldWithTheZeroValueFails(t *testing.T) {
	for name, change := range map[string]func(*opened){
		"name":      func(o *opened) { o.Name = "" },
		"note":      func(o *opened) { o.Note = "" },
		"at":        func(o *opened) { o.At = time.Time{} },
		"tags":      func(o *opened) { o.Tags = nil },
		"tags[0]":   func(o *opened) { o.Tags = []string{""} },
		"labels":    func(o *opened) { o.Labels = map[string]string{} },
		"labels[k]": func(o *opened) { o.Labels = map[string]string{"k": ""} },
		"owner":     func(o *opened) { o.Owner = nil },
		"owner.id":  func(o *opened) { o.Owner = &owner{} },
		"active":    func(o *opened) { o.Active = false },
	} {
		t.Run(name, func(t *testing.T) {
			sample := fullOpened()
			change(&sample)

			err := verify(goldenPath(t), write, newCodec(), sample, fullClosed())

			require.Error(t, err)
			assert.Contains(t, err.Error(), "door/Opened: the field "+name+" has the zero value")
		})
	}
}

// Verifies: RLS-0085
func TestAFieldOfAnEmbeddedStructWithTheZeroValueFails(t *testing.T) {
	err := verify(goldenPath(t), write, newCodec(), fullOpened(), closed{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "door/Closed: the field why has the zero value")
}

func TestUnexportedAndSkippedFieldsCanHaveTheZeroValue(t *testing.T) {
	require.NoError(t, verify(goldenPath(t), write, newCodec(), fullOpened(), fullClosed()))
}

// Verifies: RLS-0086
func TestAnEntryThatTheCodecCannotDecodeToItsSampleFails(t *testing.T) {
	err := verify(goldenPath(t), write, halfCodec{newCodec()}, fullOpened(), fullClosed())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "door/Closed")
	assert.Contains(t, err.Error(), "decode")
}

// Verifies: RLS-0083
func TestAChangedEntryFails(t *testing.T) {
	path := goldenPath(t)
	require.NoError(t, verify(path, write, newCodec(), fullOpened(), fullClosed()))

	changed := fullOpened()
	changed.Name = "back"
	err := verify(path, compare, newCodec(), changed, fullClosed())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "door/Opened")
	assert.NotContains(t, err.Error(), "door/Closed")
	assert.Contains(t, err.Error(), "-update")
}

// Verifies: RLS-0083
func TestARenamedEventTypeFails(t *testing.T) {
	path := goldenPath(t)
	require.NoError(t, verify(path, write, newCodec(), fullOpened(), fullClosed()))

	renamed := eventsourcing.NewJSONCodec[event]()
	eventsourcing.RegisterJSON[event, opened](renamed, "Opened")
	eventsourcing.RegisterJSON[event, closed](renamed, "Shut")
	err := verify(path, compare, renamed, fullOpened(), fullClosed())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "door/Shut")
	assert.Contains(t, err.Error(), "door/Closed")
}

// Verifies: RLS-0087
func TestAnEntryWithNoRegisteredEventTypeFails(t *testing.T) {
	path := goldenPath(t)
	require.NoError(t, verify(path, write, newCodec(), fullOpened(), fullClosed()))

	fewer := eventsourcing.NewJSONCodec[event]()
	eventsourcing.RegisterJSON[event, opened](fewer, "Opened")
	err := verify(path, compare, fewer, fullOpened())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "door/Closed")
}

// Verifies: RLS-0084
func TestWithoutUpdateTheFileDoesNotChange(t *testing.T) {
	path := goldenPath(t)

	err := verify(path, compare, newCodec(), fullOpened(), fullClosed())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "-update")
	assert.NoFileExists(t, path)
}

func TestUpdateDoesNotWriteTheFileWhenASampleIsNotValid(t *testing.T) {
	path := goldenPath(t)

	require.Error(t, verify(path, write, newCodec(), fullOpened()))
	assert.NoFileExists(t, path)
}
