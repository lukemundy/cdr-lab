package register_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/charts/internal/charttest"
)

func TestEachRequiredValueFailsTheRenderWithItsName(t *testing.T) {
	for _, value := range []string{
		"image.tag", "publicBaseURL", "secureBaseURL", "crlURL", "operator.allowedHosts", "telemetry.otlpEndpoint",
		"database.host", "database.name", "database.ownerSecret", "database.serviceSecret",
		"pki.publicTLSSecret", "pki.secureTLSSecret", "pki.intermediateSecret", "pki.caBundleSecret",
		"operator.tokenSecret",
	} {
		_, err := charttest.Render(t, value+"=null")

		require.Error(t, err, value)
		assert.Contains(t, err.Error(), value)
	}
}

func TestTheConsoleNeedsItsAllowedHosts(t *testing.T) {
	_, err := charttest.Render(t, "console.enabled=true")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "console.allowedHosts")
}

func TestOnlyTheMigrationsGetTheOwnerCredentials(t *testing.T) {
	spec := charttest.Pod(t, charttest.MustRender(t))

	require.Len(t, spec.InitContainers, 1)
	migrate := spec.InitContainers[0]
	assert.Equal(t, []string{"migrate"}, migrate.Args)
	assert.Equal(t, "register-owner:username", migrate.Vars()["PGUSER"])
	assert.Equal(t, "register-owner:password", migrate.Vars()["PGPASSWORD"])
	assert.Empty(t, migrate.VolumeMounts, "the migrations need no keys or certificates")

	main := charttest.Main(t, charttest.MustRender(t))
	assert.Equal(t, []string{"serve"}, main.Args)
	assert.Equal(t, "register-service:username", main.Vars()["PGUSER"])
	assert.Equal(t, "register-service:password", main.Vars()["PGPASSWORD"])

	for name, value := range main.Vars() {
		assert.NotContains(t, value, "register-owner", name)
	}
}

func TestOnePodRunsAtATime(t *testing.T) {
	deployment := charttest.Only(t, charttest.MustRender(t), "Deployment")

	require.NotNil(t, deployment.Spec.Replicas)
	assert.Equal(t, 1, *deployment.Spec.Replicas)
	assert.Equal(t, "Recreate", deployment.Spec.Strategy.Type)
}

func TestTheImageIsTheTagAndTheOptionalDigest(t *testing.T) {
	assert.Equal(t, "ghcr.io/lukemundy/cdr-lab/register:0123abc", charttest.Main(t, charttest.MustRender(t)).Image)

	digest := "sha256:" + string(bytes.Repeat([]byte("a"), 64))
	objects := charttest.MustRender(t, "image.digest="+digest)

	assert.Equal(t, "ghcr.io/lukemundy/cdr-lab/register:0123abc@"+digest, charttest.Main(t, objects).Image)
	assert.Equal(t, charttest.Main(t, objects).Image, charttest.Pod(t, objects).InitContainers[0].Image)
}

func TestTheRegisterGetsItsConfiguration(t *testing.T) {
	env := charttest.Main(t, charttest.MustRender(t)).Vars()

	for name, want := range map[string]string{
		"PGHOST":                      "postgres.test",
		"PGPORT":                      "5432",
		"PGDATABASE":                  "register",
		"PGSSLMODE":                   "require",
		"PUBLIC_BASE_URL":             "https://register.test:8081",
		"SECURE_BASE_URL":             "https://register.test:8083",
		"CDR_INTERMEDIATE_CRL_URL":    "http://register.test:8085/crl/cdr-intermediate.crl",
		"OPERATOR_TOKEN":              "register-operator-token:token",
		"OPERATOR_ALLOWED_HOSTS":      "register.test:8082",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://otel.test:4318",
		"CONSOLE_ENABLED":             "false",
	} {
		assert.Equal(t, want, env[name], name)
	}

	for _, name := range []string{
		"PUBLIC_TLS_CERT_FILE", "PUBLIC_TLS_KEY_FILE", "SECURE_TLS_CERT_FILE", "SECURE_TLS_KEY_FILE",
		"CDR_CA_BUNDLE_FILE", "CDR_INTERMEDIATE_CERT_FILE", "CDR_INTERMEDIATE_KEY_FILE",
	} {
		assert.NotEmpty(t, env[name], name)
	}

	assert.NotContains(t, env, "SSL_CERT_FILE", "without a public root, the Register uses the system roots")
}

func TestEachSecretIsMountedReadOnly(t *testing.T) {
	objects := charttest.MustRender(t, "pki.publicRootSecret=public-root")
	spec := charttest.Pod(t, objects)

	var secrets []string
	for _, v := range spec.Volumes {
		secrets = append(secrets, v.Secret.SecretName)
	}

	assert.ElementsMatch(t,
		[]string{"register-public-tls", "register-secure-tls", "cdr-intermediate", "cdr-ca-bundle", "public-root"},
		secrets)

	main := charttest.Main(t, objects)
	require.Len(t, main.VolumeMounts, len(secrets))

	for _, m := range main.VolumeMounts {
		assert.True(t, m.ReadOnly, m.Name)
	}

	assert.NotEmpty(t, main.Vars()["SSL_CERT_FILE"])
}

func TestTheConsoleIsOffByDefault(t *testing.T) {
	objects := charttest.MustRender(t)

	for _, p := range charttest.Only(t, objects, "Service").Spec.Ports {
		assert.NotEqual(t, 8084, p.Port, p.Name)
	}

	for _, p := range charttest.Main(t, objects).Ports {
		assert.NotEqual(t, 8084, p.ContainerPort, p.Name)
	}
}

func TestTheConsoleCanBeTurnedOn(t *testing.T) {
	objects := charttest.MustRender(t, "console.enabled=true", "console.allowedHosts={localhost:8084}")
	main := charttest.Main(t, objects)

	assert.Equal(t, "true", main.Vars()["CONSOLE_ENABLED"])
	assert.Equal(t, "localhost:8084", main.Vars()["CONSOLE_ALLOWED_HOSTS"])
	assert.Contains(t, charttest.Only(t, objects, "Service").Spec.Ports, charttest.Port{Name: "console", Port: 8084})
}

func TestTheServiceGivesEachAPIPort(t *testing.T) {
	assert.Equal(t, map[string]int{"public": 8081, "operator": 8082, "secure": 8083, "crl": 8085},
		charttest.ServicePorts(t, charttest.MustRender(t)))
}

func TestAnotherServiceDoesNotRenameTheServiceOfTheChart(t *testing.T) {
	var names []string

	for _, o := range charttest.MustRender(t, "service.lb.controller=main", "service.lb.ports.public.port=8081") {
		if o.Kind == "Service" {
			names = append(names, o.Metadata.Name)
		}
	}

	assert.Contains(t, names, "test-register")
}

func TestTheProbesUseTheHealthListener(t *testing.T) {
	main := charttest.Main(t, charttest.MustRender(t))

	assert.Equal(t, "/livez", main.LivenessProbe.HTTPGet.Path)
	assert.Equal(t, 8080, main.LivenessProbe.HTTPGet.Port)
	assert.Equal(t, "/readyz", main.ReadinessProbe.HTTPGet.Path)
	assert.Equal(t, 8080, main.ReadinessProbe.HTTPGet.Port)
}

func TestThePodHasTheLeastPrivilege(t *testing.T) {
	spec := charttest.Pod(t, charttest.MustRender(t, "seed.configMap=register-seed"))

	require.NotNil(t, spec.AutomountServiceAccountToken)
	assert.False(t, *spec.AutomountServiceAccountToken)
	assert.True(t, spec.SecurityContext.RunAsNonRoot)
	assert.Equal(t, "RuntimeDefault", spec.SecurityContext.SeccompProfile.Type)

	for _, c := range append(spec.InitContainers, spec.Containers...) {
		require.NotNil(t, c.SecurityContext.AllowPrivilegeEscalation, c.Name)
		assert.False(t, *c.SecurityContext.AllowPrivilegeEscalation, c.Name)
		assert.True(t, c.SecurityContext.ReadOnlyRootFilesystem, c.Name)
		assert.Equal(t, []string{"ALL"}, c.SecurityContext.Capabilities.Drop, c.Name)
	}
}

// Verifies: REG-0913
func TestTheSeedIsOffByDefault(t *testing.T) {
	spec := charttest.Pod(t, charttest.MustRender(t))

	require.Len(t, spec.InitContainers, 1)
	assert.Equal(t, "migrate", spec.InitContainers[0].Name)

	for _, v := range spec.Volumes {
		assert.Empty(t, v.ConfigMap.Name, v.Name)
	}
}

// Verifies: REG-0911, REG-0914, REG-0915, REG-0916, REG-0918
func TestASeedConfigMapLoadsBeforeTheRegisterStarts(t *testing.T) {
	objects := charttest.MustRender(t, "seed.configMap=register-seed")
	spec := charttest.Pod(t, objects)

	require.Len(t, spec.InitContainers, 2)
	assert.Equal(t, "migrate", spec.InitContainers[0].Name)

	seed := spec.InitContainers[1]
	assert.Equal(t, "seed", seed.Name)
	assert.Equal(t, []string{"seed", "-if-empty", "/seed/participants.json"}, seed.Args)
	assert.Equal(t, "register-service:username", seed.Vars()["PGUSER"])
	assert.Equal(t, "register-service:password", seed.Vars()["PGPASSWORD"])
	assert.Equal(t, "register", seed.Vars()["PGDATABASE"])

	for name, value := range seed.Vars() {
		assert.NotContains(t, value, "register-owner", name)
	}

	var volume string

	for _, v := range spec.Volumes {
		if v.ConfigMap.Name == "register-seed" {
			volume = v.Name
		}
	}

	require.NotEmpty(t, volume, "a volume of the ConfigMap")
	assert.Equal(t, []charttest.VolumeMount{{Name: volume, MountPath: "/seed", ReadOnly: true}}, seed.VolumeMounts)

	for _, m := range charttest.Main(t, objects).VolumeMounts {
		assert.NotEqual(t, volume, m.Name, "only the load gets the seed file")
	}
}
