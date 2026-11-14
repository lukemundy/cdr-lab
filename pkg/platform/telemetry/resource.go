package telemetry

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/lukemundy/cdr-lab/pkg/platform/buildinfo"
)

func newResource(ctx context.Context, serviceName string) *resource.Resource {
	build := buildinfo.Get()

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		// Last, so that OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES cannot change the service identity.
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(build.CommitShort),
			semconv.ServiceInstanceID(uuid.NewString()),
			semconv.VCSRefHeadRevision(build.Commit),
			attribute.String("app.build.time", build.Time),
			attribute.String("app.build.id", build.ID),
		),
	)
	if err != nil {
		otel.Handle(err)
	}

	return res
}
