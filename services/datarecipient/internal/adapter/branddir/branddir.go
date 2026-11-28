// Package branddir reads the brand directory of a data recipient. The shared files and rules are in
// pkg/platform/branddir. This package adds the photos of the data recipient.
//
// The specification is services/datarecipient/design/brand.md.
package branddir

import (
	"io/fs"
	"maps"
	"slices"

	shared "github.com/lukemundy/cdr-lab/pkg/platform/branddir"
)

type photo struct {
	limit shared.Limit
	rule  shared.PhotoRule
}

// photos has the optional photos of the web app, by file name. The widths give a sharp photo on a screen with a pixel
// ratio of 2.
var photos = map[string]photo{
	"hero.webp": {
		limit: shared.Limit{Bytes: 256 << 10, Text: "256 KiB"},
		rule:  shared.PhotoRule{MinWidth: 1200, MaxWidth: 2400, AspectRatio: 4.0 / 3, RatioOwner: "the hero photo"},
	},
	"access.webp": {
		limit: shared.Limit{Bytes: 160 << 10, Text: "160 KiB"},
		rule:  shared.PhotoRule{MinWidth: 800, MaxWidth: 1600, AspectRatio: 4.0 / 5, RatioOwner: "the access photo"},
	},
}

type Brand struct {
	shared.Directory

	// photos has the photos of the brand, by file name.
	photos map[string][]byte
}

// Photos gives the file name of each photo of the brand, in order.
func (b Brand) Photos() []string {
	return slices.Sorted(maps.Keys(b.photos))
}

// Load validates each file of the directory at the root of dir, and gives one error that names each file and rule that
// fails.
func Load(dir fs.FS) (Brand, error) {
	d, err := shared.Load(dir, shared.Spec{
		Optional: func(name string) (shared.Limit, bool) {
			p, found := photos[name]
			return p.limit, found
		},
		Check: func(d shared.Directory) []error {
			var errs []error

			for _, name := range slices.Sorted(maps.Keys(d.Files)) {
				if err := shared.FileError(name, shared.CheckPhoto(d.Files[name], photos[name].rule)); err != nil {
					errs = append(errs, err)
				}
			}

			return errs
		},
	})
	if err != nil {
		return Brand{}, err
	}

	return Brand{Directory: d, photos: d.Files}, nil
}
