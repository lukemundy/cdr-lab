package branddir

import (
	"strings"

	shared "github.com/lukemundy/cdr-lab/pkg/platform/branddir"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	heroFile       = "hero.webp"
	categoryPrefix = "category-"
	photoSuffix    = ".webp"
)

var (
	heroLimit     = shared.Limit{Bytes: 256 << 10, Text: "256 KiB"}
	categoryLimit = shared.Limit{Bytes: 96 << 10, Text: "96 KiB"}
)

// The widths give a sharp photo on a screen with a pixel ratio of 2 at the sizes of the layouts.
var (
	heroWidths   = [2]int{1200, 2400}
	categoryRule = shared.PhotoRule{MinWidth: 800, MaxWidth: 1200, AspectRatio: 4.0 / 3, RatioOwner: "a category photo"}
	layoutNames  = []string{"meadow", "ledger", "horizon", "updraft"}
	heroRatios   = map[string]float64{"meadow": 4.0 / 3, "ledger": 21.0 / 9, "horizon": 3, "updraft": 1}
)

// photoLimit gives the size limit of a photo of the contract, and false for a name that the contract does not give.
func photoLimit(name string) (shared.Limit, bool) {
	if name == heroFile {
		return heroLimit, true
	}

	_, found := photoCategory(name)

	return categoryLimit, found
}

// photoCategory gives the CDS product category of a category photo, for example TERM_DEPOSITS for
// category-term-deposits.webp.
func photoCategory(name string) (domain.ProductCategory, bool) {
	anchor, found := strings.CutPrefix(name, categoryPrefix)
	if !found {
		return "", false
	}

	anchor, found = strings.CutSuffix(anchor, photoSuffix)
	if !found || strings.Contains(anchor, "_") {
		return "", false
	}

	return domain.ParseProductCategory(strings.ToUpper(strings.ReplaceAll(anchor, "-", "_")))
}

// checkPhoto has no aspect ratio check for a hero photo when the layout is not valid.
func checkPhoto(name string, data []byte, layout string) error {
	rule := categoryRule
	if name == heroFile {
		rule = shared.PhotoRule{
			MinWidth: heroWidths[0], MaxWidth: heroWidths[1], AspectRatio: heroRatios[layout],
			RatioOwner: "the layout " + layout,
		}
	}

	return shared.CheckPhoto(data, rule)
}
