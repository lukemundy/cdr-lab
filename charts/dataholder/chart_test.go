package dataholder_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/charts/internal/charttest"
)

const brandConfigMap = "brand.configMap=bank-brand"

func initContainer(t *testing.T, objects []charttest.Object, name string) charttest.Container {
	t.Helper()

	for _, c := range charttest.Pod(t, objects).InitContainers {
		if c.Name == name {
			return c
		}
	}

	require.FailNow(t, "no init container "+name)

	return charttest.Container{}
}

// Verifies: DH-1342
func TestEachRequiredValueFailsTheRenderWithItsName(t *testing.T) {
	for _, value := range []string{
		"image.tag", "brand.id", "publicBaseURL", "secureBaseURL", "register.publicBaseURL", "register.statusInterval",
		"mail.smtpAddress", "mail.from", "internal.allowedHosts", "operator.tokenSecret",
		"database.host", "database.name", "database.ownerSecret", "database.serviceSecret",
		"pki.publicTLSSecret", "pki.secureTLSSecret", "pki.caBundleSecret",
	} {
		_, err := charttest.Render(t, value+"=null")

		require.Error(t, err, value)
		assert.Contains(t, err.Error(), value)
	}
}

// Verifies: DH-1345, DH-1346
func TestTheValuesGiveExactlyOneBrand(t *testing.T) {
	for name, set := range map[string][]string{
		"no brand":   {"brand.preset=null"},
		"two brands": {brandConfigMap},
	} {
		_, err := charttest.Render(t, set...)

		require.Error(t, err, name)
		assert.Contains(t, err.Error(), "brand.preset", name)
		assert.Contains(t, err.Error(), "brand.configMap", name)
	}
}

func TestAPresetBrandNeedsNoFiles(t *testing.T) {
	objects := charttest.MustRender(t)

	for _, c := range []charttest.Container{initContainer(t, objects, "seed"), charttest.Main(t, objects)} {
		assert.Equal(t, "wattle-bank", c.Vars()["DATA_HOLDER_BRAND"], c.Name)
		assert.NotContains(t, c.Vars(), "DATA_HOLDER_BRAND_DIR", c.Name)
	}

	for _, v := range charttest.Pod(t, objects).Volumes {
		assert.Empty(t, v.ConfigMap.Name, v.Name)
	}
}

// Verifies: DH-1347
func TestABrandConfigMapIsOneDirectory(t *testing.T) {
	objects := charttest.MustRender(t, "brand.preset=null", brandConfigMap)

	var volume string

	for _, v := range charttest.Pod(t, objects).Volumes {
		if v.ConfigMap.Name == "bank-brand" {
			volume = v.Name
		}
	}

	require.NotEmpty(t, volume, "a volume of the ConfigMap")

	for _, c := range []charttest.Container{initContainer(t, objects, "seed"), charttest.Main(t, objects)} {
		assert.NotContains(t, c.Vars(), "DATA_HOLDER_BRAND", c.Name)

		var mounts []charttest.VolumeMount

		for _, m := range c.VolumeMounts {
			if m.Name == volume {
				mounts = append(mounts, m)
			}
		}

		require.Len(t, mounts, 1, c.Name)
		assert.Equal(t, mounts[0].MountPath, c.Vars()["DATA_HOLDER_BRAND_DIR"], c.Name)
		assert.Empty(t, mounts[0].SubPath, "a subPath mount gets no updates")
		assert.True(t, mounts[0].ReadOnly, c.Name)
	}
}

// Verifies: DH-1348, DH-1349
func TestTheMigrationsRunBeforeTheProductLoad(t *testing.T) {
	spec := charttest.Pod(t, charttest.MustRender(t))

	require.Len(t, spec.InitContainers, 2)
	assert.Equal(t, []string{"migrate"}, spec.InitContainers[0].Args)
	assert.Equal(t, []string{"seed", "-if-empty"}, spec.InitContainers[1].Args)
}

func TestOnlyTheMigrationsGetTheOwnerCredentials(t *testing.T) {
	objects := charttest.MustRender(t, "brand.preset=null", brandConfigMap)

	migrate := initContainer(t, objects, "migrate")
	assert.Equal(t, "bank-owner:username", migrate.Vars()["PGUSER"])
	assert.Equal(t, "bank-owner:password", migrate.Vars()["PGPASSWORD"])
	assert.Empty(t, migrate.VolumeMounts, "the migrations need no files")

	seed := initContainer(t, objects, "seed")
	require.Len(t, seed.VolumeMounts, 1, "the product load needs only the brand")

	for _, c := range []charttest.Container{seed, charttest.Main(t, objects)} {
		assert.Equal(t, "bank-service:username", c.Vars()["PGUSER"], c.Name)
		assert.Equal(t, "bank-service:password", c.Vars()["PGPASSWORD"], c.Name)

		for name, value := range c.Vars() {
			assert.NotContains(t, value, "bank-owner", c.Name+" "+name)
		}
	}

	assert.Equal(t, []string{"serve"}, charttest.Main(t, objects).Args)
}

// Verifies: DH-1350
func TestOnePodRunsAtATime(t *testing.T) {
	deployment := charttest.Only(t, charttest.MustRender(t), "Deployment")

	require.NotNil(t, deployment.Spec.Replicas)
	assert.Equal(t, 1, *deployment.Spec.Replicas)
	assert.Equal(t, "Recreate", deployment.Spec.Strategy.Type)
}

func TestTheImageIsTheTagAndTheOptionalDigest(t *testing.T) {
	assert.Equal(t, "ghcr.io/lukemundy/cdr-lab/dataholder:0123abc", charttest.Main(t, charttest.MustRender(t)).Image)

	digest := "sha256:" + string(bytes.Repeat([]byte("a"), 64))
	objects := charttest.MustRender(t, "image.digest="+digest)
	spec := charttest.Pod(t, objects)

	for _, c := range append(spec.InitContainers, spec.Containers...) {
		assert.Equal(t, "ghcr.io/lukemundy/cdr-lab/dataholder:0123abc@"+digest, c.Image, c.Name)
	}
}

func TestTheDataHolderGetsItsConfiguration(t *testing.T) {
	env := charttest.Main(t, charttest.MustRender(t)).Vars()

	for name, want := range map[string]string{
		"PGHOST":                      "postgres.test",
		"PGPORT":                      "5432",
		"PGDATABASE":                  "bank",
		"PGSSLMODE":                   "require",
		"DATA_HOLDER_BRAND_ID":        "00000000-0000-4000-8000-000000000001",
		"PUBLIC_BASE_URL":             "https://bank.test:8081",
		"SECURE_BASE_URL":             "https://bank.test:8083",
		"REGISTER_PUBLIC_BASE_URL":    "https://register.test:8081",
		"REGISTER_JWKS_URL":           "https://register.test:8081/cdr-register/v1/jwks",
		"REGISTER_STATUS_INTERVAL":    "1m",
		"SMTP_ADDRESS":                "mail.test:25",
		"MAIL_FROM":                   "no-reply@bank.test",
		"OPERATOR_TOKEN":              "bank-operator-token:token",
		"INTERNAL_ALLOWED_HOSTS":      "bank.test:8080",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://otel.test:4318",
		"CONSOLE_ENABLED":             "false",
		"BANK_SITE_ENABLED":           "false",
	} {
		assert.Equal(t, want, env[name], name)
	}

	for _, name := range []string{
		"PUBLIC_TLS_CERT_FILE", "PUBLIC_TLS_KEY_FILE", "SECURE_TLS_CERT_FILE", "SECURE_TLS_KEY_FILE",
		"CDR_CA_BUNDLE_FILE",
	} {
		assert.NotEmpty(t, env[name], name)
	}

	assert.NotContains(t, env, "SSL_CERT_FILE", "without a public root, the data holder uses the system roots")
}

// Verifies: DH-0467
func TestTheCRLMaximumAgeIsOptional(t *testing.T) {
	assert.NotContains(t, charttest.Main(t, charttest.MustRender(t)).Vars(), "CRL_MAX_AGE",
		"without a value, the data holder uses its default")

	env := charttest.Main(t, charttest.MustRender(t, "crlMaxAge=30s")).Vars()

	assert.Equal(t, "30s", env["CRL_MAX_AGE"])
}

// Verifies: DH-1351
func TestTheRegisterJWKSURLIsTheCDSPathOnTheRegisterBaseURL(t *testing.T) {
	env := charttest.Main(t, charttest.MustRender(t, "register.publicBaseURL=https://register.test/base/")).Vars()

	assert.Equal(t, "https://register.test/base/cdr-register/v1/jwks", env["REGISTER_JWKS_URL"])
}

// Verifies: DH-1353
func TestTheChartHasNoFixedOTPs(t *testing.T) {
	spec := charttest.Pod(
		t,
		charttest.MustRender(
			t,
			"console.enabled=true",
			"bankSite.enabled=true",
			"bankSite.allowedHosts={localhost:8085}",
		),
	)

	for _, c := range append(spec.InitContainers, spec.Containers...) {
		assert.NotContains(t, c.Vars(), "FIXED_OTPS", c.Name)
	}
}

func TestEachSecretIsMountedReadOnly(t *testing.T) {
	objects := charttest.MustRender(t, "pki.publicRootSecret=public-root")

	var secrets []string

	for _, v := range charttest.Pod(t, objects).Volumes {
		secrets = append(secrets, v.Secret.SecretName)
	}

	assert.ElementsMatch(t, []string{"bank-public-tls", "bank-secure-tls", "cdr-ca-bundle", "public-root"}, secrets)

	main := charttest.Main(t, objects)
	require.Len(t, main.VolumeMounts, len(secrets))

	for _, m := range main.VolumeMounts {
		assert.True(t, m.ReadOnly, m.Name)
	}

	assert.NotEmpty(t, main.Vars()["SSL_CERT_FILE"])
}

// Verifies: DH-1354
func TestTheLabFeaturesAreOffByDefault(t *testing.T) {
	objects := charttest.MustRender(t)
	main := charttest.Main(t, objects)

	assert.Equal(t, "false", main.Vars()["CONSOLE_ENABLED"])
	assert.Equal(t, "false", main.Vars()["BANK_SITE_ENABLED"])
	assert.NotContains(t, main.Vars(), "BANK_SITE_ALLOWED_HOSTS")
	assert.NotContains(t, charttest.ServicePorts(t, objects), "bank-site")
}

// Verifies: DH-2626
func TestTheContainerAndTheServiceGiveOnePortForEachListener(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values []string
		want   map[string]int
	}{
		{name: "default", want: map[string]int{"internal": 8080, "public": 8081, "secure": 8083}},
		{
			name:   "with the online bank site",
			values: []string{"bankSite.enabled=true", "bankSite.allowedHosts={localhost:8085}"},
			want:   map[string]int{"internal": 8080, "public": 8081, "secure": 8083, "bank-site": 8085},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			objects := charttest.MustRender(t, tc.values...)

			assert.Equal(t, tc.want, charttest.ServicePorts(t, objects), "Service ports")

			got := map[string]int{}
			for _, p := range charttest.Main(t, objects).Ports {
				got[p.Name] = p.ContainerPort
			}

			assert.Equal(t, tc.want, got, "container ports")
		})
	}
}

func TestTheBankSiteNeedsItsAllowedHosts(t *testing.T) {
	_, err := charttest.Render(t, "bankSite.enabled=true")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bankSite.allowedHosts")
}

func TestTheBankSiteCanBeTurnedOn(t *testing.T) {
	main := charttest.Main(
		t,
		charttest.MustRender(t, "bankSite.enabled=true", "bankSite.allowedHosts={localhost:9999}"),
	)

	assert.Equal(t, "true", main.Vars()["BANK_SITE_ENABLED"])
	assert.Equal(t, "localhost:9999", main.Vars()["BANK_SITE_ALLOWED_HOSTS"])
}

// Verifies: DH-2624, DH-2627
func TestTheConsoleCanBeTurnedOnWithNoPortOrAllowlistOfItsOwn(t *testing.T) {
	objects := charttest.MustRender(t, "console.enabled=true")
	main := charttest.Main(t, objects)

	assert.Equal(t, "true", main.Vars()["CONSOLE_ENABLED"])
	assert.Equal(t, []string{"INTERNAL_ALLOWED_HOSTS"}, main.AllowlistVars())
	assert.Equal(t, charttest.ServicePorts(t, charttest.MustRender(t)), charttest.ServicePorts(t, objects))
}

func TestAnotherServiceDoesNotRenameTheServiceOfTheChart(t *testing.T) {
	var names []string

	for _, o := range charttest.MustRender(t, "service.lb.controller=main", "service.lb.ports.public.port=8081") {
		if o.Kind == "Service" {
			names = append(names, o.Metadata.Name)
		}
	}

	assert.Contains(t, names, "test-dataholder")
}

func TestTheProbesUseTheInternalListener(t *testing.T) {
	main := charttest.Main(t, charttest.MustRender(t))

	assert.Equal(t, "/livez", main.LivenessProbe.HTTPGet.Path)
	assert.Equal(t, 8080, main.LivenessProbe.HTTPGet.Port)
	assert.Equal(t, "/readyz", main.ReadinessProbe.HTTPGet.Path)
	assert.Equal(t, 8080, main.ReadinessProbe.HTTPGet.Port)
}

func TestThePodHasTheLeastPrivilege(t *testing.T) {
	spec := charttest.Pod(t, charttest.MustRender(t))

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

func TestWithNoOTLPEndpointNoContainerGetsATelemetryVariable(t *testing.T) {
	spec := charttest.Pod(t, charttest.MustRender(t, "telemetry.otlpEndpoint=null"))

	for _, c := range append(spec.InitContainers, spec.Containers...) {
		for name := range c.Vars() {
			assert.NotContains(t, name, "OTEL_", c.Name)
		}
	}
}
