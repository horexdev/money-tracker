package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRecurringService(repo *mocks.MockRecurringStorer, txRepo *mocks.MockTransactionStorer) *service.RecurringService {
	return service.NewRecurringService(repo, txRepo, testutil.TestLogger())
}

func TestRecurringService_Create_ZeroAmount(t *testing.T) {
	svc := newRecurringService(&mocks.MockRecurringStorer{}, &mocks.MockTransactionStorer{})
	_, err := svc.Create(context.Background(), &domain.RecurringTransaction{
		AmountCents: 0, Frequency: domain.FrequencyMonthly,
	})
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestRecurringService_Create_InvalidFrequency(t *testing.T) {
	svc := newRecurringService(&mocks.MockRecurringStorer{}, &mocks.MockTransactionStorer{})
	_, err := svc.Create(context.Background(), &domain.RecurringTransaction{
		AmountCents: 1000, Frequency: "biweekly",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidFrequency)
}

func TestRecurringService_Create_SetsNextRunIfZero(t *testing.T) {
	repo := &mocks.MockRecurringStorer{}
	svc := newRecurringService(repo, &mocks.MockTransactionStorer{})

	rt := &domain.RecurringTransaction{AmountCents: 1000, Frequency: domain.FrequencyMonthly}
	before := time.Now()

	result := &domain.RecurringTransaction{ID: 1, AmountCents: 1000, Frequency: domain.FrequencyMonthly}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RecurringTransaction")).Return(result, nil)

	_, err := svc.Create(context.Background(), rt)
	require.NoError(t, err)
	assert.True(t, rt.NextRunAt.After(before), "NextRunAt should be set in the future")
}

func TestRecurringService_Update_InvalidFrequency(t *testing.T) {
	svc := newRecurringService(&mocks.MockRecurringStorer{}, &mocks.MockTransactionStorer{})
	_, err := svc.Update(context.Background(), &domain.RecurringTransaction{
		AmountCents: 1000, Frequency: "unknown",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidFrequency)
}

func TestRecurringService_ProcessDue_EmptyList(t *testing.T) {
	repo := &mocks.MockRecurringStorer{}
	txRepo := &mocks.MockTransactionStorer{}
	svc := newRecurringService(repo, txRepo)

	repo.On("GetDue", mock.Anything, mock.AnythingOfType("time.Time")).Return([]*domain.RecurringTransaction{}, nil)

	n, err := svc.ProcessDue(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestRecurringService_ProcessDue_CreatesTransactionAndAdvancesRun(t *testing.T) {
	repo := &mocks.MockRecurringStorer{}
	txRepo := &mocks.MockTransactionStorer{}
	svc := newRecurringService(repo, txRepo)

	rt := &domain.RecurringTransaction{
		ID:           1,
		UserID:       42,
		AmountCents:  5000,
		Frequency:    domain.FrequencyMonthly,
		Type:         domain.TransactionTypeExpense,
		CurrencyCode: "USD",
	}
	repo.On("GetDue", mock.Anything, mock.AnythingOfType("time.Time")).Return([]*domain.RecurringTransaction{rt}, nil)
	txRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(&domain.Transaction{ID: 10}, nil)
	repo.On("UpdateNextRun", mock.Anything, int64(1), mock.AnythingOfType("time.Time")).Return(nil)

	n, err := svc.ProcessDue(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	repo.AssertExpectations(t)
	txRepo.AssertExpectations(t)
}

func TestRecurringService_ProcessDue_SkipsOnTxError(t *testing.T) {
	repo := &mocks.MockRecurringStorer{}
	txRepo := &mocks.MockTransactionStorer{}
	svc := newRecurringService(repo, txRepo)

	rt := &domain.RecurringTransaction{
		ID:           1,
		UserID:       42,
		AmountCents:  5000,
		Frequency:    domain.FrequencyMonthly,
		Type:         domain.TransactionTypeExpense,
		CurrencyCode: "USD",
	}
	repo.On("GetDue", mock.Anything, mock.AnythingOfType("time.Time")).Return([]*domain.RecurringTransaction{rt}, nil)
	txRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil, errors.New("db error"))

	n, err := svc.ProcessDue(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, n) // skipped
	repo.AssertNotCalled(t, "UpdateNextRun")
}
