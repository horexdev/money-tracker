package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
)

func newTemplateService(
	repo *mocks.MockTransactionTemplateStorer,
	txAdder *mocks.MockTransactionAdder,
	accRepo *mocks.MockAccountStorer,
) *service.TransactionTemplateService {
	return service.NewTransactionTemplateService(repo, txAdder, accRepo, testutil.TestLogger())
}

func TestTemplateService_Create_ZeroAmount(t *testing.T) {
	svc := newTemplateService(&mocks.MockTransactionTemplateStorer{}, &mocks.MockTransactionAdder{}, &mocks.MockAccountStorer{})
	_, err := svc.Create(context.Background(), &domain.TransactionTemplate{
		UserID: 1, Type: domain.TransactionTypeExpense, AmountCents: 0,
		CategoryID: 1, AccountID: 1, CurrencyCode: "USD",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestTemplateService_Create_InvalidType(t *testing.T) {
	svc := newTemplateService(&mocks.MockTransactionTemplateStorer{}, &mocks.MockTransactionAdder{}, &mocks.MockAccountStorer{})
	_, err := svc.Create(context.Background(), &domain.TransactionTemplate{
		UserID: 1, Type: "bogus", AmountCents: 100,
		CategoryID: 1, AccountID: 1, CurrencyCode: "USD",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction type")
}

func TestTemplateService_Create_InheritsCurrencyFromAccount(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	accRepo := &mocks.MockAccountStorer{}
	svc := newTemplateService(repo, &mocks.MockTransactionAdder{}, accRepo)

	accRepo.On("GetByID", mock.Anything, int64(7), int64(1)).Return(&domain.Account{ID: 7, UserID: 1, CurrencyCode: "EUR"}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.TransactionTemplate) bool {
		return t.CurrencyCode == "EUR"
	})).Return(&domain.TransactionTemplate{ID: 99, CurrencyCode: "EUR"}, nil)

	got, err := svc.Create(context.Background(), &domain.TransactionTemplate{
		UserID: 1, Type: domain.TransactionTypeExpense, AmountCents: 1000,
		CategoryID: 2, AccountID: 7,
		// CurrencyCode intentionally empty
	})
	require.NoError(t, err)
	assert.Equal(t, "EUR", got.CurrencyCode)
	accRepo.AssertExpectations(t)
}

func TestTemplateService_Apply_FixedAmount_CallsAddExpense(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	txAdder := &mocks.MockTransactionAdder{}
	svc := newTemplateService(repo, txAdder, &mocks.MockAccountStorer{})

	tpl := &domain.TransactionTemplate{
		ID: 5, UserID: 1, Type: domain.TransactionTypeExpense,
		AmountCents: 30000, AmountFixed: true,
		CategoryID: 2, AccountID: 7, CurrencyCode: "USD", Note: "coffee",
	}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(tpl, nil)

	expectedTx := &domain.Transaction{ID: 100, AmountCents: 30000}
	txAdder.On("AddExpense", mock.Anything, int64(1), int64(30000), int64(2), "coffee", "USD", int64(7), (*time.Time)(nil)).
		Return(expectedTx, nil)

	got, err := svc.Apply(context.Background(), 5, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(100), got.ID)
	txAdder.AssertExpectations(t)
}

func TestTemplateService_Apply_VariableWithOverride_UsesOverride(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	txAdder := &mocks.MockTransactionAdder{}
	svc := newTemplateService(repo, txAdder, &mocks.MockAccountStorer{})

	tpl := &domain.TransactionTemplate{
		ID: 5, UserID: 1, Type: domain.TransactionTypeExpense,
		AmountCents: 50000, AmountFixed: false,
		CategoryID: 2, AccountID: 7, CurrencyCode: "USD",
	}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(tpl, nil)

	override := int64(75000)
	txAdder.On("AddExpense", mock.Anything, int64(1), int64(75000), int64(2), "", "USD", int64(7), (*time.Time)(nil)).
		Return(&domain.Transaction{ID: 101, AmountCents: 75000}, nil)

	got, err := svc.Apply(context.Background(), 5, 1, &override)
	require.NoError(t, err)
	assert.Equal(t, int64(101), got.ID)
}

func TestTemplateService_Apply_Income_CallsAddIncome(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	txAdder := &mocks.MockTransactionAdder{}
	svc := newTemplateService(repo, txAdder, &mocks.MockAccountStorer{})

	tpl := &domain.TransactionTemplate{
		ID: 6, UserID: 1, Type: domain.TransactionTypeIncome,
		AmountCents: 1000000, AmountFixed: true,
		CategoryID: 4, AccountID: 7, CurrencyCode: "USD",
	}
	repo.On("GetByID", mock.Anything, int64(6), int64(1)).Return(tpl, nil)

	txAdder.On("AddIncome", mock.Anything, int64(1), int64(1000000), int64(4), "", "USD", int64(7), (*time.Time)(nil)).
		Return(&domain.Transaction{ID: 102}, nil)

	got, err := svc.Apply(context.Background(), 6, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(102), got.ID)
}

func TestTemplateService_Apply_NotFound(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	svc := newTemplateService(repo, &mocks.MockTransactionAdder{}, &mocks.MockAccountStorer{})

	repo.On("GetByID", mock.Anything, int64(404), int64(1)).Return(nil, domain.ErrTemplateNotFound)

	_, err := svc.Apply(context.Background(), 404, 1, nil)
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}

func TestTemplateService_Apply_OverrideZero_Rejected(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	svc := newTemplateService(repo, &mocks.MockTransactionAdder{}, &mocks.MockAccountStorer{})

	tpl := &domain.TransactionTemplate{
		ID: 5, UserID: 1, Type: domain.TransactionTypeExpense,
		AmountCents: 100, AmountFixed: false, CurrencyCode: "USD",
	}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(tpl, nil)

	zero := int64(0)
	_, err := svc.Apply(context.Background(), 5, 1, &zero)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestTemplateService_Delete_PassesThroughError(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	svc := newTemplateService(repo, &mocks.MockTransactionAdder{}, &mocks.MockAccountStorer{})

	repo.On("Delete", mock.Anything, int64(5), int64(1)).Return(domain.ErrTemplateNotFound)

	err := svc.Delete(context.Background(), 5, 1)
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}

func TestTemplateService_Reorder_DelegatesToRepo(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	svc := newTemplateService(repo, &mocks.MockTransactionAdder{}, &mocks.MockAccountStorer{})

	repo.On("Reorder", mock.Anything, int64(1), []int64{3, 1, 2}).Return(nil)

	require.NoError(t, svc.Reorder(context.Background(), 1, []int64{3, 1, 2}))
	repo.AssertExpectations(t)
}
