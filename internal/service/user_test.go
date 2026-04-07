package service_test

import (
	"context"
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_Upsert_PassesThroughAsIs(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	// Currency is now set by the caller (ensureUser derives it from language);
	// the service no longer injects a default.
	input := &domain.User{ID: 1, FirstName: "Alice", CurrencyCode: "UAH", Language: "uk"}
	repo.On("Upsert", context.Background(), input).Return(input, nil)

	got, err := svc.Upsert(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "UAH", got.CurrencyCode)
	repo.AssertExpectations(t)
}

func TestUserService_Upsert_PreservesExistingCurrency(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	input := &domain.User{ID: 1, CurrencyCode: "EUR"}
	repo.On("Upsert", context.Background(), input).Return(input, nil)

	got, err := svc.Upsert(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "EUR", got.CurrencyCode)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateCurrency_Invalid(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	_, err := svc.UpdateCurrency(context.Background(), 1, "INVALID")
	assert.ErrorIs(t, err, domain.ErrInvalidCurrency)
	repo.AssertNotCalled(t, "UpdateCurrency")
}

func TestUserService_UpdateCurrency_Valid(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	expected := &domain.User{ID: 1, CurrencyCode: "EUR"}
	repo.On("UpdateCurrency", context.Background(), int64(1), "EUR").Return(expected, nil)

	got, err := svc.UpdateCurrency(context.Background(), 1, "EUR")
	require.NoError(t, err)
	assert.Equal(t, "EUR", got.CurrencyCode)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateLanguage_Invalid(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	_, err := svc.UpdateLanguage(context.Background(), 1, "zh")
	assert.ErrorIs(t, err, domain.ErrInvalidLanguage)
	repo.AssertNotCalled(t, "UpdateLanguage")
}

func TestUserService_UpdateLanguage_Valid(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	expected := &domain.User{ID: 1, Language: domain.LangEN}
	repo.On("UpdateLanguage", context.Background(), int64(1), "en").Return(expected, nil)

	got, err := svc.UpdateLanguage(context.Background(), 1, "en")
	require.NoError(t, err)
	assert.Equal(t, domain.LangEN, got.Language)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateDisplayCurrencies_TooMany(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	_, err := svc.UpdateDisplayCurrencies(context.Background(), 1, []string{"USD", "EUR", "GBP", "JPY"})
	assert.ErrorIs(t, err, domain.ErrTooManyDisplayCurrencies)
	repo.AssertNotCalled(t, "UpdateDisplayCurrencies")
}

func TestUserService_UpdateDisplayCurrencies_InvalidCode(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	_, err := svc.UpdateDisplayCurrencies(context.Background(), 1, []string{"USD", "FAKE"})
	assert.ErrorIs(t, err, domain.ErrInvalidCurrency)
	repo.AssertNotCalled(t, "UpdateDisplayCurrencies")
}

func TestUserService_UpdateDisplayCurrencies_Valid(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	expected := &domain.User{ID: 1}
	repo.On("UpdateDisplayCurrencies", context.Background(), int64(1), "USD,EUR").Return(expected, nil)

	_, err := svc.UpdateDisplayCurrencies(context.Background(), 1, []string{"USD", "EUR"})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateNotificationPreferences_Success(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	prefs := domain.NotificationPrefs{
		BudgetAlerts:       true,
		RecurringReminders: false,
		WeeklySummary:      true,
		GoalMilestones:     false,
	}
	expected := &domain.User{
		ID:                       1,
		NotifyBudgetAlerts:       true,
		NotifyRecurringReminders: false,
		NotifyWeeklySummary:      true,
		NotifyGoalMilestones:     false,
	}
	repo.On("UpdateNotificationPreferences", context.Background(), int64(1), prefs).Return(expected, nil)

	got, err := svc.UpdateNotificationPreferences(context.Background(), 1, prefs)
	require.NoError(t, err)
	assert.True(t, got.NotifyBudgetAlerts)
	assert.True(t, got.NotifyWeeklySummary)
	assert.False(t, got.NotifyRecurringReminders)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateNotificationPreferences_RepoError(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	svc := service.NewUserService(repo, testutil.TestLogger())

	prefs := domain.NotificationPrefs{}
	repo.On("UpdateNotificationPreferences", context.Background(), int64(1), prefs).Return(nil, domain.ErrUserNotFound)

	_, err := svc.UpdateNotificationPreferences(context.Background(), 1, prefs)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	repo.AssertExpectations(t)
}
