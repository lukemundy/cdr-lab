// Package certlist reads the certificate list of a compose project: the Register operator interface on the host, and
// the inputs and the files of each certificate of the compose lab (cli/cdrlab/design/compose-lab.md#cli-0147).
package certlist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// File is the path of the list in the lab directory.
const File = "lab/certificates.json"

type List struct {
	Register Register `json:"register"`
	Public   []Public `json:"public"`
	CDR      []CDR    `json:"cdr"`
}

type Register struct {
	URL           string `json:"url"`
	OperatorToken string `json:"operator-token"`
}

type Public struct {
	Cert  string   `json:"cert"`
	Key   string   `json:"key"`
	Names []string `json:"names"`
}

// CDR has the flag names of cdrlab cert issue as its keys (CLI-0194).
type CDR struct {
	Kind              string   `json:"kind"`
	LegalEntityID     string   `json:"legal-entity-id"`
	BrandID           string   `json:"brand-id,omitempty"`
	SoftwareProductID string   `json:"software-product-id,omitempty"`
	CN                string   `json:"cn"`
	O                 string   `json:"o"`
	DNS               []string `json:"dns,omitempty"`
	Cert              string   `json:"cert"`
	Key               string   `json:"key"`
}

// Flags gives the flags of cdrlab cert issue for the certificate, with the files in labDir.
func (c CDR) Flags(labDir string) []string {
	flags := []string{"-kind", c.Kind, "-legal-entity-id", c.LegalEntityID}

	if c.BrandID != "" {
		flags = append(flags, "-brand-id", c.BrandID)
	}

	if c.SoftwareProductID != "" {
		flags = append(flags, "-software-product-id", c.SoftwareProductID)
	}

	flags = append(flags, "-cn", c.CN, "-o", c.O)
	for _, name := range c.DNS {
		flags = append(flags, "-dns", name)
	}

	return append(flags, "-cert", filepath.Join(labDir, c.Cert), "-key", filepath.Join(labDir, c.Key))
}

func Load(labDir string) (List, error) {
	data, err := fs.ReadFile(os.DirFS(labDir), File)
	if err != nil {
		return List{}, fmt.Errorf("read the certificate list: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var list List
	if err := dec.Decode(&list); err != nil {
		return List{}, fmt.Errorf("read the certificate list %s: %w", File, err)
	}

	if err := list.check(); err != nil {
		return List{}, fmt.Errorf("the certificate list %s: %w", File, err)
	}

	return list, nil
}

func (l List) check() error {
	if l.Register.URL == "" || l.Register.OperatorToken == "" {
		return errors.New("register.url and register.operator-token are required")
	}

	var paths []string

	for _, p := range l.Public {
		if len(p.Names) == 0 {
			return fmt.Errorf("the public certificate %s has no names", p.Cert)
		}

		paths = append(paths, p.Cert, p.Key)
	}

	for _, c := range l.CDR {
		paths = append(paths, c.Cert, c.Key)
	}

	for _, p := range paths {
		// A path outside the lab directory can change a file of the host (CLI-0193).
		if !filepath.IsLocal(p) {
			return fmt.Errorf("the path %q is not in the lab directory", p)
		}
	}

	return nil
}
