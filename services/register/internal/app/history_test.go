package app_test

import (
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func dataRecipientEventTypes(t *testing.T) []string {
	t.Helper()

	codec, ok := app.DataRecipientType.Codec.(*eventsourcing.JSONCodec[domain.DataRecipientEvent])
	require.True(t, ok)

	return codec.EventTypes()
}

func dataHolderEventTypes(t *testing.T) []string {
	t.Helper()

	codec, ok := app.DataHolderType.Codec.(*eventsourcing.JSONCodec[domain.DataHolderEvent])
	require.True(t, ok)

	return codec.EventTypes()
}

// designTableEventTypes gives the event types of the table of the history in the operator console specification.
func designTableEventTypes(t *testing.T) []string {
	t.Helper()

	design, err := os.ReadFile("../../design/operator-console.md")
	require.NoError(t, err)

	_, table, found := strings.Cut(string(design), "| ID | Event type | Text in the history | Fields | Source |\n")
	require.True(t, found, "the operator console specification has the table of event types")

	table, _, _ = strings.Cut(table, "\n\n")
	code := regexp.MustCompile("`([A-Za-z]+)`")

	var types []string

	for row := range strings.Lines(table) {
		cells := strings.Split(row, "|")
		if len(cells) < 3 || strings.HasPrefix(strings.TrimSpace(cells[2]), "---") {
			continue
		}

		for _, m := range code.FindAllStringSubmatch(cells[2], -1) {
			types = append(types, m[1])
		}
	}

	slices.Sort(types)

	return types
}

// The table can have a row before the codec has its event type, because the design comes first.
//
// Verifies: REG-1018
func TestEachEventTypeOfTheCodecsHasARowInTheTableOfTheDesign(t *testing.T) {
	table := designTableEventTypes(t)
	require.NotEmpty(t, table)

	for _, eventType := range slices.Concat(dataRecipientEventTypes(t), dataHolderEventTypes(t)) {
		assert.Contains(t, table, eventType, "the operator console specification has a row for the event type")
	}
}

var historyTime = time.Date(2026, 10, 23, 1, 2, 3, 456789000, time.UTC)

func recorded(
	t *testing.T,
	codec eventsourcing.Codec[domain.DataRecipientEvent],
	events ...domain.DataRecipientEvent,
) []eventsourcing.RecordedEvent {
	t.Helper()

	stream := make([]eventsourcing.RecordedEvent, 0, len(events))

	for i, e := range events {
		data, err := codec.Encode(e)
		require.NoError(t, err)

		stream = append(stream, eventsourcing.RecordedEvent{
			EventData:     data,
			StreamVersion: int64(i + 1),
			RecordedAt:    historyTime.Add(time.Duration(i) * time.Second),
		})
	}

	return stream
}

func historyRecipient() domain.DataRecipient {
	return domain.DataRecipient{
		LegalEntityID: "dr-1", LegalEntityName: "Recipient", Status: domain.DataRecipientActive,
		Brands: []domain.DataRecipientBrand{{
			DataRecipientBrandID: "b-1",
			BrandName:            "Budget",
			SoftwareProducts: []domain.SoftwareProduct{
				{SoftwareProductID: "sp-1", SoftwareProductName: "Budget app", Status: domain.Active},
			},
		}},
	}
}

// Verifies: REG-1015, REG-1143
func TestTheHistoryOfADataRecipientGivesOnlyTheFieldsOfTheTableForEachEventType(t *testing.T) {
	product := &app.HistorySoftwareProduct{ID: "sp-1", Name: "Budget app"}
	brand := &app.HistoryBrand{ID: "b-1", Name: "Budget"}
	samples := map[string]struct {
		event domain.DataRecipientEvent
		want  app.HistoryEntry
	}{
		"DataRecipientImported": {
			domain.DataRecipientImported{DataRecipient: historyRecipient()},
			app.HistoryEntry{},
		},
		"AccreditationSuspended":   {domain.AccreditationSuspended{}, app.HistoryEntry{}},
		"AccreditationReinstated":  {domain.AccreditationReinstated{}, app.HistoryEntry{}},
		"AccreditationRevoked":     {domain.AccreditationRevoked{}, app.HistoryEntry{}},
		"AccreditationSurrendered": {domain.AccreditationSurrendered{}, app.HistoryEntry{}},
		"SoftwareProductActivated": {
			domain.SoftwareProductActivated{SoftwareProductID: "sp-1"}, app.HistoryEntry{SoftwareProduct: product},
		},
		"SoftwareProductDeactivated": {
			domain.SoftwareProductDeactivated{SoftwareProductID: "sp-1"}, app.HistoryEntry{SoftwareProduct: product},
		},
		"SoftwareProductRemoved": {
			domain.SoftwareProductRemoved{SoftwareProductID: "sp-1"}, app.HistoryEntry{SoftwareProduct: product},
		},
		"ClientCertificateRegistered": {
			domain.ClientCertificateRegistered{SoftwareProductID: "sp-1", Thumbprint: "abc123"},
			app.HistoryEntry{SoftwareProduct: product, Thumbprint: "abc123"},
		},
		"ClientCertificateRemoved": {
			domain.ClientCertificateRemoved{SoftwareProductID: "sp-1", Thumbprint: "abc123"},
			app.HistoryEntry{SoftwareProduct: product, Thumbprint: "abc123"},
		},
		"ClientCertificateIssued": {
			domain.ClientCertificateIssued{SoftwareProductID: "sp-1", Thumbprint: "abc123"},
			app.HistoryEntry{SoftwareProduct: product, Thumbprint: "abc123"},
		},
		"DataRecipientCreated": {
			domain.DataRecipientCreated{DataRecipient: historyRecipient()},
			app.HistoryEntry{},
		},
		"DataRecipientDetailsChanged": {
			domain.DataRecipientDetailsChanged{LegalEntityName: "Renamed"},
			app.HistoryEntry{},
		},
		"DataRecipientBrandAdded": {
			domain.DataRecipientBrandAdded{
				Brand: domain.DataRecipientBrand{DataRecipientBrandID: "b-2", BrandName: "Pay"},
			},
			app.HistoryEntry{Brand: &app.HistoryBrand{ID: "b-2", Name: "Pay"}},
		},
		"DataRecipientBrandDetailsChanged": {
			domain.DataRecipientBrandDetailsChanged{BrandID: "b-1", BrandName: "Money"},
			app.HistoryEntry{Brand: &app.HistoryBrand{ID: "b-1", Name: "Money"}},
		},
		"DataRecipientBrandActivated": {
			domain.DataRecipientBrandActivated{BrandID: "b-1"},
			app.HistoryEntry{Brand: brand},
		},
		"DataRecipientBrandDeactivated": {
			domain.DataRecipientBrandDeactivated{BrandID: "b-1"}, app.HistoryEntry{Brand: brand},
		},
		"DataRecipientBrandRemoved": {domain.DataRecipientBrandRemoved{BrandID: "b-1"}, app.HistoryEntry{Brand: brand}},
		"SoftwareProductAdded": {
			domain.SoftwareProductAdded{
				BrandID:         "b-1",
				SoftwareProduct: domain.SoftwareProduct{SoftwareProductID: "sp-2", SoftwareProductName: "Pay app"},
			},
			app.HistoryEntry{SoftwareProduct: &app.HistorySoftwareProduct{ID: "sp-2", Name: "Pay app"}},
		},
		"SoftwareProductDetailsChanged": {
			domain.SoftwareProductDetailsChanged{SoftwareProductID: "sp-1", SoftwareProductName: "Saver app"},
			app.HistoryEntry{SoftwareProduct: &app.HistorySoftwareProduct{ID: "sp-1", Name: "Saver app"}},
		},
	}

	var sampleTypes []string
	for eventType := range samples {
		sampleTypes = append(sampleTypes, eventType)
	}

	slices.Sort(sampleTypes)
	require.Equal(t, dataRecipientEventTypes(t), sampleTypes, "the test has a sample of each event type")

	for _, eventType := range sampleTypes {
		if eventType == "DataRecipientImported" || eventType == "DataRecipientCreated" {
			continue
		}

		t.Run(eventType, func(t *testing.T) {
			sample := samples[eventType]
			stream := recorded(t, app.DataRecipientType.Codec,
				samples["DataRecipientImported"].event, sample.event)

			entries, err := app.HistoryOfDataRecipient("dr-1", stream)
			require.NoError(t, err)

			want := sample.want
			want.StreamVersion, want.EventType, want.RecordedAt = 2, eventType, stream[1].RecordedAt

			assert.Equal(t, []app.HistoryEntry{
				{StreamVersion: 1, EventType: "DataRecipientImported", RecordedAt: stream[0].RecordedAt},
				want,
			}, entries)
		})
	}
}

func TestTheHistoryOfADataRecipientRefusesAnEventOfAnUnknownSoftwareProduct(t *testing.T) {
	stream := recorded(t, app.DataRecipientType.Codec,
		domain.DataRecipientImported{DataRecipient: historyRecipient()},
		domain.SoftwareProductDeactivated{SoftwareProductID: "sp-unknown"},
	)

	_, err := app.HistoryOfDataRecipient("dr-1", stream)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func historyHolder() domain.DataHolder {
	return domain.DataHolder{
		LegalEntityID: "dh-1", LegalEntityName: "Holder", Status: domain.DataHolderActive,
		Brands: []domain.DataHolderBrand{{DataHolderBrandID: "dhb-1", BrandName: "Holder Bank", Status: domain.Active}},
	}
}

// Verifies: REG-1015, REG-1143
func TestTheHistoryOfADataHolderGivesOnlyTheFieldsOfTheTableForEachEventType(t *testing.T) {
	bank := &app.HistoryBrand{ID: "dhb-1", Name: "Holder Bank"}
	samples := map[string]struct {
		event domain.DataHolderEvent
		want  app.HistoryEntry
	}{
		"DataHolderImported": {domain.DataHolderImported{DataHolder: historyHolder()}, app.HistoryEntry{}},
		"DataHolderCreated":  {domain.DataHolderCreated{DataHolder: historyHolder()}, app.HistoryEntry{}},
		"DataHolderDetailsChanged": {
			domain.DataHolderDetailsChanged{LegalEntityName: "New"},
			app.HistoryEntry{},
		},
		"DataHolderLegalEntityRemoved": {domain.DataHolderLegalEntityRemoved{}, app.HistoryEntry{}},
		"DataHolderBrandAdded": {
			domain.DataHolderBrandAdded{Brand: domain.DataHolderBrand{InterimID: "dhb-2", BrandName: "Holder Energy"}},
			app.HistoryEntry{Brand: &app.HistoryBrand{ID: "dhb-2", Name: "Holder Energy"}},
		},
		"DataHolderBrandDetailsChanged": {
			domain.DataHolderBrandDetailsChanged{BrandID: "dhb-1", BrandName: "Holder Money"},
			app.HistoryEntry{Brand: &app.HistoryBrand{ID: "dhb-1", Name: "Holder Money"}},
		},
		"DataHolderBrandActivated": {
			domain.DataHolderBrandActivated{BrandID: "dhb-1"},
			app.HistoryEntry{Brand: bank},
		},
		"DataHolderBrandDeactivated": {
			domain.DataHolderBrandDeactivated{BrandID: "dhb-1"},
			app.HistoryEntry{Brand: bank},
		},
		"DataHolderBrandRemoved": {domain.DataHolderBrandRemoved{BrandID: "dhb-1"}, app.HistoryEntry{Brand: bank}},
		"ServerCertificateIssued": {
			domain.ServerCertificateIssued{BrandID: "dhb-1", Thumbprint: "abc123"},
			app.HistoryEntry{Brand: bank, Thumbprint: "abc123"},
		},
	}

	sampleTypes := slices.Sorted(maps.Keys(samples))
	require.Equal(t, dataHolderEventTypes(t), sampleTypes, "the test has a sample of each event type")

	for _, eventType := range sampleTypes {
		t.Run(eventType, func(t *testing.T) {
			events := []domain.DataHolderEvent{samples["DataHolderImported"].event, samples[eventType].event}
			if eventType == "DataHolderImported" || eventType == "DataHolderCreated" {
				events = events[1:]
			}

			stream := recordedDataHolderEvents(t, events...)

			entries, err := app.HistoryOfDataHolder("dh-1", stream)
			require.NoError(t, err)

			want := samples[eventType].want
			last := stream[len(stream)-1]
			want.StreamVersion, want.EventType, want.RecordedAt = last.StreamVersion, eventType, last.RecordedAt
			assert.Equal(t, want, entries[len(entries)-1])
		})
	}
}

func recordedDataHolderEvents(t *testing.T, events ...domain.DataHolderEvent) []eventsourcing.RecordedEvent {
	t.Helper()

	stream := make([]eventsourcing.RecordedEvent, 0, len(events))

	for i, e := range events {
		data, err := app.DataHolderType.Codec.Encode(e)
		require.NoError(t, err)

		stream = append(stream, eventsourcing.RecordedEvent{
			EventData:     data,
			StreamVersion: int64(i + 1),
			RecordedAt:    historyTime.Add(time.Duration(i) * time.Second),
		})
	}

	return stream
}

func TestTheHistoryOfADataHolderRefusesAnEventOfAnUnknownBrand(t *testing.T) {
	stream := recordedDataHolderEvents(t,
		domain.DataHolderImported{DataHolder: historyHolder()},
		domain.DataHolderBrandRemoved{BrandID: "dhb-unknown"},
	)

	_, err := app.HistoryOfDataHolder("dh-1", stream)

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// Verifies: REG-1020
func TestTheHistoryOfACreatedDataRecipientGivesNoFields(t *testing.T) {
	stream := recorded(t, app.DataRecipientType.Codec, domain.DataRecipientCreated{DataRecipient: historyRecipient()})

	entries, err := app.HistoryOfDataRecipient("dr-1", stream)

	require.NoError(t, err)
	assert.Equal(t, []app.HistoryEntry{
		{StreamVersion: 1, EventType: "DataRecipientCreated", RecordedAt: stream[0].RecordedAt},
	}, entries)
}
