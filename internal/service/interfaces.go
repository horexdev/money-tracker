package service

import (
	"context"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// UserStorer is the repository interface consumed by UserService.
type UserStorer interface {
	Upsert(ctx context.Context, u *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateCurrency(ctx context.Context, id int64, code string) (*domain.User, error)
	UpdateDisplayCurrencies(ctx context.Context, id int64, codes string) (*domain.User, error)
	UpdateLanguage(ctx context.Context, id int64, lang string) (*domain.User, error)
}

// TransactionStorer is the repository interface consumed by TransactionService and StatsService.
type TransactionStorer interface {
	Create(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error)
	Delete(ctx context.Context, id, userID int64) error
	GetBalance(ctx context.Context, userID int64) (incomeCents, expenseCents int64, err error)
	GetBalanceByCurrency(ctx context.Context, userID int64) ([]domain.BalanceByCurrency, error)
	List(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transaction, error)
	Count(ctx context.Context, userID int64) (int64, error)
	StatsByCategory(ctx context.Context, userID int64, from, to time.Time) ([]domain.CategoryStat, error)
}

// CategoryStorer is the repository interface for category operations.
type CategoryStorer interface {
	GetByID(ctx context.Context, id int64) (*domain.Category, error)
	ListForUser(ctx context.Context, userID int64) ([]*domain.Category, error)
	ListForUserByType(ctx context.Context, userID int64, catType string) ([]*domain.Category, error)
	CreateForUser(ctx context.Context, userID int64, name, emoji, catType string) (*domain.Category, error)
	Update(ctx context.Context, userID, id int64, name, emoji, catType string) (*domain.Category, error)
	SoftDelete(ctx context.Context, id, userID int64) error
	CountTransactions(ctx context.Context, categoryID int64) (int64, error)
}

// BudgetStorer is the repository interface for budget operations.
type BudgetStorer interface {
	Create(ctx context.Context, b *domain.Budget) (*domain.Budget, error)
	GetByID(ctx context.Context, id, userID int64) (*domain.Budget, error)
	ListByUser(ctx context.Context, userID int64) ([]*domain.Budget, error)
	Update(ctx context.Context, b *domain.Budget) (*domain.Budget, error)
	Delete(ctx context.Context, id, userID int64) error
	GetByUserCategoryPeriod(ctx context.Context, userID, categoryID int64, period string) (*domain.Budget, error)
	GetSpentInPeriod(ctx context.Context, userID, categoryID int64, currency string, from, to time.Time) (int64, error)
}

// RecurringStorer is the repository interface for recurring transaction operations.
type RecurringStorer interface {
	Create(ctx context.Context, rt *domain.RecurringTransaction) (*domain.RecurringTransaction, error)
	GetByID(ctx context.Context, id, userID int64) (*domain.RecurringTransaction, error)
	ListByUser(ctx context.Context, userID int64) ([]*domain.RecurringTransaction, error)
	Update(ctx context.Context, rt *domain.RecurringTransaction) (*domain.RecurringTransaction, error)
	ToggleActive(ctx context.Context, id, userID int64) (*domain.RecurringTransaction, error)
	Delete(ctx context.Context, id, userID int64) error
	GetDue(ctx context.Context, before time.Time) ([]*domain.RecurringTransaction, error)
	UpdateNextRun(ctx context.Context, id int64, nextRun time.Time) error
}

// SavingsGoalStorer is the repository interface for savings goal operations.
type SavingsGoalStorer interface {
	Create(ctx context.Context, g *domain.SavingsGoal) (*domain.SavingsGoal, error)
	GetByID(ctx context.Context, id, userID int64) (*domain.SavingsGoal, error)
	ListByUser(ctx context.Context, userID int64) ([]*domain.SavingsGoal, error)
	Update(ctx context.Context, g *domain.SavingsGoal) (*domain.SavingsGoal, error)
	Deposit(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error)
	Withdraw(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error)
	Delete(ctx context.Context, id, userID int64) error
}
