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
}

// TransactionStorer is the repository interface consumed by TransactionService and StatsService.
type TransactionStorer interface {
	Create(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error)
	GetBalance(ctx context.Context, userID int64) (incomeCents, expenseCents int64, err error)
	List(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transaction, error)
	StatsByCategory(ctx context.Context, userID int64, from, to time.Time) ([]domain.CategoryStat, error)
}

// CategoryStorer is the repository interface consumed by TransactionService.
type CategoryStorer interface {
	GetByID(ctx context.Context, id int64) (*domain.Category, error)
	ListForUser(ctx context.Context, userID int64) ([]*domain.Category, error)
}
