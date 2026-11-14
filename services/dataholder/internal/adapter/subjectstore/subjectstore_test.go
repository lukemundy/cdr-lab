package subjectstore_test

import (
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/subjectstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

const (
	firstCustomer  = "00000001"
	secondCustomer = "00000002"
)

// newPool gives a database with two customers.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	for _, number := range []string{firstCustomer, secondCustomer} {
		require.NoError(t, customerstore.New(pool, nil, nil).AddCustomer(t.Context(), domain.Customer{
			CustomerNumber: number, LastName: "Example", EmailAddress: number + "@example.com",
		}))
	}

	return pool
}

func TestTheFirstCandidateOfAPairStaysItsSubject(t *testing.T) {
	pool := newPool(t)

	first, err := subjectstore.New(pool).Subject(t.Context(), firstCustomer, "app.example", "subject-1")
	require.NoError(t, err)
	second, err := subjectstore.New(pool).Subject(t.Context(), firstCustomer, "app.example", "subject-2")
	require.NoError(t, err)

	assert.Equal(t, "subject-1", first)
	assert.Equal(t, "subject-1", second)
}

func TestADifferentCustomerOrSectorIdentifierHasADifferentSubject(t *testing.T) {
	store := subjectstore.New(newPool(t))
	_, err := store.Subject(t.Context(), firstCustomer, "app.example", "subject-1")
	require.NoError(t, err)

	otherCustomer, err := store.Subject(t.Context(), secondCustomer, "app.example", "subject-2")
	require.NoError(t, err)
	otherSector, err := store.Subject(t.Context(), firstCustomer, "other.example", "subject-3")
	require.NoError(t, err)

	assert.Equal(t, "subject-2", otherCustomer)
	assert.Equal(t, "subject-3", otherSector)
}

func TestASubjectIsNotUsedForTwoPairs(t *testing.T) {
	store := subjectstore.New(newPool(t))
	_, err := store.Subject(t.Context(), firstCustomer, "app.example", "subject-1")
	require.NoError(t, err)

	_, err = store.Subject(t.Context(), secondCustomer, "app.example", "subject-1")

	require.Error(t, err)
}

func TestACustomerThatDoesNotExistGetsNoSubject(t *testing.T) {
	_, err := subjectstore.New(newPool(t)).Subject(t.Context(), "99999999", "app.example", "subject-1")

	require.Error(t, err)
}

func TestRequestsAtTheSameTimeGetOneSubject(t *testing.T) {
	store := subjectstore.New(newPool(t))
	candidates := []string{"subject-1", "subject-2", "subject-3", "subject-4"}
	subjects := make([]string, len(candidates))

	var wg sync.WaitGroup

	for i, candidate := range candidates {
		wg.Go(func() {
			subject, err := store.Subject(t.Context(), firstCustomer, "app.example", candidate)
			assert.NoError(t, err)

			subjects[i] = subject
		})
	}

	wg.Wait()

	for _, subject := range subjects {
		assert.Equal(t, subjects[0], subject)
	}
}
