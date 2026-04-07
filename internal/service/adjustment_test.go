package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var adjustmentCat = &domain.Category{ID: 77, Name: "Adjustment"}

func newAdjustmentService(txRepo *mocks.MockTransactionStorer, accounts *mocks.MockAccountStorer, catRepo *mocks.MockCategoryStorer) *service.AdjustmentService {
	return service.NewAdjustmentService(txRepo, accounts, catRepo, testutil.TestLogger())
}

func TestAdjustmentService_Apply_ZeroDelta(t *testing.T) {
	svc := newAdjustmentService(&mocks.MockTransactionStorer{}, &mocks.MockAccountStorer{}, &mocks.MockCategoryStorer{})
	_, err := svc.Apply(context.Background(), 1, 10, 0, "")
	assert.ErrorIs(t, err, domain.ErrAdjustmentZeroAmount)
}

func TestAdjustmentService_Apply_AccountNotFound(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	accounts := &mocks.MockAccountStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newAdjustmentService(txRepo, accounts, catRepo)

	accounts.On("GetByID", mock.Anything, int64(10), int64(1)).Return(nil, domain.ErrAccountNotFound)

	_, err := svc.Apply(context.Background(), 1, 10, 500, "")
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestAdjustmentService_Apply_PositiveDelta_CreatesIncomeTransaction(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	accounts := &mocks.MockAccountStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newAdjustmentService(txRepo, accounts, catRepo)

	acc := &domain.Account{ID: 10, UserID: 1, CurrencyCode: "USD"}
	accounts.On("GetByID", mock.Anything, int64(10), int64(1)).Return(acc, nil)
	catRepo.On("GetSystemCategoryByType", mock.Anything, "adjustment").Return(adjustmentCat, nil)

	var captured *domain.Transaction
	txRepo.On("CreateAdjustment", mock.Anything, mock.MatchedBy(func(t *domain.Transaction) bool {
		return t.Type == domain.TransactionTypeIncome && t.AmountCents == 500
	})).Run(func(args mock.Arguments) {
		captured = args.Get(1).(*domain.Transaction)
	}).Return(&domain.Transaction{ID: 1, Type: domain.TransactionTypeIncome, AmountCents: 500, IsAdjustment: true}, nil)

	tx, err := svc.Apply(context.Background(), 1, 10, 500, "reconcile")
	require.NoError(t, err)
	assert.Equal(t, domain.TransactionTypeIncome, tx.Type)
	assert.Equal(t, int64(500), tx.AmountCents)
	assert.True(t, tx.IsAdjustment)
	require.NotNil(t, captured)
	assert.Equal(t, "reconcile", captured.Note)
	assert.Equal(t, int64(77), captured.CategoryID)
}

func TestAdjustmentService_Apply_NegativeDelta_CreatesExpenseTransaction(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	accounts := &mocks.MockAccountStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newAdjustmentService(txRepo, accounts, catRepo)

	acc := &domain.Account{ID: 10, UserID: 1, CurrencyCode: "EUR"}
	accounts.On("GetByID", mock.Anything, int64(10), int64(1)).Return(acc, nil)
	catRepo.On("GetSystemCategoryByType", mock.Anything, "adjustment").Return(adjustmentCat, nil)

	txRepo.On("CreateAdjustment", mock.Anything, mock.MatchedBy(func(t *domain.Transaction) bool {
		return t.Type == domain.TransactionTypeExpense && t.AmountCents == 300
	})).Return(&domain.Transaction{ID: 2, Type: domain.TransactionTypeExpense, AmountCents: 300, IsAdjustment: true}, nil)

	tx, err := svc.Apply(context.Background(), 1, 10, -300, "")
	require.NoError(t, err)
	assert.Equal(t, domain.TransactionTypeExpense, tx.Type)
	assert.Equal(t, int64(300), tx.AmountCents)
	assert.True(t, tx.IsAdjustment)
}

func TestAdjustmentService_Apply_CategoryNotFound_ReturnsError(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	accounts := &mocks.MockAccountStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newAdjustmentService(txRepo, accounts, catRepo)

	acc := &domain.Account{ID: 10, UserID: 1, CurrencyCode: "USD"}
	accounts.On("GetByID", mock.Anything, int64(10), int64(1)).Return(acc, nil)
	catRepo.On("GetSystemCategoryByType", mock.Anything, "adjustment").Return(nil, errors.New("not found"))

	_, err := svc.Apply(context.Background(), 1, 10, 100, "")
	assert.Error(t, err)
}
