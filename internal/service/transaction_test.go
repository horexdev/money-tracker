package service_test

import (
	"context"
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

func newTxService(txRepo *mocks.MockTransactionStorer, catRepo *mocks.MockCategoryStorer) *service.TransactionService {
	return service.NewTransactionService(txRepo, catRepo, testutil.TestLogger())
}

func TestTransactionService_AddExpense_ZeroAmount(t *testing.T) {
	svc := newTxService(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{})
	_, err := svc.AddExpense(context.Background(), 1, 0, 1, "", "USD", "USD", 1.0, nil, nil)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestTransactionService_AddExpense_NegativeAmount(t *testing.T) {
	svc := newTxService(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{})
	_, err := svc.AddExpense(context.Background(), 1, -100, 1, "", "USD", "USD", 1.0, nil, nil)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestTransactionService_AddExpense_CategoryNotFound(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTxService(txRepo, catRepo)

	catRepo.On("GetByID", mock.Anything, int64(99)).Return(nil, domain.ErrCategoryNotFound)

	_, err := svc.AddExpense(context.Background(), 1, 1000, 99, "", "USD", "USD", 1.0, nil, nil)
	assert.Error(t, err)
}

func TestTransactionService_AddExpense_WrongUser(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTxService(txRepo, catRepo)

	// Category owned by user 2, but requested by user 1.
	cat := &domain.Category{ID: 5, UserID: 2}
	catRepo.On("GetByID", mock.Anything, int64(5)).Return(cat, nil)

	_, err := svc.AddExpense(context.Background(), 1, 1000, 5, "", "USD", "USD", 1.0, nil, nil)
	assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
}

func TestTransactionService_AddExpense_SystemCategoryAllowed(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTxService(txRepo, catRepo)

	// System category: UserID == 0.
	cat := &domain.Category{ID: 5, UserID: 0, Name: "Food"}
	catRepo.On("GetByID", mock.Anything, int64(5)).Return(cat, nil)

	tx := &domain.Transaction{ID: 1, AmountCents: 1000}
	txRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(tx, nil)

	got, err := svc.AddExpense(context.Background(), 1, 1000, 5, "", "USD", "USD", 1.0, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), got.ID)
}

func TestTransactionService_AddExpense_DefaultCurrency(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTxService(txRepo, catRepo)

	cat := &domain.Category{ID: 1, UserID: 0}
	catRepo.On("GetByID", mock.Anything, int64(1)).Return(cat, nil)

	var capturedTx *domain.Transaction
	tx := &domain.Transaction{ID: 1}
	txRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).
		Run(func(args mock.Arguments) { capturedTx = args.Get(1).(*domain.Transaction) }).
		Return(tx, nil)

	_, err := svc.AddExpense(context.Background(), 1, 1000, 1, "", "", "", 1.0, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "USD", capturedTx.CurrencyCode)
}

func TestTransactionService_AddExpense_NegativeExchangeRateForcedToOne(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTxService(txRepo, catRepo)

	cat := &domain.Category{ID: 1, UserID: 0}
	catRepo.On("GetByID", mock.Anything, int64(1)).Return(cat, nil)

	var capturedTx *domain.Transaction
	tx := &domain.Transaction{ID: 1}
	txRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).
		Run(func(args mock.Arguments) { capturedTx = args.Get(1).(*domain.Transaction) }).
		Return(tx, nil)

	_, err := svc.AddExpense(context.Background(), 1, 1000, 1, "", "USD", "USD", -5.0, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 1.0, capturedTx.ExchangeRateSnapshot)
}

func TestTransactionService_AddExpense_WithCustomDate(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTxService(txRepo, catRepo)

	cat := &domain.Category{ID: 1, UserID: 0}
	catRepo.On("GetByID", mock.Anything, int64(1)).Return(cat, nil)

	tx := &domain.Transaction{ID: 1}
	txRepo.On("CreateWithDate", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(tx, nil)

	customTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	_, err := svc.AddExpense(context.Background(), 1, 1000, 1, "", "USD", "USD", 1.0, nil, &customTime)
	require.NoError(t, err)
	txRepo.AssertNotCalled(t, "Create")
	txRepo.AssertCalled(t, "CreateWithDate", mock.Anything, mock.Anything)
}

func TestTransactionService_ListPaged_PageClampedToMin(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTxService(txRepo, catRepo)

	txRepo.On("Count", mock.Anything, int64(1)).Return(int64(50), nil)
	txRepo.On("List", mock.Anything, int64(1), 20, 0).Return([]*domain.Transaction{}, nil)

	_, _, err := svc.ListPaged(context.Background(), 1, -5, 20)
	require.NoError(t, err)
	txRepo.AssertCalled(t, "List", mock.Anything, int64(1), 20, 0) // page 1 → offset 0
}

func TestTransactionService_ListPaged_PageClampedToMax(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTxService(txRepo, catRepo)

	txRepo.On("Count", mock.Anything, int64(1)).Return(int64(20), nil) // 1 page of 20
	txRepo.On("List", mock.Anything, int64(1), 20, 0).Return([]*domain.Transaction{}, nil)

	_, pages, err := svc.ListPaged(context.Background(), 1, 999, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, pages)
}

func TestTransactionService_UpdateTransaction_InvalidAmount(t *testing.T) {
	svc := newTxService(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{})
	_, err := svc.UpdateTransaction(context.Background(), 1, 1, 0, 1, "", time.Now())
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}
