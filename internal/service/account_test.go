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
	// ExchangeService depends on *redis.Client — pass nil and avoid exchange paths in unit tests.
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

	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(10), nil)

	err := svc.Delete(context.Background(), 5, 1)
	assert.ErrorIs(t, err, domain.ErrAccountHasTransactions)
}

func TestAccountService_Delete_NotFound(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("Delete", mock.Anything, int64(5), int64(1)).Return(domain.ErrAccountNotFound)

	err := svc.Delete(context.Background(), 5, 1)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestAccountService_Delete_Success(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	svc := newAccountService(repo)

	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(0), nil)
	repo.On("Delete", mock.Anything, int64(5), int64(1)).Return(nil)

	err := svc.Delete(context.Background(), 5, 1)
	require.NoError(t, err)
}
