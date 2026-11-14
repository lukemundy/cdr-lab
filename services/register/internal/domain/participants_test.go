package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func TestParticipantsWithUniqueIDsAreValid(t *testing.T) {
	p := domain.Participants{
		DataHolders:    []domain.DataHolder{validDataHolder()},
		DataRecipients: []domain.DataRecipient{validDataRecipient()},
	}

	assert.NoError(t, p.Validate())
}

func TestParticipantsRejectAnIDInMoreThanOneAggregate(t *testing.T) {
	other := validDataRecipient()
	other.LegalEntityID = "dr-2"
	other.Brands[0].DataRecipientBrandID = "drb-2"
	holder := validDataHolder()
	holder.Brands[0].DataHolderBrandID = "sp-1"
	p := domain.Participants{
		DataHolders:    []domain.DataHolder{holder},
		DataRecipients: []domain.DataRecipient{validDataRecipient(), other},
	}

	assert.EqualError(t, p.Validate(), `ID "sp-1" is not unique: it is in data holder "dh-1" and data recipient "dr-1"
ID "sp-1" is not unique: it is in data holder "dh-1" and data recipient "dr-2"`)
}

func TestParticipantsIgnoreADuplicateIDInOneAggregate(t *testing.T) {
	dr := validDataRecipient()
	dr.Brands[0].SoftwareProducts[0].SoftwareProductID = "drb-1"

	assert.NoError(
		t,
		domain.Participants{DataRecipients: []domain.DataRecipient{dr}}.Validate(),
		"import of the aggregate reports it",
	)
}

func TestParticipantsRejectAnInterimIDInMoreThanOneAggregate(t *testing.T) {
	holder := validDataHolder()
	holder.Brands[0].InterimID = "dr-1"

	assert.EqualError(t, domain.Participants{
		DataHolders:    []domain.DataHolder{holder},
		DataRecipients: []domain.DataRecipient{validDataRecipient()},
	}.Validate(), `ID "dr-1" is not unique: it is in data holder "dh-1" and data recipient "dr-1"`)
}

func TestParticipantsAcceptInterimOnlyBrandsInMoreThanOneAggregate(t *testing.T) {
	first := validDataHolder()
	first.Brands = append(first.Brands, interimOnlyBrand())
	second := validDataHolder()
	second.LegalEntityID = "dh-2"
	second.Brands[0] = interimOnlyBrand()
	second.Brands[0].InterimID = "dhi-2"

	assert.NoError(t, domain.Participants{DataHolders: []domain.DataHolder{first, second}}.Validate())
}
