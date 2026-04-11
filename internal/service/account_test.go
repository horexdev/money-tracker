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

func newAccountService(repo *mocks.MockAccountStorer) *service.AccountService {
	return service.NewAccountService(repo, nil, testutil.TestLogger())
}

func TestAccountService_Create_FirstAccount_IsDefault(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	// No existing accounts → first account should be default.
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Account{}, nil)
	created := &domain.Account{ID: 1, IsDefault: true}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(created, nil)

	got, err := svc.Create(context.Background(), 1, "Main", "wallet", "#fff", domain.AccountTypeChecking, "USD", true)
	require.NoError(t, err)
	assert.True(t, got.IsDefault)
}

func TestAccountService_Create_SubsequentAccount_NotDefault(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	existing := []*domain.Account{{ID: 1, IsDefault: true}}
	repo.On("ListByUser", mock.Anything, int64(1)).Return(existing, nil)
	created := &domain.Account{ID: 2, IsDefault: false}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(created, nil)

	got, err := svc.Create(context.Background(), 1, "Savings", "piggy", "#fff", domain.AccountTypeSavings, "USD", true)
	require.NoError(t, err)
	assert.False(t, got.IsDefault)
}

func TestAccountService_Delete_HasTransactions(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	acc := &domain.Account{ID: 5, UserID: 1, IsDefault: false}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(acc, nil)
	repo.On("CountAccounts", mock.Anything, int64(1)).Return(int64(2), nil)
	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(10), nil)

	err := svc.Delete(context.Background(), 5, 1)
	assert.ErrorIs(t, err, domain.ErrAccountHasTransactions)
}

func TestAccountService_Delete_HasTransfers(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	acc := &domain.Account{ID: 5, UserID: 1, IsDefault: false}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(acc, nil)
	repo.On("CountAccounts", mock.Anything, int64(1)).Return(int64(2), nil)
	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("CountTransfers", mock.Anything, int64(5), int64(1)).Return(int64(3), nil)

	err := svc.Delete(context.Background(), 5, 1)
	assert.ErrorIs(t, err, domain.ErrAccountHasTransfers)
}

func TestAccountService_Delete_LastAccount(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	acc := &domain.Account{ID: 5, UserID: 1, IsDefault: true}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(acc, nil)
	repo.On("CountAccounts", mock.Anything, int64(1)).Return(int64(1), nil)

	err := svc.Delete(context.Background(), 5, 1)
	assert.ErrorIs(t, err, domain.ErrCannotDeleteLastAccount)
}

func TestAccountService_Delete_DefaultWithMultiple(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	acc := &domain.Account{ID: 5, UserID: 1, IsDefault: true}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(acc, nil)
	repo.On("CountAccounts", mock.Anything, int64(1)).Return(int64(3), nil)
	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("CountTransfers", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("CountRecurring", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)

	err := svc.Delete(context.Background(), 5, 1)
	assert.ErrorIs(t, err, domain.ErrMustSetNewDefault)
}

func TestAccountService_Delete_DefaultWithTwo_AutoPromotes(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	acc := &domain.Account{ID: 5, UserID: 1, IsDefault: true}
	other := &domain.Account{ID: 6, UserID: 1, IsDefault: false}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(acc, nil)
	repo.On("CountAccounts", mock.Anything, int64(1)).Return(int64(2), nil)
	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("CountTransfers", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("CountRecurring", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Account{acc, other}, nil)
	repo.On("SetDefault", mock.Anything, int64(6), int64(1)).Return(other, nil)
	repo.On("Delete", mock.Anything, int64(5), int64(1)).Return(nil)

	err := svc.Delete(context.Background(), 5, 1)
	require.NoError(t, err)
	repo.AssertCalled(t, "SetDefault", mock.Anything, int64(6), int64(1))
}

func TestAccountService_Delete_NotFound(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(nil, domain.ErrAccountNotFound)

	err := svc.Delete(context.Background(), 5, 1)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestAccountService_Delete_Success(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	acc := &domain.Account{ID: 5, UserID: 1, IsDefault: false}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(acc, nil)
	repo.On("CountAccounts", mock.Anything, int64(1)).Return(int64(2), nil)
	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("CountTransfers", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("CountRecurring", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("Delete", mock.Anything, int64(5), int64(1)).Return(nil)

	err := svc.Delete(context.Background(), 5, 1)
	require.NoError(t, err)
}
