//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository"
	"github.com/horexdev/money-tracker/internal/testutil"
)

func newUserRepoFixtures(t *testing.T) (*repository.UserRepository, *testutilPool) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })
	return repository.NewUserRepository(pool), &testutilPool{pool: pool, idCounter: time.Now().UnixNano()}
}

// testutilPool tracks the next user ID to keep tests deterministic across calls
// to time.Now().UnixNano(), which can collide on fast Windows clocks.
type testutilPool struct {
	pool      interface{}
	idCounter int64
}

func (p *testutilPool) nextID() int64 {
	p.idCounter++
	return p.idCounter
}

func TestUserRepository_Upsert_CreatesNewUser(t *testing.T) {
	repo, ids := newUserRepoFixtures(t)
	ctx := context.Background()

	user, err := repo.Upsert(ctx, &domain.User{
		ID:        ids.nextID(),
		Username:  "alice",
		FirstName: "Alice",
		LastName:  "A.",
		Language:  domain.LangEN,
	})
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "Alice", user.FirstName)
	assert.Equal(t, domain.LangEN, user.Language)
}

func TestUserRepository_Upsert_UpdatesExisting(t *testing.T) {
	repo, ids := newUserRepoFixtures(t)
	ctx := context.Background()

	id := ids.nextID()
	_, err := repo.Upsert(ctx, &domain.User{ID: id, Username: "old", FirstName: "Old", Language: domain.LangEN})
	require.NoError(t, err)

	updated, err := repo.Upsert(ctx, &domain.User{ID: id, Username: "renamed", FirstName: "New", Language: domain.LangEN})
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Username)
	assert.Equal(t, "New", updated.FirstName)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	repo, _ := newUserRepoFixtures(t)
	_, err := repo.GetByID(context.Background(), 99_999_999)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserRepository_UpdateLanguage(t *testing.T) {
	repo, ids := newUserRepoFixtures(t)
	ctx := context.Background()

	id := ids.nextID()
	_, err := repo.Upsert(ctx, &domain.User{ID: id, Username: "u", FirstName: "U", Language: domain.LangEN})
	require.NoError(t, err)

	updated, err := repo.UpdateLanguage(ctx, id, "ru")
	require.NoError(t, err)
	assert.Equal(t, domain.Language("ru"), updated.Language)
}

func TestUserRepository_UpdateNotificationPreferences(t *testing.T) {
	repo, ids := newUserRepoFixtures(t)
	ctx := context.Background()

	id := ids.nextID()
	_, err := repo.Upsert(ctx, &domain.User{ID: id, Username: "u", FirstName: "U", Language: domain.LangEN})
	require.NoError(t, err)

	prefs := domain.NotificationPrefs{
		BudgetAlerts:       false,
		RecurringReminders: true,
		WeeklySummary:      false,
		GoalMilestones:     true,
	}
	updated, err := repo.UpdateNotificationPreferences(ctx, id, prefs)
	require.NoError(t, err)
	assert.False(t, updated.NotifyBudgetAlerts)
	assert.True(t, updated.NotifyRecurringReminders)
	assert.False(t, updated.NotifyWeeklySummary)
	assert.True(t, updated.NotifyGoalMilestones)
}

func TestUserRepository_UpdateDisplayCurrencies(t *testing.T) {
	repo, ids := newUserRepoFixtures(t)
	ctx := context.Background()

	id := ids.nextID()
	_, err := repo.Upsert(ctx, &domain.User{ID: id, Username: "u", FirstName: "U", Language: domain.LangEN})
	require.NoError(t, err)

	updated, err := repo.UpdateDisplayCurrencies(ctx, id, "USD,EUR,RUB")
	require.NoError(t, err)
	assert.Equal(t, []string{"USD", "EUR", "RUB"}, updated.DisplayCurrencies)
}
