// Package dataholder has the black-box tests of a deployed data holder brand.
package dataholder

import (
	_ "embed"

	"github.com/lukemundy/cdr-lab/pkg/openapi"
)

//go:embed spec/cds_common.json
var commonSpecJSON []byte

//go:embed spec/cds_banking.json
var bankingSpecJSON []byte

//go:embed spec/cds_dcr.json
var dcrSpecJSON []byte

// The CDS gives the paths of the specification below this path of the base URI of the brand.
const basePath = "/cds-au/v1"

// LoadCommonSpec gives the pinned CDS common OpenAPI specification, which has Get Status and Get Outages.
func LoadCommonSpec() (*openapi.Spec, error) {
	return openapi.Load(commonSpecJSON, basePath)
}

// LoadBankingSpec gives the pinned CDS banking OpenAPI specification, which has Get Products and Get Product Detail.
func LoadBankingSpec() (*openapi.Spec, error) {
	return openapi.Load(bankingSpecJSON, basePath)
}

// LoadDCRSpec gives the pinned CDS Dynamic Client Registration OpenAPI specification. It has no base path, because a
// data holder gives the URL of its registration endpoint in its OpenID Provider configuration.
func LoadDCRSpec() (*openapi.Spec, error) {
	return openapi.Load(dcrSpecJSON, "")
}
