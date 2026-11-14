package readmodel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func TestAnIDOfADifferentParticipantIsNotUnique(t *testing.T) {
	store, _ := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataHolders: []domain.DataHolder{holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking})},
		DataRecipients: []domain.DataRecipient{
			recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Removed)}),
		},
	}))
	operator := app.NewOperator(store)

	tests := map[string]struct {
		legalEntityID string
		command       domain.DataHolderCommand
		wantID        string
	}{
		"a data holder with the ID of a data recipient": {
			legalEntityID: "dr-1",
			command:       domain.CreateDataHolder{DataHolder: holder("dr-1", domain.DataHolderActive)},
			wantID:        "dr-1",
		},
		"a brand with the ID of a removed software product": {
			legalEntityID: "dh-1",
			command: domain.AddDataHolderBrand{
				Brand: withBrandID(
					holder("dh-9", domain.DataHolderActive, []domain.Industry{domain.Energy}),
					"sp-1",
				).Brands[0],
			},
			wantID: "sp-1",
		},
		"a data holder with a brand ID of a different data holder": {
			legalEntityID: "dh-2",
			command: domain.CreateDataHolder{
				DataHolder: withBrandID(
					holder("dh-2", domain.DataHolderActive, []domain.Industry{domain.Energy}),
					"dh-1-b1",
				),
			},
			wantID: "dh-1-b1",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := operator.ExecuteDataHolder(t.Context(), tt.legalEntityID, tt.command)

			var notUnique domain.IDNotUniqueError
			require.ErrorAs(t, err, &notUnique)
			assert.Equal(t, tt.wantID, notUnique.ID)
		})
	}
}

func withBrandID(dh domain.DataHolder, id string) domain.DataHolder {
	dh.Brands[0].DataHolderBrandID = id
	return dh
}

func TestAChangeOfAParticipantKeepsItsOwnIDs(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataHolders: []domain.DataHolder{holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking})},
	}))

	err := app.NewOperator(store).ExecuteDataHolder(t.Context(), "dh-1", domain.ChangeDataHolderDetails{
		Details: domain.DataHolderDetails{LegalEntityName: "Changed", LogoURI: "https://holder.example/logo.png"},
	})

	require.NoError(t, err)

	var ids []string
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT array_agg(id ORDER BY id) FROM participant_ids").Scan(&ids))
	assert.Equal(t, []string{"dh-1", "dh-1-b1"}, ids)
}
