package service_test

import (
	"context"
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTransferService(transfers *mocks.MockTransferStorer, accounts *mocks.MockAccountStorer, goals *mocks.MockSavingsGoalStorer, txRepo *mocks.MockTransactionStorer, catRepo *mocks.MockCategoryStorer) *service.TransferService {
	return service.NewTransferService(transfers, accounts, goals, txRepo, catRepo, testutil.TestLogger())
}

var transferCat = &domain.Category{ID: 99, Name: "Transfer"}

func setupTransferMocks(catRepo *mocks.MockCategoryStorer, txRepo *mocks.MockTransactionStorer, userID, fromAccID, toAccID, fromTxID, toTxID int64) {
	catRepo.On("GetByName", mock.Anything, int64(0), "transfer").Return(transferCat, nil)
	txRepo.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.Transaction) bool {
		return t.Type == domain.TransactionTypeExpense
	})).Return(&domain.Transaction{ID: fromTxID}, nil)
	txRepo.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.Transaction) bool {
		return t.Type == domain.TransactionTypeIncome
	})).Return(&domain.Transaction{ID: toTxID}, nil)
}

func TestTransferService_Execute_SameAccount(t *testing.T) {
	svc := newTransferService(&mocks.MockTransferStorer{}, &mocks.MockAccountStorer{}, &mocks.MockSavingsGoalStorer{}, &mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{})
	_, err := svc.Execute(context.Background(), 1, 5, 5, 1000, 1.0, "")
	assert.ErrorIs(t, err, domain.ErrTransferSameAccount)
}

func TestTransferService_Execute_ZeroAmount(t *testing.T) {
	svc := newTransferService(&mocks.MockTransferStorer{}, &mocks.MockAccountStorer{}, &mocks.MockSavingsGoalStorer{}, &mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{})
	_, err := svc.Execute(context.Background(), 1, 1, 2, 0, 1.0, "")
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestTransferService_Execute_FromAccountNotFound(t *testing.T) {
	accounts := &mocks.MockAccountStorer{}
	svc := newTransferService(&mocks.MockTransferStorer{}, accounts, &mocks.MockSavingsGoalStorer{}, &mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{})

	accounts.On("GetByID", mock.Anything, int64(1), int64(99)).Return(nil, domain.ErrAccountNotFound)

	_, err := svc.Execute(context.Background(), 99, 1, 2, 1000, 1.0, "")
	assert.Error(t, err)
}

func TestTransferService_Execute_NegativeExchangeRateForcedToOne(t *testing.T) {
	transfers := &mocks.MockTransferStorer{}
	accounts := &mocks.MockAccountStorer{}
	goals := &mocks.MockSavingsGoalStorer{}
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTransferService(transfers, accounts, goals, txRepo, catRepo)

	fromAcc := &domain.Account{ID: 1, CurrencyCode: "USD"}
	toAcc := &domain.Account{ID: 2, CurrencyCode: "EUR"}
	accounts.On("GetByID", mock.Anything, int64(1), int64(99)).Return(fromAcc, nil)
	accounts.On("GetByID", mock.Anything, int64(2), int64(99)).Return(toAcc, nil)
	setupTransferMocks(catRepo, txRepo, 99, 1, 2, 10, 11)

	var capturedTransfer *domain.Transfer
	transfers.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transfer")).
		Run(func(args mock.Arguments) { capturedTransfer = args.Get(1).(*domain.Transfer) }).
		Return(&domain.Transfer{ID: 10}, nil)
	goals.On("GetByAccountID", mock.Anything, int64(2)).Return([]*domain.SavingsGoal{}, nil)

	_, err := svc.Execute(context.Background(), 99, 1, 2, 1000, -5.0, "")
	require.NoError(t, err)
	assert.Equal(t, 1.0, capturedTransfer.ExchangeRate)
}

func TestTransferService_Execute_AutoIncrementsLinkedGoal(t *testing.T) {
	transfers := &mocks.MockTransferStorer{}
	accounts := &mocks.MockAccountStorer{}
	goals := &mocks.MockSavingsGoalStorer{}
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	svc := newTransferService(transfers, accounts, goals, txRepo, catRepo)

	fromAcc := &domain.Account{ID: 1, CurrencyCode: "USD"}
	toAcc := &domain.Account{ID: 2, CurrencyCode: "USD"}
	accounts.On("GetByID", mock.Anything, int64(1), int64(99)).Return(fromAcc, nil)
	accounts.On("GetByID", mock.Anything, int64(2), int64(99)).Return(toAcc, nil)
	setupTransferMocks(catRepo, txRepo, 99, 1, 2, 10, 11)
	transfers.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transfer")).Return(&domain.Transfer{ID: 1}, nil)

	linkedGoal := &domain.SavingsGoal{ID: 7, UserID: 99}
	goals.On("GetByAccountID", mock.Anything, int64(2)).Return([]*domain.SavingsGoal{linkedGoal}, nil)
	goals.On("Deposit", mock.Anything, int64(7), int64(99), int64(1000)).Return(linkedGoal, nil)

	_, err := svc.Execute(context.Background(), 99, 1, 2, 1000, 1.0, "")
	require.NoError(t, err)
	goals.AssertCalled(t, "Deposit", mock.Anything, int64(7), int64(99), int64(1000))
}
