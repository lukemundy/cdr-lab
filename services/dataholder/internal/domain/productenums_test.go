package domain_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func TestProductEnumerationsGivesEachEnumerationOfAProduct(t *testing.T) {
	enums := domain.ProductEnumerations()

	assert.ElementsMatch(t, []string{
		"ProductCategory", "FeatureType", "FeeType", "FeeMethodUType", "FeeRateType", "DepositRateType",
		"LendingRateType", "ApplicationType", "RepaymentType", "LoanPurpose",
	}, slices.Collect(maps.Keys(enums)))

	for _, category := range enums["ProductCategory"] {
		_, ok := domain.ParseProductCategory(category)
		assert.True(t, ok, category)
	}

	assert.Contains(t, enums["FeatureType"], string(domain.FeatureNppPayid))
	assert.Contains(t, enums["LoanPurpose"], string(domain.PurposeOwnerOccupied))
}

func TestTheCallerCannotChangeTheEnumerations(t *testing.T) {
	domain.ProductEnumerations()["ProductCategory"][0] = "CHANGED"

	assert.NotContains(t, domain.ProductEnumerations()["ProductCategory"], "CHANGED")
}
