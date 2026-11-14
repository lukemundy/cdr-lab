package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

func TestNewResource(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "from-env")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.version=from-env,deployment.environment.name=lab")

	res := newResource(t.Context(), "register")

	value := func(key attribute.Key) string {
		v, _ := res.Set().Value(key)
		return v.AsString()
	}

	t.Run("code sets the service identity and environment variables cannot change it", func(t *testing.T) {
		assert.Equal(t, "register", value(semconv.ServiceNameKey))
		assert.Equal(t, "unknown", value(semconv.ServiceVersionKey))
		assert.NotEmpty(t, value(semconv.ServiceInstanceIDKey))
	})

	t.Run("build information", func(t *testing.T) {
		assert.Equal(t, "unknown", value(semconv.VCSRefHeadRevisionKey))
		assert.Equal(t, "unknown", value("app.build.time"))
		assert.Equal(t, "local", value("app.build.id"))
	})

	t.Run("environment variables add other attributes", func(t *testing.T) {
		assert.Equal(t, "lab", value("deployment.environment.name"))
	})

	t.Run("detectors add SDK, host, OS and process attributes", func(t *testing.T) {
		for _, key := range []attribute.Key{
			semconv.TelemetrySDKNameKey,
			semconv.HostNameKey,
			semconv.OSTypeKey,
			semconv.ProcessPIDKey,
			semconv.ProcessRuntimeNameKey,
		} {
			assert.True(t, res.Set().HasValue(key), "resource has no %s", key)
		}
	})
}
