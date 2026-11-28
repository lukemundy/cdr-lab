// Package branddir reads the brand directory of a data holder, and serves the brand to the pages, to the online bank
// site and to the operator console. The shared files and rules are in pkg/platform/branddir. This package adds the
// files and the rules of the data holder.
//
// ADR 0023 (docs/adr/0023-give-the-brand-of-a-data-holder-as-a-strict-brand-directory.md) gives the contract.
package branddir

import (
	"bytes"
	"errors"
	"io/fs"
	"maps"
	"slices"
	"strconv"
	"strings"

	shared "github.com/lukemundy/cdr-lab/pkg/platform/branddir"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const productsFile = "products.json"

type Brand struct {
	shared.Directory

	// BSB is the bsb of the accounts that a consumer opens on the online bank site.
	BSB string
	// Layout is the layout of the online bank site.
	Layout string
	// Products is the product catalogue of the brand, each product as JSON text.
	Products [][]byte

	// photos has the optional photos of the online bank site, by file name.
	photos map[string][]byte
}

// Photos gives the file name of each photo of the brand, in order.
func (b Brand) Photos() []string {
	return slices.Sorted(maps.Keys(b.photos))
}

// ProductCheck validates the products of the catalogue. Its error names each product and rule that fails.
type ProductCheck func(products [][]byte) error

type brandJSON struct {
	shared.BrandFields

	BSB    string `json:"bsb"`
	Layout string `json:"layout"`
}

// Load validates each file of the directory at the root of dir, and gives one error that names each file and rule that
// fails.
func Load(dir fs.FS, checkProducts ProductCheck) (Brand, error) {
	var (
		fields     brandJSON
		brandValid bool
		b          Brand
	)

	spec := shared.Spec{
		BrandFile: &fields,
		CheckBrandFile: func() error {
			if err := checkBrandFields(fields); err != nil {
				return err
			}

			brandValid = true

			return nil
		},
		Files:    map[string]shared.Limit{productsFile: shared.JSONLimit},
		Optional: photoLimit,
		Check: func(d shared.Directory) []error {
			// A hero photo has no aspect ratio check when the brand file is not valid.
			layout := ""
			if brandValid {
				layout = fields.Layout
			}

			var errs []error

			b.photos = map[string][]byte{}

			for _, name := range slices.Sorted(maps.Keys(d.Files)) {
				if _, isPhoto := photoLimit(name); isPhoto {
					errs = appendError(errs, shared.FileError(name, checkPhoto(name, d.Files[name], layout)))
					b.photos[name] = d.Files[name]
				}
			}

			if data, found := d.Files[productsFile]; found {
				products, err := ReadProducts(bytes.NewReader(data))
				if err == nil {
					err = checkProducts(products)
				}

				errs = appendError(errs, shared.FileError(productsFile, err))
				b.Products = products
			}

			return errs
		},
	}

	d, err := shared.Load(dir, spec)
	if err != nil {
		return Brand{}, err
	}

	b.Directory = d
	b.BSB, b.Layout = fields.BSB, fields.Layout

	return b, nil
}

func appendError(errs []error, err error) []error {
	if err == nil {
		return errs
	}

	return append(errs, err)
}

func checkBrandFields(b brandJSON) error {
	if !domain.ValidBSB(b.BSB) {
		return errors.New("bsb: give 6 digits")
	}

	if !slices.Contains(layoutNames, b.Layout) {
		quoted := make([]string, len(layoutNames))
		for i, name := range layoutNames {
			quoted[i] = strconv.Quote(name)
		}

		return errors.New("layout: give one of " + strings.Join(quoted, ", "))
	}

	return nil
}
