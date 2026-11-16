package domain_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type uriField struct {
	record string
	field  string
	set    func(dr *domain.DataRecipient, dh *domain.DataHolder, value string)
}

func (f uriField) importWith(value string) error {
	dr, dh := validDataRecipient(), validDataHolder()
	f.set(&dr, &dh, value)

	if _, err := importDataRecipient(dr); err != nil {
		return err
	}

	_, err := importDataHolder(dh)

	return err
}

// want gives the error of the field. The error of a list is about one of its values.
func (f uriField) want(problem string) string {
	if f.field == "redirectUris" {
		return fmt.Sprintf("%s: %s has a value that %s", f.record, f.field, problem)
	}

	return fmt.Sprintf("%s: %s %s", f.record, f.field, problem)
}

func product(dr *domain.DataRecipient) *domain.SoftwareProduct {
	return &dr.Brands[0].SoftwareProducts[0]
}

func endpoints(dh *domain.DataHolder) *domain.EndpointDetail { return &dh.Brands[0].EndpointDetail }

var requiredURIFields = []uriField{
	{
		record: `data recipient "dr-1"`,
		field:  "logoUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			dr.LogoURI = v
		},
	},
	{
		record: `data recipient brand "drb-1"`,
		field:  "logoUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			dr.Brands[0].LogoURI = v
		},
	},
	{
		record: `software product "sp-1"`,
		field:  "logoUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).LogoURI = v
		},
	},
	{
		record: `software product "sp-1"`,
		field:  "clientUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).ClientURI = v
		},
	},
	{
		record: `software product "sp-1"`,
		field:  "jwksUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).JWKSURI = v
		},
	},
	{
		record: `software product "sp-1"`,
		field:  "revocationUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).RevocationURI = v
		},
	},
	{
		record: `software product "sp-1"`,
		field:  "recipientBaseUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).RecipientBaseURI = v
		},
	},
	{
		record: `data holder "dh-1"`,
		field:  "logoUri",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			dh.LogoURI = v
		},
	},
	{
		record: `data holder brand "dhb-1"`,
		field:  "logoUri",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			dh.Brands[0].LogoURI = v
		},
	},
	{
		record: `data holder brand "dhb-1"`,
		field:  "endpointDetail.publicBaseUri",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			endpoints(dh).PublicBaseURI = v
		},
	},
	{
		record: `data holder brand "dhb-1"`,
		field:  "endpointDetail.resourceBaseUri",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			endpoints(dh).ResourceBaseURI = v
		},
	},
	{
		record: `data holder brand "dhb-1"`,
		field:  "endpointDetail.infosecBaseUri",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			endpoints(dh).InfosecBaseURI = v
		},
	},
	{
		record: `data holder brand "dhb-1"`,
		field:  "endpointDetail.websiteUri",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			endpoints(dh).WebsiteURI = v
		},
	},
	{
		record: `data holder brand "dhb-1"`,
		field:  "jwksEndpoint",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			dh.Brands[0].JWKSEndpoint = v
		},
	},
}

var optionalURIFields = []uriField{
	{
		record: `software product "sp-1"`,
		field:  "sectorIdentifierUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).SectorIdentifierURI = v
		},
	},
	{
		record: `software product "sp-1"`,
		field:  "tosUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).TOSURI = v
		},
	},
	{
		record: `software product "sp-1"`,
		field:  "policyUri",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).PolicyURI = v
		},
	},
	{
		record: `software product "sp-1"`,
		field:  "redirectUris",
		set: func(dr *domain.DataRecipient, _ *domain.DataHolder, v string) {
			product(dr).RedirectURIs = append(product(dr).RedirectURIs, v)
		},
	},
	{
		record: `data holder brand "dhb-1"`,
		field:  "endpointDetail.productBaseUri",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			endpoints(dh).ProductBaseURI = v
		},
	},
	{
		record: `data holder brand "dhb-1"`,
		field:  "endpointDetail.extensionBaseUri",
		set: func(_ *domain.DataRecipient, dh *domain.DataHolder, v string) {
			endpoints(dh).ExtensionBaseURI = v
		},
	},
}

// Verifies: REG-0215
func TestAURIFieldRefusesAValueThatIsNotAnHTTPSURIWithAHost(t *testing.T) {
	values := map[string]string{
		"not a URI":              "foo",
		"relative reference":     "//holder.example/logo.png",
		"http scheme":            "http://holder.example",
		"other scheme":           "urn:example:logo",
		"no host":                "https:///logo.png",
		"space":                  "https://holder.example/a logo.png",
		"white space at the end": "https://holder.example\n",
		"character not in a URI": "https://holder.example/<logo>",
		"character not in ASCII": "https://holder.example/é",
	}
	for _, f := range append(requiredURIFields, optionalURIFields...) {
		for name, value := range values {
			t.Run(f.record+" "+f.field+" "+name, func(t *testing.T) {
				err := f.importWith(value)

				assert.EqualError(t, err, f.want("is not an https URI with a host"))
			})
		}
	}
}

// Verifies: REG-0216
func TestAURIFieldAcceptsAnHTTPSURIAtTheMaximumLength(t *testing.T) {
	prefix := "https://holder.example/"
	value := prefix + strings.Repeat("a", domain.MaxURILength-len(prefix))

	for _, f := range append(requiredURIFields, optionalURIFields...) {
		t.Run(f.record+" "+f.field, func(t *testing.T) {
			require.NoError(t, f.importWith(value))
		})
	}
}

// Verifies: REG-0217
func TestAURIFieldAcceptsAnHTTPSURIWithAPortAQueryAndPercentEncoding(t *testing.T) {
	for _, f := range append(requiredURIFields, optionalURIFields...) {
		t.Run(f.record+" "+f.field, func(t *testing.T) {
			require.NoError(t, f.importWith("HTTPS://Holder.example:8443/a%20logo.png?size=large&v=1"))
		})
	}
}

// Verifies: REG-0216
func TestAURIFieldRefusesAValueLongerThanTheMaximumLength(t *testing.T) {
	prefix := "https://holder.example/"
	value := prefix + strings.Repeat("a", domain.MaxURILength-len(prefix)+1)

	for _, f := range append(requiredURIFields, optionalURIFields...) {
		t.Run(f.record+" "+f.field, func(t *testing.T) {
			err := f.importWith(value)

			assert.EqualError(t, err, f.want(fmt.Sprintf("is longer than %d characters", domain.MaxURILength)))
		})
	}
}

func TestARequiredURIFieldRefusesNoValue(t *testing.T) {
	for _, f := range requiredURIFields {
		t.Run(f.record+" "+f.field, func(t *testing.T) {
			err := f.importWith("")

			assert.EqualError(t, err, fmt.Sprintf("%s: %s is required", f.record, f.field))
		})
	}
}

// Verifies: REG-0218
func TestAnOptionalURIFieldAcceptsNoValue(t *testing.T) {
	dr := validDataRecipient()
	product(&dr).SectorIdentifierURI, product(&dr).TOSURI, product(&dr).PolicyURI = "", "", ""
	dh := validDataHolder()
	endpoints(&dh).ProductBaseURI, endpoints(&dh).ExtensionBaseURI = "", ""

	_, drErr := importDataRecipient(dr)
	_, dhErr := importDataHolder(dh)

	require.NoError(t, drErr)
	require.NoError(t, dhErr)
}

func TestTheSecureEndpointsOfAnInterimOnlyBrandAreCheckedWhenTheyHaveAValue(t *testing.T) {
	dh := validDataHolder()
	dh.Brands[0] = interimOnlyBrand()
	dh.Brands[0].EndpointDetail.ResourceBaseURI = "foo"

	_, err := importDataHolder(dh)

	assert.EqualError(
		t,
		err,
		`data holder brand "dhi-1": endpointDetail.resourceBaseUri is not an https URI with a host`,
	)
}
