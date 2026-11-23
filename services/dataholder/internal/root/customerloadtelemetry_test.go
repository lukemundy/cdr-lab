package root

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// The customers, accounts and transactions are invented, with values that telemetry has no other reason to contain.
const telemetryLoad = `{"customers": [
	{"firstName": "Quenda", "lastName": "Bettong", "emailAddress": "pademelon@mail.test", "accounts": [
		{"productId": "everyday", "accountNumber": "73519246", "displayName": "Cassowary Saver",
			"creditLimit": "321.00", "transactions": [
				{"type": "PAYMENT", "status": "POSTED", "description": "Wombat Bakery", "amount": "-43.21",
					"reference": "Numbat", "merchantName": "Quoll Mart", "age": "P2DT3H",
					"extendedData": {"payee": "Quoll Mart", "extensionUType": "nppPayload",
						"nppPayload": {"extendedDescription": "Potoroo", "service": "X2P1", "serviceVersion": "03"}}}
			]},
		{"productId": "everyday", "accountNumber": "73519247", "openStatus": "CLOSED"}
	]},
	{"lastName": "Antechinus", "emailAddress": "dunnart@mail.test"}
]}`

var loadValues = []string{
	"quenda", "bettong", "pademelon", "pademelon@mail.test", "73519246", "73519247", "cassowary saver", "321.00",
	"wombat bakery", "43.21", "numbat", "quoll mart", "potoroo", "antechinus", "dunnart", "dunnart@mail.test",
}

// Verifies: DH-0008
func TestACustomerLoadPutsNoCustomerAccountOrTransactionDataInTheSpansMetricsOrLogs(t *testing.T) {
	spans, metrics, logs := recordTelemetry(t)
	l := newLoadLab(t)
	setUpSpans := len(spans.Ended())

	results := l.loaded(t, telemetryLoad)
	l.loaded(t, telemetryLoad)

	// A load with an invalid record, a load with an account number in use, and a load of a product that the brand
	// does not have.
	invalid := strings.Replace(telemetryLoad, `"age": "P2DT3H"`, `"age": "P2W"`, 1)
	invalid = strings.Replace(invalid, "pademelon@", "potoroo@", 1)
	l.load(t, invalid).problemDetail(t, http.StatusBadRequest)

	conflict := strings.ReplaceAll(telemetryLoad, "pademelon@", "bettong@")
	conflict = strings.ReplaceAll(conflict, "dunnart@", "quenda@")
	l.load(t, conflict).problemDetail(t, http.StatusConflict)

	unknown := strings.Replace(strings.ReplaceAll(telemetryLoad, "@mail.test", "@numbat.test"),
		`"productId": "everyday", "accountNumber": "73519247"`,
		`"productId": "quokka-gold", "accountNumber": "73519247"`, 1)
	l.load(t, unknown).problemDetail(t, http.StatusBadRequest)

	values := append([]string{"quokka-gold", "numbat.test"}, loadValues...)
	for _, result := range results {
		values = append(values, result.CustomerNumber)
	}

	routes := 0

	for _, span := range spans.Ended()[setUpSpans:] {
		text := spanText(span)
		for _, value := range values {
			assertNoValue(t, text, value, "span %q", span.Name())
		}

		for _, a := range span.Attributes() {
			if a.Key == semconv.HTTPRouteKey && a.Value.AsString() == "/v1/customer-loads" {
				routes++
			}
		}
	}

	assert.Equal(t, 5, routes, "each load has a server span with the route pattern")
	require.Greater(t, len(spans.Ended()), setUpSpans+routes, "the test sees the database spans of the loads")

	metricsText, logsText := metricText(t, metrics), strings.ToLower(logs.String())
	for _, value := range values {
		assertNoValue(t, metricsText, value)
		assertNoValue(t, logsText, value)
	}
}
