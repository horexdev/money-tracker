package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// UserRepository handles persistence of User entities.
type UserRepository struct {
	q *sqlcgen.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{q: sqlcgen.New(pool)}
}

// Upsert creates a new user or updates username/name fields if already exists.
func (r *UserRepository) Upsert(ctx context.Context, u *domain.User) (*domain.User, error) {
	row, err := r.q.UpsertUser(ctx, sqlcgen.UpsertUserParams{
		ID:           u.ID,
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		CurrencyCode: u.CurrencyCode,
	})
	if err != nil {
		return nil, err
	}
	return rowToUser(row), nil
}

// GetByID returns the user with the given Telegram ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return rowToUser(row), nil
}

// UpdateCurrency changes the user's preferred currency code.
func (r *UserRepository) UpdateCurrency(ctx context.Context, id int64, code string) (*domain.User, error) {
	row, err := r.q.UpdateUserCurrency(ctx, sqlcgen.UpdateUserCurrencyParams{
		ID:           id,
		CurrencyCode: code,
	})
	if err != nil {
		return nil, err
	}
	return rowToUser(row), nil
}

// UpdateDisplayCurrencies changes the user's display currencies (comma-separated in DB).
func (r *UserRepository) UpdateDisplayCurrencies(ctx context.Context, id int64, codes string) (*domain.User, error) {
	row, err := r.q.UpdateDisplayCurrencies(ctx, sqlcgen.UpdateDisplayCurrenciesParams{
		ID:                id,
		DisplayCurrencies: codes,
	})
	if err != nil {
		return nil, err
	}
	return rowToUser(row), nil
}

// UpdateLanguage changes the user's preferred language.
func (r *UserRepository) UpdateLanguage(ctx context.Context, id int64, lang string) (*domain.User, error) {
	row, err := r.q.UpdateUserLanguage(ctx, sqlcgen.UpdateUserLanguageParams{
		ID:       id,
		Language: lang,
	})
	if err != nil {
		return nil, err
	}
	return rowToUser(row), nil
}

func rowToUser(row sqlcgen.User) *domain.User {
	var dc []string
	if row.DisplayCurrencies != "" {
		dc = strings.Split(row.DisplayCurrencies, ",")
	}
	return &domain.User{
		ID:                row.ID,
		Username:          row.Username,
		FirstName:         row.FirstName,
		LastName:          row.LastName,
		CurrencyCode:      row.CurrencyCode,
		Language:          domain.Language(row.Language),
		DisplayCurrencies: dc,
		CreatedAt:         goTime(row.CreatedAt),
		UpdatedAt:         goTime(row.UpdatedAt),
	}
}
