package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func TestARefusedChangeNamesTheRecordAndTheFieldOfEachValue(t *testing.T) {
	_, err := domain.HandleDataRecipient(
		dataRecipientState(validDataRecipient()),
		domain.ChangeDataRecipientDetails{Details: domain.DataRecipientDetails{
			LegalEntityName: "Recipient",
			LogoURI:         "https://recipient.example/logo.png",
		}},
	)

	fields, others := domain.FieldErrors(err)

	assert.Empty(t, others)
	assert.Equal(t, []domain.FieldError{
		{RecordID: "dr-1", Field: "accreditationNumber", Detail: "accreditationNumber is required"},
		{
			RecordID: "dr-1",
			Field:    "accreditationLevel",
			Detail:   `accreditationLevel "" is not UNRESTRICTED or SPONSORED`,
		},
	}, withoutText(fields))
	require.ErrorIs(t, err, domain.ErrNotValid)
	assert.EqualError(t, err, `data recipient "dr-1": accreditationNumber is required`+"\n"+
		`data recipient "dr-1": accreditationLevel "" is not UNRESTRICTED or SPONSORED`)
}

func TestABrokenStatusRuleIsAnErrorOfTheStatusField(t *testing.T) {
	dr := validDataRecipient()
	dr.Brands[0].Status = domain.Inactive
	dr.Brands[0].SoftwareProducts[0].Status = domain.Inactive

	_, err := domain.HandleDataRecipient(
		dataRecipientState(dr),
		domain.AddSoftwareProduct{BrandID: "drb-1", SoftwareProduct: secondSoftwareProduct()},
	)

	fields, others := domain.FieldErrors(err)

	assert.Empty(t, others)
	assert.Equal(t, []domain.FieldError{{
		RecordID: "sp-2",
		Field:    "status",
		Detail: `status is ACTIVE, but data recipient brand "drb-1" is INACTIVE: ` +
			`a software product of an INACTIVE data recipient brand must be INACTIVE or REMOVED`,
	}}, withoutText(fields))

	var ruleErr domain.StatusRuleError
	require.ErrorAs(t, err, &ruleErr)
	assert.EqualError(t, err, `software product "sp-2" is ACTIVE, but data recipient brand "drb-1" is INACTIVE: `+
		`a software product of an INACTIVE data recipient brand must be INACTIVE or REMOVED`)
}

func TestANestedFieldHasTheNameOfEachLevel(t *testing.T) {
	dh := validDataHolder()
	dh.Brands[0].EndpointDetail.PublicBaseURI = ""

	_, err := domain.HandleDataHolder(domain.DataHolderState{}, domain.CreateDataHolder{DataHolder: dh})

	fields, _ := domain.FieldErrors(err)

	require.Len(t, fields, 1)
	assert.Equal(t, "dhb-1", fields[0].RecordID)
	assert.Equal(t, "endpointDetail.publicBaseUri", fields[0].Field)
}

func TestAnErrorThatNamesNoFieldHasNoFieldErrors(t *testing.T) {
	_, err := domain.HandleDataRecipient(
		dataRecipientState(validDataRecipient()),
		domain.ReinstateAccreditation{},
	)

	fields, others := domain.FieldErrors(err)

	assert.Empty(t, fields)
	assert.Equal(t, []error{err}, others)
}

func TestFieldErrorsOfNoErrorIsEmpty(t *testing.T) {
	fields, others := domain.FieldErrors(nil)

	assert.Empty(t, fields)
	assert.Empty(t, others)
}

// withoutText keeps the fields that the test compares: the text of the error is the text of the operator interface.
func withoutText(fields []domain.FieldError) []domain.FieldError {
	out := make([]domain.FieldError, 0, len(fields))
	for _, f := range fields {
		out = append(out, domain.FieldError{RecordID: f.RecordID, Field: f.Field, Detail: f.Detail})
	}

	return out
}
