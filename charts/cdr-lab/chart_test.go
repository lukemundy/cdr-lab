package cdrlab_test

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/charts/internal/charttest"
)

var presetBrands = []string{"harbour-bank", "kestrel-bank", "redgum-bank", "wattle-bank"}

type seedFile struct {
	DataHolders []struct {
		LegalEntityID    string `json:"legalEntityId"`
		LegalEntityName  string `json:"legalEntityName"`
		LogoURI          string `json:"logoUri"`
		DataHolderBrands []struct {
			DataHolderBrandID string `json:"dataHolderBrandId"`
			BrandName         string `json:"brandName"`
			LogoURI           string `json:"logoUri"`
			EndpointDetail    struct {
				PublicBaseURI   string `json:"publicBaseUri"`
				ProductBaseURI  string `json:"productBaseUri"`
				ResourceBaseURI string `json:"resourceBaseUri"`
				InfosecBaseURI  string `json:"infosecBaseUri"`
				WebsiteURI      string `json:"websiteUri"`
			} `json:"endpointDetail"`
			JWKSEndpoint string `json:"jwksEndpoint"`
		} `json:"dataHolderBrands"`
	} `json:"dataHolders"`
	DataRecipients []struct {
		LegalEntityID string `json:"legalEntityId"`
	} `json:"dataRecipients"`
}

func seed(t *testing.T, objects []object) seedFile {
	t.Helper()

	configMaps := ofKind(objects, "ConfigMap")
	require.Len(t, configMaps, 1)

	var s seedFile
	require.NoError(t, json.Unmarshal([]byte(str(configMaps[0], "data", "participants.json")), &s))

	return s
}

func brandMain(t *testing.T, objects []object, brand string) map[string]string {
	t.Helper()

	return env(container(t, named(t, objects, "Deployment", "test-cdr-lab-"+brand), "main"))
}

func registerMain(t *testing.T, objects []object) map[string]string {
	t.Helper()

	return env(container(t, named(t, objects, "Deployment", "test-cdr-lab-register"), "main"))
}

// Verifies: LAB-0001, LAB-0003
func TestOneReleaseRunsTheRegisterAndEachBrandWithTheServiceNames(t *testing.T) {
	objects := renderLint(t)

	var names []string
	for _, s := range ofKind(objects, "Service") {
		names = append(names, str(s, "metadata", "name"))
	}

	assert.ElementsMatch(t, append([]string{"register"}, presetBrands...), names)
	assert.Len(t, ofKind(objects, "Deployment"), 5)
}

// Verifies: LAB-0002
func TestEachResourceIsInTheNamespaceOfTheRelease(t *testing.T) {
	for _, o := range renderLint(t, "--namespace", "lab-one") {
		assert.Equal(t, "lab-one", str(o, "metadata", "namespace"), str(o, "kind")+" "+str(o, "metadata", "name"))
	}
}

// Verifies: LAB-0005
func TestEachRequiredValueFailsTheRenderWithItsName(t *testing.T) {
	for _, value := range []string{
		"images.register.tag", "images.dataholder.tag", "images.cdrlab.tag", "operator.tokenSecret", "database.host",
		"register.database.name", "register.database.ownerSecret", "register.database.serviceSecret",
		"dataHolders.registerStatusInterval",
		"brands.wattle-bank.database.name", "brands.wattle-bank.database.ownerSecret",
		"brands.wattle-bank.database.serviceSecret", "brands.wattle-bank.mail.smtpAddress",
		"brands.wattle-bank.mail.from",
	} {
		err := renderErr(t, set(value+"=")...)

		assert.Contains(t, err.Error(), value)
	}
}

// Verifies: LAB-0006
func TestTheLabOnlyUIsAreOffByDefault(t *testing.T) {
	objects := render(t)

	assert.Equal(t, "false", registerMain(t, objects)["CONSOLE_ENABLED"])
	assert.Equal(t, "false", brandMain(t, objects, "wattle-bank")["CONSOLE_ENABLED"])
	assert.Equal(t, "false", brandMain(t, objects, "wattle-bank")["BANK_SITE_ENABLED"])

	objects = renderLint(t)
	assert.Equal(t, "true", registerMain(t, objects)["CONSOLE_ENABLED"])
	assert.Equal(t, "true", brandMain(t, objects, "wattle-bank")["CONSOLE_ENABLED"])
	assert.Equal(t, "true", brandMain(t, objects, "wattle-bank")["BANK_SITE_ENABLED"])
}

// Verifies: LAB-0007, LAB-0008
func TestEachServiceUsesItsDatabaseAndTheOperatorToken(t *testing.T) {
	objects := render(t)
	register := named(t, objects, "Deployment", "test-cdr-lab-register")
	brand := named(t, objects, "Deployment", "test-cdr-lab-wattle-bank")

	for _, c := range []struct {
		workload       object
		container, db  string
		user, password string
	}{
		{register, "migrate", "register", "register-owner:username", "register-owner:password"},
		{register, "seed", "register", "register-service:username", "register-service:password"},
		{register, "main", "register", "register-service:username", "register-service:password"},
		{brand, "migrate", "wattle_bank", "wattle-bank-owner:username", "wattle-bank-owner:password"},
		{brand, "seed", "wattle_bank", "wattle-bank-service:username", "wattle-bank-service:password"},
		{brand, "main", "wattle_bank", "wattle-bank-service:username", "wattle-bank-service:password"},
	} {
		vars := env(container(t, c.workload, c.container))
		assert.Equal(t, "postgres.test", vars["PGHOST"], c.container)
		assert.Equal(t, c.db, vars["PGDATABASE"], c.container)
		assert.Equal(t, c.user, vars["PGUSER"], c.container)
		assert.Equal(t, c.password, vars["PGPASSWORD"], c.container)
	}

	assert.Equal(t, "lab-operator-token:token", registerMain(t, objects)["OPERATOR_TOKEN"])
	assert.Equal(t, "lab-operator-token:token", brandMain(t, objects, "wattle-bank")["OPERATOR_TOKEN"])

	for _, o := range ofKind(objects, "Job") {
		if strings.Contains(str(o, "metadata", "name"), "-cert-") {
			assert.Equal(t, "lab-operator-token:token", env(container(t, o, "main"))["REGISTER_OPERATOR_TOKEN"])
		}
	}
}

// Verifies: LAB-0010, LAB-0011
func TestABrandKeyMustBeAPresetName(t *testing.T) {
	err := renderErr(t, set("brands.gum-bank.database.name=x")...)

	assert.Contains(t, err.Error(), "brands.gum-bank")

	for _, name := range presetBrands {
		assert.Contains(t, err.Error(), name)
	}
}

// Verifies: LAB-0012, LAB-0013
func TestThePresetRecordsAreTheExampleSeedOfTheRegister(t *testing.T) {
	chart, err := os.ReadFile("files/preset-brands.json")
	require.NoError(t, err)

	example, err := os.ReadFile("../../services/register/seed/preset-banks.json")
	require.NoError(t, err)

	assert.Equal(t, string(example), string(chart))
}

// Verifies: LAB-0014
func TestTheValuesOfABrandHaveNoFieldOfItsRecord(t *testing.T) {
	for _, value := range []string{
		"brands.wattle-bank.id=x", "brands.wattle-bank.dataHolderBrandId=x", "brands.wattle-bank.brandName=x",
		"brands.wattle-bank.legalEntityId=x", "brands.wattle-bank.status=INACTIVE",
		"brands.wattle-bank.publicBaseURL=x",
		"brands.wattle-bank.endpointDetail.publicBaseUri=x", "brands.wattle-bank.brand.id=x",
		"brands.wattle-bank.public.uri=x",
	} {
		err := renderErr(t, set(value)...)

		assert.Contains(t, err.Error(), "additional properties", value)
	}
}

// Verifies: LAB-0015, LAB-0016, LAB-0017, LAB-0018
func TestTheSeedHasTheRecordOfEachBrandWithTheSettingsOfItsDataHolder(t *testing.T) {
	for _, c := range []struct {
		desc   string
		args   []string
		brands []string
	}{
		{desc: "one brand with no hosts", args: nil, brands: []string{"wattle-bank"}},
		{desc: "four brands with hosts", args: lintArgs, brands: presetBrands},
	} {
		objects := render(t, c.args...)
		s := seed(t, objects)

		var names []string

		for _, holder := range s.DataHolders {
			require.Len(t, holder.DataHolderBrands, 1)
			brand := holder.DataHolderBrands[0]
			name := strings.ToLower(strings.ReplaceAll(brand.BrandName, " ", "-"))
			names = append(names, name)

			vars := brandMain(t, objects, name)
			public, secure := vars["PUBLIC_BASE_URL"], vars["SECURE_BASE_URL"]

			assert.Equal(t, brand.DataHolderBrandID, vars["DATA_HOLDER_BRAND_ID"], name)
			assert.Equal(t, public, brand.EndpointDetail.PublicBaseURI, name)
			assert.Equal(t, public, brand.EndpointDetail.ProductBaseURI, name)
			assert.Equal(t, public, brand.EndpointDetail.InfosecBaseURI, name)
			assert.Equal(t, secure, brand.EndpointDetail.ResourceBaseURI, name)
			assert.Equal(t, public+"/brand/logo.svg", brand.LogoURI, name)
			assert.Equal(t, public+"/brand/logo.svg", holder.LogoURI, name)
			assert.Equal(t, public+"/jwks", brand.JWKSEndpoint, name)
			assert.Equal(t, "https://"+strings.ReplaceAll(name, "-", "")+".test", brand.EndpointDetail.WebsiteURI, name)
		}

		assert.ElementsMatch(t, c.brands, names, c.desc)
	}
}

// Verifies: LAB-0019
func TestTheValuesCanAddParticipants(t *testing.T) {
	s := seed(t, renderLint(t))

	require.Len(t, s.DataRecipients, 1)
	assert.Equal(t, "00000000-0000-4000-8000-0000000000a1", s.DataRecipients[0].LegalEntityID)
}

// Verifies: LAB-0020
func TestAParticipantWithAPresetIdentifierFails(t *testing.T) {
	for id, values := range map[string][]string{
		"dd4d6bc6-22f4-4f61-aae9-9580382718c8": {
			"participants.dataRecipients[0].legalEntityId=dd4d6bc6-22f4-4f61-aae9-9580382718c8",
		},
		"4967dd80-2cd2-43c3-8b13-f6faafb684ad": {
			"participants.dataHolders[0].legalEntityId=x",
			"participants.dataHolders[0].dataHolderBrands[0].dataHolderBrandId=4967dd80-2cd2-43c3-8b13-f6faafb684ad",
		},
	} {
		err := renderErr(t, set(values...)...)

		assert.Contains(t, err.Error(), id)
	}
}

// Verifies: LAB-0030, LAB-0031, LAB-0032, LAB-0033
func TestTheBaseURIsComeFromTheHostsOrTheServiceNames(t *testing.T) {
	objects := render(t)
	assert.Equal(t, "https://register:8081", registerMain(t, objects)["PUBLIC_BASE_URL"])
	assert.Equal(t, "https://register:8083", registerMain(t, objects)["SECURE_BASE_URL"])
	assert.Equal(t, "https://wattle-bank:8081", brandMain(t, objects, "wattle-bank")["PUBLIC_BASE_URL"])
	assert.Equal(t, "https://wattle-bank:8083", brandMain(t, objects, "wattle-bank")["SECURE_BASE_URL"])

	objects = renderLint(t)
	assert.Equal(t, "https://register.lab.test", registerMain(t, objects)["PUBLIC_BASE_URL"])
	assert.Equal(t, "https://mtls.register.lab.test", registerMain(t, objects)["SECURE_BASE_URL"])
	assert.Equal(t, "https://wattle.lab.test", brandMain(t, objects, "wattle-bank")["PUBLIC_BASE_URL"])
	assert.Equal(t, "https://mtls.wattle.lab.test:8443", brandMain(t, objects, "wattle-bank")["SECURE_BASE_URL"])

	objects = renderLint(t, set("brands.wattle-bank.secure.port=443")...)
	assert.Equal(t, "https://mtls.wattle.lab.test", brandMain(t, objects, "wattle-bank")["SECURE_BASE_URL"])
}

// Verifies: LAB-0034
func TestEachBrandCallsTheRegisterAtItsServiceName(t *testing.T) {
	for _, brand := range presetBrands {
		vars := brandMain(t, renderLint(t), brand)

		assert.Equal(t, "https://register:8081", vars["REGISTER_PUBLIC_BASE_URL"], brand)
		assert.Equal(t, "https://register:8081/cdr-register/v1/jwks", vars["REGISTER_JWKS_URL"], brand)
	}
}

// Verifies: LAB-0035
func TestTheHostAllowlistsHaveTheServiceNameAndTheValues(t *testing.T) {
	assert.Equal(t, "register:8080", registerMain(t, render(t))["INTERNAL_ALLOWED_HOSTS"])
	assert.Equal(t, "wattle-bank:8080", brandMain(t, render(t), "wattle-bank")["INTERNAL_ALLOWED_HOSTS"])

	objects := renderLint(t)
	assert.Equal(t, "register:8080,localhost:8080", registerMain(t, objects)["INTERNAL_ALLOWED_HOSTS"])
	assert.Equal(t, "wattle-bank:8080,localhost:8180", brandMain(t, objects, "wattle-bank")["INTERNAL_ALLOWED_HOSTS"])
	assert.Equal(t, "online.wattle.lab.test,wattle-bank:8085",
		brandMain(t, objects, "wattle-bank")["BANK_SITE_ALLOWED_HOSTS"])
}

// Verifies: LAB-0036
func TestTheCRLURLIsOnThePublicHostOfTheRegister(t *testing.T) {
	assert.Equal(t, "http://register:8080/crl/cdr-intermediate.crl",
		registerMain(t, render(t))["CDR_INTERMEDIATE_CRL_URL"])
	assert.Equal(t, "http://register.lab.test/crl/cdr-intermediate.crl",
		registerMain(t, renderLint(t))["CDR_INTERMEDIATE_CRL_URL"])
}

// Verifies: LAB-0040, LAB-0041, LAB-0044
func TestCertManagerIssuesEachPublicCertificateOnly(t *testing.T) {
	objects := renderLint(t)
	want := map[string][]string{
		"register-public-tls":     {"register.lab.test", "register"},
		"wattle-bank-public-tls":  {"wattle.lab.test", "wattle-bank", "online.wattle.lab.test"},
		"redgum-bank-public-tls":  {"redgum.lab.test", "redgum-bank"},
		"harbour-bank-public-tls": {"harbour.lab.test", "harbour-bank"},
		"kestrel-bank-public-tls": {"kestrel.lab.test", "kestrel-bank"},
	}

	got := map[string][]string{}
	for _, c := range ofKind(objects, "Certificate") {
		got[str(c, "spec", "secretName")] = strs(c, "spec", "dnsNames")
		assert.Equal(t, "lab-ca", str(c, "spec", "issuerRef", "name"))
	}

	assert.Equal(t, want, got)
}

// Verifies: LAB-0042
func TestWithNoIssuerTheServicesMountTheExistingPublicSecrets(t *testing.T) {
	objects := render(t)

	assert.Empty(t, ofKind(objects, "Certificate"))
	assert.Contains(t, mountedSecrets(named(t, objects, "Deployment", "test-cdr-lab-register")), "register-public-tls")
	assert.Contains(t, mountedSecrets(named(t, objects, "Deployment", "test-cdr-lab-wattle-bank")),
		"wattle-bank-public-tls")
}

// Verifies: LAB-0043
func TestAPublicRootReplacesTheSystemRoots(t *testing.T) {
	assert.NotContains(t, registerMain(t, render(t)), "SSL_CERT_FILE")

	objects := renderLint(t)
	for _, name := range []string{"register", "wattle-bank"} {
		d := named(t, objects, "Deployment", "test-cdr-lab-"+name)

		assert.Equal(t, "/pki/public-root/ca.crt", env(container(t, d, "main"))["SSL_CERT_FILE"], name)
		assert.Contains(t, mountedSecrets(d), "public-root", name)
	}
}

// Verifies: LAB-0050, LAB-0051, LAB-0052
func TestAJobOfEachRevisionMakesTheCAThatTheRegisterMounts(t *testing.T) {
	objects := render(t)
	job := named(t, objects, "Job", "test-cdr-lab-ca-r1")

	assert.Equal(t, []string{
		"ca", "init", "-intermediate-secret", "cdr-ca-intermediate", "-ca-bundle-secret", "cdr-ca-bundle",
	}, strs(container(t, job, "main"), "args"))
	assert.Equal(t, "ghcr.io/lukemundy/cdr-lab/cdrlab:89abcde", str(container(t, job, "main"), "image"))
	assert.Equal(t, []string{"cdr-ca-bundle", "cdr-ca-intermediate", "register-public-tls"},
		mountedSecrets(named(t, objects, "Deployment", "test-cdr-lab-register")))
}

// Verifies: LAB-0060
func TestTheRegisterIssuesItsSecureCertificateForTheHostAndTheServiceName(t *testing.T) {
	assert.Equal(t, "register", registerMain(t, render(t))["SECURE_TLS_DNS_NAMES"])
	assert.Equal(t, "mtls.register.lab.test,register", registerMain(t, renderLint(t))["SECURE_TLS_DNS_NAMES"])
}

// Verifies: LAB-0061, LAB-0062, LAB-0063
func TestAJobAndADailyCronJobGetTheServerCertificateOfEachBrand(t *testing.T) {
	objects := renderLint(t)
	want := []string{
		"cert", "issue", "-register-url", "http://register:8080", "-kind", "server",
		"-legal-entity-id", "dd4d6bc6-22f4-4f61-aae9-9580382718c8", "-brand-id", "d7385786-d8b1-4d4b-bb80-b21616df1a4d",
		"-cn", "mtls.wattle.lab.test", "-o", "Wattle Bank Limited",
		"-dns", "mtls.wattle.lab.test", "-dns", "wattle-bank", "-secret", "wattle-bank-secure-tls",
	}

	job := named(t, objects, "Job", "test-cdr-lab-wattle-bank-cert-r1")
	cron := named(t, objects, "CronJob", "test-cdr-lab-wattle-bank-renew")

	assert.Equal(t, want, strs(container(t, job, "main"), "args"))
	assert.Equal(t, want, strs(container(t, cron, "main"), "args"))
	assert.Equal(t, "0 3 * * *", str(cron, "spec", "schedule"))
	assert.Len(t, ofKind(objects, "CronJob"), 4)

	args := strs(container(t, named(t, render(t), "Job", "test-cdr-lab-wattle-bank-cert-r1"), "main"), "args")
	require.Len(t, args, 18)
	assert.Equal(t, []string{"-cn", "wattle-bank", "-o"}, args[10:13])
	assert.Equal(t, []string{"-dns", "wattle-bank", "-secret"}, args[14:17])
}

// Verifies: LAB-0064
func TestTheCertificateJobsTryAgainForTenMinutes(t *testing.T) {
	objects := render(t)

	for _, job := range ofKind(objects, "Job") {
		assert.Equal(t, "600", str(job, "spec", "activeDeadlineSeconds"), str(job, "metadata", "name"))
	}

	job := named(t, objects, "Job", "test-cdr-lab-wattle-bank-cert-r1")
	assert.Equal(t, "20", str(job, "spec", "backoffLimit"))

	cron := named(t, objects, "CronJob", "test-cdr-lab-wattle-bank-renew")
	assert.Equal(t, "600", str(cron, "spec", "jobTemplate", "spec", "activeDeadlineSeconds"))
}

// Verifies: LAB-0065
func TestABrandMountsItsServerCertificate(t *testing.T) {
	d := named(t, render(t), "Deployment", "test-cdr-lab-wattle-bank")

	assert.Equal(t, []string{"cdr-ca-bundle", "wattle-bank-public-tls", "wattle-bank-secure-tls"}, mountedSecrets(d))
	assert.Equal(t, "/tls/secure/tls.crt", env(container(t, d, "main"))["SECURE_TLS_CERT_FILE"])
}

// Verifies: LAB-0066
func TestNoServiceMountsASecretWithASubPath(t *testing.T) {
	for _, w := range workloads(renderLint(t)) {
		spec := podSpec(w)
		for _, c := range append(list(spec, "initContainers"), list(spec, "containers")...) {
			for _, m := range list(c, "volumeMounts") {
				assert.Empty(t, str(m, "subPath"), str(w, "metadata", "name"))
			}
		}
	}
}

// Verifies: LAB-0070
func TestTheRoutesAreOffByDefault(t *testing.T) {
	objects := render(t)

	assert.Empty(t, ofKind(objects, "TLSRoute"))
	assert.Empty(t, ofKind(objects, "HTTPRoute"))

	for _, s := range ofKind(objects, "Service") {
		assert.Equal(t, "ClusterIP", str(s, "spec", "type"), str(s, "metadata", "name"))
	}
}

// Verifies: LAB-0071, LAB-0073, LAB-0074
func TestTheRoutesCoverEachTLSListenerAndTheCRLPath(t *testing.T) {
	objects := renderLint(t)

	routes := map[string]string{}

	for _, r := range ofKind(objects, "TLSRoute") {
		assert.Equal(t, "gateway.networking.k8s.io/v1", r["apiVersion"])
		assert.Equal(t, "lab", str(r, "spec", "parentRefs", 0, "name"))
		assert.Equal(t, "tls", str(r, "spec", "parentRefs", 0, "sectionName"))
		routes[str(r, "spec", "hostnames", 0)] = str(r, "spec", "rules", 0, "backendRefs", 0, "name") + ":" +
			str(r, "spec", "rules", 0, "backendRefs", 0, "port")
	}

	assert.Equal(t, map[string]string{
		"register.lab.test":      "register:8081",
		"mtls.register.lab.test": "register:8083",
		"wattle.lab.test":        "wattle-bank:8081",
		"mtls.wattle.lab.test":   "wattle-bank:8083",
		"online.wattle.lab.test": "wattle-bank:8085",
		"redgum.lab.test":        "redgum-bank:8081",
		"mtls.redgum.lab.test":   "redgum-bank:8083",
		"harbour.lab.test":       "harbour-bank:8081",
		"mtls.harbour.lab.test":  "harbour-bank:8083",
		"kestrel.lab.test":       "kestrel-bank:8081",
		"mtls.kestrel.lab.test":  "kestrel-bank:8083",
	}, routes)

	crl := ofKind(objects, "HTTPRoute")
	require.Len(t, crl, 1)
	assert.Equal(t, []string{"register.lab.test"}, strs(crl[0], "spec", "hostnames"))
	assert.Equal(t, "http", str(crl[0], "spec", "parentRefs", 0, "sectionName"))
	assert.Equal(t, "Exact", str(crl[0], "spec", "rules", 0, "matches", 0, "path", "type"))
	assert.Equal(t, "/crl/cdr-intermediate.crl", str(crl[0], "spec", "rules", 0, "matches", 0, "path", "value"))
	assert.Equal(t, "register:8080", str(crl[0], "spec", "rules", 0, "backendRefs", 0, "name")+":"+
		str(crl[0], "spec", "rules", 0, "backendRefs", 0, "port"))
}

// Verifies: LAB-0072
func TestWithRoutesEachTLSListenerNeedsAHost(t *testing.T) {
	for _, listener := range []string{
		"register.public", "register.secure", "brands.wattle-bank.public", "brands.wattle-bank.secure",
		"brands.wattle-bank.bankSite", "brands.kestrel-bank.secure",
	} {
		err := renderErr(t, append(slices.Clone(lintArgs), set(listener+".host=")...)...)

		assert.Contains(t, err.Error(), listener+".host")
	}
}

// Verifies: LAB-0075
func TestNoRouteGoesToAnInternalListenerExceptTheCRLPath(t *testing.T) {
	objects := renderLint(t)

	for _, r := range ofKind(objects, "TLSRoute") {
		assert.NotEqual(t, "8080", str(r, "spec", "rules", 0, "backendRefs", 0, "port"))
	}

	for _, r := range ofKind(objects, "HTTPRoute") {
		assert.Len(t, list(r, "spec", "rules"), 1)
		assert.Len(t, list(r, "spec", "rules", 0, "matches"), 1)
	}

	for _, s := range ofKind(objects, "Service") {
		assert.Equal(t, "ClusterIP", str(s, "spec", "type"), str(s, "metadata", "name"))
	}

	assert.Empty(t, ofKind(objects, "Ingress"))
	assert.Empty(t, ofKind(objects, "Gateway"))
}

// Verifies: LAB-0085, LAB-0087
func TestTheDefaultTagOfEachImageIsTheAppVersion(t *testing.T) {
	objects, err := charttest.RenderPackaged(t, "v1.2.3",
		"images.register.tag=null", "images.dataholder.tag=null", "images.cdrlab.tag=null")
	require.NoError(t, err)

	images := map[string]bool{}

	for _, o := range objects {
		for _, c := range append(o.Spec.Template.Spec.InitContainers, o.Spec.Template.Spec.Containers...) {
			images[c.Image] = true
		}
	}

	assert.Equal(t, map[string]bool{
		"ghcr.io/lukemundy/cdr-lab/register:v1.2.3":   true,
		"ghcr.io/lukemundy/cdr-lab/dataholder:v1.2.3": true,
		"ghcr.io/lukemundy/cdr-lab/cdrlab:v1.2.3":     true,
	}, images)

	assert.Equal(t, "ghcr.io/lukemundy/cdr-lab/register:0123abc",
		str(container(t, named(t, render(t), "Deployment", "test-cdr-lab-register"), "main"), "image"))
}

// Verifies: LAB-0123
func TestOnlyTheRegisterMountsTheIntermediateKey(t *testing.T) {
	for _, w := range workloads(renderLint(t)) {
		name := str(w, "metadata", "name")
		if name == "test-cdr-lab-register" {
			assert.Contains(t, mountedSecrets(w), "cdr-ca-intermediate")
		} else {
			assert.NotContains(t, mountedSecrets(w), "cdr-ca-intermediate", name)
		}
	}
}

// Verifies: LAB-0124
func TestTheChartMakesNoSecret(t *testing.T) {
	assert.Empty(t, ofKind(renderLint(t), "Secret"))
}

// Verifies: LAB-0125
func TestEachJobCanChangeOnlyItsOwnSecrets(t *testing.T) {
	objects := renderLint(t)
	want := map[string][]string{
		"test-cdr-lab-ca": {"cdr-ca-intermediate", "cdr-ca-bundle"},
	}

	for _, brand := range presetBrands {
		want["test-cdr-lab-"+brand+"-cert"] = []string{brand + "-secure-tls"}
	}

	roles := ofKind(objects, "Role")
	require.Len(t, roles, len(want))

	for _, role := range roles {
		name := str(role, "metadata", "name")
		rules := list(role, "rules")

		require.Len(t, rules, 2, name)
		assert.Equal(t, []string{"secrets"}, strs(rules[0], "resources"), name)
		assert.Equal(t, want[name], strs(rules[0], "resourceNames"), name)
		assert.Equal(t, []string{"get", "update", "patch", "delete"}, strs(rules[0], "verbs"), name)
		assert.Equal(t, []string{"secrets"}, strs(rules[1], "resources"), name)
		assert.Nil(t, field(rules[1], "resourceNames"), name)
		assert.Equal(t, []string{"create"}, strs(rules[1], "verbs"), name)
	}

	accounts := map[string]string{}
	for _, b := range ofKind(objects, "RoleBinding") {
		accounts[str(b, "subjects", 0, "name")] = str(b, "roleRef", "name")
	}

	for _, d := range ofKind(objects, "Deployment") {
		name := str(d, "metadata", "name")

		assert.Equal(t, "default", str(podSpec(d), "serviceAccountName"), name)
		assert.Equal(t, "false", str(podSpec(d), "automountServiceAccountToken"), name)
	}

	for _, j := range append(ofKind(objects, "Job"), ofKind(objects, "CronJob")...) {
		assert.Contains(t, want, accounts[str(podSpec(j), "serviceAccountName")], str(j, "metadata", "name"))
	}
}

// Verifies: LAB-0126
func TestTheJobsAreNotHooksAndHaveTheRevisionInTheirNames(t *testing.T) {
	for _, job := range ofKind(render(t), "Job") {
		name := str(job, "metadata", "name")

		assert.True(t, strings.HasSuffix(name, "-r1"), name)
		assert.Nil(t, field(job, "metadata", "annotations", "helm.sh/hook"), name)
	}
}
