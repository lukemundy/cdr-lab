package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var postedAt = time.Date(2026, 10, 14, 9, 30, 0, 0, time.UTC)

func postedTransaction() domain.Transaction {
	return domain.Transaction{
		Type:            domain.TransactionPayment,
		Status:          domain.TransactionPosted,
		Description:     "Coffee",
		Amount:          -450,
		PostingDateTime: &postedAt,
	}
}

func npp(amount domain.Amount) domain.Transaction {
	t := postedTransaction()
	t.Amount = amount
	t.ExtendedData = &domain.ExtendedData{
		ExtensionUType: domain.ExtensionNPPPayload,
		NPPPayload: domain.NPPPayload{
			Service: domain.NPPServiceX2P1, ServiceVersion: "03", ExtendedDescription: "Rent for October",
		},
	}

	if amount < 0 {
		t.ExtendedData.Payee = "Landlord"
	} else {
		t.ExtendedData.Payer = "Employer"
	}

	return t
}

func TestAValidTransactionHasNoError(t *testing.T) {
	pending := postedTransaction()
	pending.Status, pending.PostingDateTime, pending.ExecutionDateTime = domain.TransactionPending, nil, &postedAt

	for name, transaction := range map[string]domain.Transaction{
		"posted":          postedTransaction(),
		"pending":         pending,
		"NPP to a payee":  npp(-120_000),
		"NPP from payer":  npp(350_000),
		"zero amount NPP": npp(0),
	} {
		require.NoError(t, transaction.Validate(), name)
	}
}

func TestATransactionNamesEachRuleThatItBreaks(t *testing.T) {
	tests := map[string]struct {
		change func(*domain.Transaction)
		field  string
		want   string
	}{
		"unknown type": {
			change: func(tr *domain.Transaction) { tr.Type = "PURCHASE" },
			field:  "type",
			want:   "type must be a value of the CDS enumeration",
		},
		"unknown status": {
			change: func(tr *domain.Transaction) { tr.Status = "SETTLED" },
			field:  "status",
			want:   "status must be PENDING or POSTED",
		},
		"no description": {
			change: func(tr *domain.Transaction) { tr.Description = "" },
			field:  "description",
			want:   "description must have 1 to 500 characters",
		},
		"long description": {
			change: func(tr *domain.Transaction) { tr.Description = strings.Repeat("a", 501) },
			field:  "description",
			want:   "description must have 1 to 500 characters",
		},
		"posted with no posting time": {
			change: func(tr *domain.Transaction) { tr.PostingDateTime = nil },
			field:  "postingDateTime",
			want:   "a POSTED transaction must have a postingDateTime",
		},
		"pending with no execution time": {
			change: func(tr *domain.Transaction) { tr.Status = domain.TransactionPending },
			field:  "executionDateTime",
			want:   "a PENDING transaction must have an executionDateTime",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			transaction := postedTransaction()
			tt.change(&transaction)

			assert.Equal(
				t,
				[]domain.FieldError{{Field: tt.field, Detail: tt.want}},
				fieldErrors(t, transaction.Validate()),
			)
		})
	}
}

func TestTheExtendedDataOfATransactionNamesEachRuleThatItBreaks(t *testing.T) {
	tests := map[string]struct {
		transaction domain.Transaction
		change      func(*domain.ExtendedData)
		field       string
		want        string
	}{
		"unknown extension": {
			transaction: npp(-100),
			change:      func(e *domain.ExtendedData) { e.ExtensionUType = "other" },
			field:       "extendedData.extensionUType",
			want:        "extendedData.extensionUType must be nppPayload",
		},
		"unknown service": {
			transaction: npp(-100),
			change:      func(e *domain.ExtendedData) { e.NPPPayload.Service = "OSKO" },
			field:       "extendedData.nppPayload.service",
			want:        "extendedData.nppPayload.service must be a value of the CDS enumeration",
		},
		"service version of one digit": {
			transaction: npp(-100),
			change:      func(e *domain.ExtendedData) { e.NPPPayload.ServiceVersion = "3" },
			field:       "extendedData.nppPayload.serviceVersion",
			want:        "extendedData.nppPayload.serviceVersion must have 2 digits",
		},
		"no extended description": {
			transaction: npp(-100),
			change:      func(e *domain.ExtendedData) { e.NPPPayload.ExtendedDescription = "" },
			field:       "extendedData.nppPayload.extendedDescription",
			want:        "extendedData.nppPayload.extendedDescription must have 1 character or more",
		},
		"outgoing payment with no payee": {
			transaction: npp(-100),
			change:      func(e *domain.ExtendedData) { e.Payee = "" },
			field:       "extendedData.payee",
			want:        "extendedData.payee must be present for a negative amount",
		},
		"incoming payment with no payer": {
			transaction: npp(100),
			change:      func(e *domain.ExtendedData) { e.Payer = "" },
			field:       "extendedData.payer",
			want:        "extendedData.payer must be present for a positive amount",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.change(tt.transaction.ExtendedData)

			assert.Equal(
				t, []domain.FieldError{{Field: tt.field, Detail: tt.want}}, fieldErrors(t, tt.transaction.Validate()),
			)
		})
	}
}

func TestTheEffectiveTimeIsThePostingTimeOrElseTheExecutionTime(t *testing.T) {
	executedAt := postedAt.Add(-time.Hour)
	posted := postedTransaction()
	posted.ExecutionDateTime = &executedAt
	pending := postedTransaction()
	pending.PostingDateTime, pending.ExecutionDateTime = nil, &executedAt

	assert.Equal(t, postedAt, posted.EffectiveTime())
	assert.Equal(t, executedAt, pending.EffectiveTime())
}

func TestOnlyATransactionWithExtendedDataHasDetail(t *testing.T) {
	assert.False(t, postedTransaction().IsDetailAvailable())
	assert.True(t, npp(-100).IsDetailAvailable())
}

func TestATransactionInUTCHasEachTimeInUTCAndTheSameInstant(t *testing.T) {
	perth := postedAt.In(time.FixedZone("AWST", 8*60*60))
	transaction := postedTransaction()
	transaction.PostingDateTime, transaction.ValueDateTime = &perth, &perth

	got := transaction.InUTC()

	assert.Equal(t, postedAt, *got.PostingDateTime)
	assert.Equal(t, time.UTC, got.ValueDateTime.Location())
	assert.Nil(t, got.ExecutionDateTime)
	assert.Equal(t, perth.Location(), transaction.PostingDateTime.Location(), "InUTC does not change its receiver")
}
