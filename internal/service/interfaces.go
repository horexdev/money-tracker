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
	UpdateDisplayCurrencies(ctx context.Context, id int64, codes string) (*domain.User, error)
	UpdateLanguage(ctx context.Context, id int64, lang string) (*domain.User, error)
	UpdateNotificationPreferences(ctx context.Context, id int64, prefs domain.NotificationPrefs) (*domain.User, error)
	ResetData(ctx context.Context, userID int64) error
}

// TransactionStorer is the repository interface consumed by TransactionService and StatsService.
type TransactionStorer interface {
	Create(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error)
	CreateWithDate(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error)
	Delete(ctx context.Context, id, userID int64) error
	GetBalance(ctx context.Context, userID int64) (incomeCents, expenseCents int64, err error)
	GetBalanceByCurrency(ctx context.Context, userID int64) ([]domain.BalanceByCurrency, error)
	GetBalanceByCurrencyAndAccount(ctx context.Context, userID, accountID int64) ([]domain.BalanceByCurrency, error)
	GetTotalInBaseCurrency(ctx context.Context, userID int64, targetCurrency string) (int64, error)
	List(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transaction, error)
	ListByAccount(ctx context.Context, userID, accountID int64, limit, offset int) ([]*domain.Transaction, error)
	ListWithDateRange(ctx context.Context, userID int64, from, to *time.Time, limit, offset int) ([]*domain.Transaction, error)
	ListByAccountWithDateRange(ctx context.Context, userID, accountID int64, from, to *time.Time, limit, offset int) ([]*domain.Transaction, error)
	Count(ctx context.Context, userID int64) (int64, error)
	CountByAccount(ctx context.Context, userID, accountID int64) (int64, error)
	CountWithDateRange(ctx context.Context, userID int64, from, to *time.Time) (int64, error)
	CountByAccountWithDateRange(ctx context.Context, userID, accountID int64, from, to *time.Time) (int64, error)
	StatsByCategory(ctx context.Context, userID int64, from, to time.Time) ([]domain.CategoryStat, error)
	StatsByCategoryAndAccount(ctx context.Context, userID, accountID int64, from, to time.Time) ([]domain.CategoryStat, error)
	ListByCategoryPeriod(ctx context.Context, userID, categoryID int64, from, to time.Time) ([]*domain.Transaction, error)
	Update(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error)
	CreateAdjustment(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error)
}

// CategoryStorer is the repository interface for category operations.
type CategoryStorer interface {
	GetByID(ctx context.Context, id int64) (*domain.Category, error)
	GetByName(ctx context.Context, userID int64, name string) (*domain.Category, error)
	GetSystemCategoryByType(ctx context.Context, catType string) (*domain.Category, error)
	GetBySavingsType(ctx context.Context, userID int64) (*domain.Category, error)
	ListForUser(ctx context.Context, userID int64) ([]*domain.Category, error)
	ListForUserByType(ctx context.Context, userID int64, catType string) ([]*domain.Category, error)
	ListSorted(ctx context.Context, userID int64, catType, order string) ([]*domain.Category, error)
	HasCategories(ctx context.Context, userID int64) (bool, error)
	CreateForUser(ctx context.Context, userID int64, name, icon, catType, color string) (*domain.Category, error)
	BulkCreateForUser(ctx context.Context, userID int64, seeds []domain.CategorySeed) error
	Update(ctx context.Context, userID, id int64, name, icon, catType, color string) (*domain.Category, error)
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
	UpdateLastNotified(ctx context.Context, id int64, notifiedPercent int) error
	ListDistinctUserIDs(ctx context.Context) ([]int64, error)
}

// TransactionTemplateStorer is the repository interface for transaction template operations.
type TransactionTemplateStorer interface {
	Create(ctx context.Context, t *domain.TransactionTemplate) (*domain.TransactionTemplate, error)
	GetByID(ctx context.Context, id, userID int64) (*domain.TransactionTemplate, error)
	ListByUser(ctx context.Context, userID int64) ([]*domain.TransactionTemplate, error)
	Update(ctx context.Context, t *domain.TransactionTemplate) (*domain.TransactionTemplate, error)
	Delete(ctx context.Context, id, userID int64) error
	Reorder(ctx context.Context, userID int64, orderedIDs []int64) error
}

// TransactionAdder is the narrow interface used by TransactionTemplateService to
// create a transaction from a template. *TransactionService satisfies it.
type TransactionAdder interface {
	AddExpense(ctx context.Context, userID, amountCents, categoryID int64, note, currencyCode string, accountID int64, createdAt *time.Time) (*domain.Transaction, error)
	AddIncome(ctx context.Context, userID, amountCents, categoryID int64, note, currencyCode string, accountID int64, createdAt *time.Time) (*domain.Transaction, error)
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
	ListHistory(ctx context.Context, goalID, userID int64) ([]*domain.GoalTransaction, error)
	GetByAccountID(ctx context.Context, accountID int64) ([]*domain.SavingsGoal, error)
}

// AccountStorer is the repository interface for account operations.
type AccountStorer interface {
	Create(ctx context.Context, a *domain.Account) (*domain.Account, error)
	GetByID(ctx context.Context, id, userID int64) (*domain.Account, error)
	GetDefault(ctx context.Context, userID int64) (*domain.Account, error)
	ListByUser(ctx context.Context, userID int64) ([]*domain.Account, error)
	Update(ctx context.Context, a *domain.Account) (*domain.Account, error)
	SetDefault(ctx context.Context, id, userID int64) (*domain.Account, error)
	Delete(ctx context.Context, id, userID int64) error
	CountTransactions(ctx context.Context, accountID, userID int64) (int64, error)
	CountAccounts(ctx context.Context, userID int64) (int64, error)
	CountTransfers(ctx context.Context, accountID, userID int64) (int64, error)
	CountRecurring(ctx context.Context, accountID, userID int64) (int64, error)
	GetBalance(ctx context.Context, accountID, userID int64) (int64, error)
}

// AdminStorer is the repository interface for admin analytics queries.
type AdminStorer interface {
	ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error)
	CountUsers(ctx context.Context) (int64, error)
	CountNewUsers(ctx context.Context, from, to time.Time) (int64, error)
	CountActiveUsers(ctx context.Context, from, to time.Time) (int64, error)
	CountRetainedUsers(ctx context.Context, signupFrom, signupTo, activeFrom, activeTo time.Time) (int64, error)
	ListAllUserIDs(ctx context.Context) ([]int64, error)
}

// TransferStorer is the repository interface for transfer operations.
type TransferStorer interface {
	Create(ctx context.Context, t *domain.Transfer) (*domain.Transfer, error)
	GetByID(ctx context.Context, id, userID int64) (*domain.Transfer, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transfer, error)
	ListByAccount(ctx context.Context, userID, accountID int64) ([]*domain.Transfer, error)
	Count(ctx context.Context, userID int64) (int64, error)
	Delete(ctx context.Context, id, userID int64) error
}
