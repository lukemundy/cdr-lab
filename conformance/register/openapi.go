// Package register contains black-box conformance tests for a deployed CDR Register.
package register

import (
	_ "embed"

	"github.com/lukemundy/cdr-lab/pkg/openapi"
)

//go:embed spec/cds_register.json
var specJSON []byte

func LoadSpec() (*openapi.Spec, error) {
	return openapi.Load(specJSON, "")
}
