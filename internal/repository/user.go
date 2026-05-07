package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// UserRepository handles persistence of User entities.
type UserRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool, q: sqlcgen.New(pool)}
}

// Upsert creates a new user or updates username/name fields if already exists.
func (r *UserRepository) Upsert(ctx context.Context, u *domain.User) (*domain.User, error) {
	row, err := r.q.UpsertUser(ctx, sqlcgen.UpsertUserParams{
		ID:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Language:  string(u.Language),
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

// UpdateNotificationPreferences saves the user's notification opt-in settings.
func (r *UserRepository) UpdateNotificationPreferences(ctx context.Context, id int64, prefs domain.NotificationPrefs) (*domain.User, error) {
	row, err := r.q.UpdateNotificationPreferences(ctx, sqlcgen.UpdateNotificationPreferencesParams{
		ID:                       id,
		NotifyBudgetAlerts:       prefs.BudgetAlerts,
		NotifyRecurringReminders: prefs.RecurringReminders,
		NotifyWeeklySummary:      prefs.WeeklySummary,
		NotifyGoalMilestones:     prefs.GoalMilestones,
	})
	if err != nil {
		return nil, err
	}
	return rowToUser(row), nil
}

// UpdateTheme changes the user's UI theme preference.
func (r *UserRepository) UpdateTheme(ctx context.Context, id int64, theme domain.ThemePref) (*domain.User, error) {
	row, err := r.q.UpdateUserTheme(ctx, sqlcgen.UpdateUserThemeParams{
		ID:    id,
		Theme: string(theme),
	})
	if err != nil {
		return nil, err
	}
	return rowToUser(row), nil
}

// UpdateHideAmounts toggles the user's privacy mode for monetary amounts.
func (r *UserRepository) UpdateHideAmounts(ctx context.Context, id int64, hide bool) (*domain.User, error) {
	row, err := r.q.UpdateUserHideAmounts(ctx, sqlcgen.UpdateUserHideAmountsParams{
		ID:          id,
		HideAmounts: hide,
	})
	if err != nil {
		return nil, err
	}
	return rowToUser(row), nil
}

// ResetData deletes all user-owned data atomically. Deletion order respects FK constraints:
// transfers → transactions → budgets → recurring → savings_goals → categories → accounts.
func (r *UserRepository) ResetData(ctx context.Context, userID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := r.q.WithTx(tx)

	if err := q.DeleteAllUserTransfers(ctx, userID); err != nil {
		return fmt.Errorf("delete transfers: %w", err)
	}
	if err := q.DeleteAllUserTransactions(ctx, userID); err != nil {
		return fmt.Errorf("delete transactions: %w", err)
	}
	if err := q.DeleteAllUserBudgets(ctx, userID); err != nil {
		return fmt.Errorf("delete budgets: %w", err)
	}
	if err := q.DeleteAllUserRecurring(ctx, userID); err != nil {
		return fmt.Errorf("delete recurring: %w", err)
	}
	if err := q.DeleteAllUserGoals(ctx, userID); err != nil {
		return fmt.Errorf("delete goals: %w", err)
	}
	if err := q.DeleteAllUserCategories(ctx, pgtype.Int8{Int64: userID, Valid: true}); err != nil {
		return fmt.Errorf("delete categories: %w", err)
	}
	if err := q.DeleteAllUserAccounts(ctx, userID); err != nil {
		return fmt.Errorf("delete accounts: %w", err)
	}
	if err := q.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func rowToUser(row sqlcgen.User) *domain.User {
	var dc []string
	if row.DisplayCurrencies != "" {
		dc = strings.Split(row.DisplayCurrencies, ",")
	}
	return &domain.User{
		ID:                       row.ID,
		Username:                 row.Username,
		FirstName:                row.FirstName,
		LastName:                 row.LastName,
		Language:                 domain.Language(row.Language),
		DisplayCurrencies:        dc,
		CreatedAt:                goTime(row.CreatedAt),
		UpdatedAt:                goTime(row.UpdatedAt),
		NotifyBudgetAlerts:       row.NotifyBudgetAlerts,
		NotifyRecurringReminders: row.NotifyRecurringReminders,
		NotifyWeeklySummary:      row.NotifyWeeklySummary,
		NotifyGoalMilestones:     row.NotifyGoalMilestones,
		Theme:                    domain.ThemePref(row.Theme),
		HideAmounts:              row.HideAmounts,
	}
}
