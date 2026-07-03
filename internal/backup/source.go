package backup

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type SourceReader interface {
	ListUsers(ctx context.Context, userID *int64) ([]SourceUser, error)
	LoadProfile(ctx context.Context, user SourceUser) (SourceProfileGraph, error)
	ListExchangeRateSnapshots(ctx context.Context) ([]SourceExchangeRateSnapshot, error)
}

type SourceUser struct {
	ID                       int64
	Language                 string
	DisplayCurrencies        string
	NotifyBudgetAlerts       bool
	NotifyRecurringReminders bool
	NotifyWeeklySummary      bool
	NotifyGoalMilestones     bool
	StatsChartStyle          string
	AnimateNumbers           *bool
	Theme                    string
	HideAmounts              bool
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type SourceProfileGraph struct {
	User                  SourceUser
	Accounts              []SourceAccount
	Categories            []SourceCategory
	Transactions          []SourceTransaction
	Transfers             []SourceTransfer
	Budgets               []SourceBudget
	RecurringTransactions []SourceRecurringTransaction
	SavingsGoals          []SourceSavingsGoal
	GoalTransactions      []SourceGoalTransaction
	TransactionTemplates  []SourceTransactionTemplate
}

type SourceAccount struct {
	ID             int64
	Name           string
	Icon           string
	Color          string
	Type           string
	CurrencyCode   string
	IsDefault      bool
	IncludeInTotal bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SourceCategory struct {
	ID          int64
	Name        string
	Icon        string
	Type        string
	Color       string
	IsProtected bool
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type SourceTransaction struct {
	ID           int64
	Type         string
	AmountCents  int64
	CategoryID   int64
	AccountID    int64
	Note         string
	CurrencyCode string
	SnapshotDate string
	CreatedAt    time.Time
	IsAdjustment bool
}

type SourceTransfer struct {
	ID               int64
	FromAccountID    int64
	ToAccountID      int64
	AmountCents      int64
	FromCurrencyCode string
	ToCurrencyCode   string
	ExchangeRateE8   int64
	Note             string
	FromTxID         *int64
	ToTxID           *int64
	CreatedAt        time.Time
}

type SourceBudget struct {
	ID                   int64
	CategoryID           int64
	LimitCents           int64
	Period               string
	CurrencyCode         string
	NotifyAtPercent      int
	NotificationsEnabled bool
	LastNotifiedPercent  int
	LastNotifiedAt       *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type SourceRecurringTransaction struct {
	ID           int64
	AccountID    int64
	CategoryID   int64
	Type         string
	AmountCents  int64
	CurrencyCode string
	Note         string
	Frequency    string
	NextRunAt    time.Time
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SourceSavingsGoal struct {
	ID           int64
	Name         string
	TargetCents  int64
	CurrentCents int64
	CurrencyCode string
	DeadlineDate *string
	AccountID    *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SourceGoalTransaction struct {
	ID          int64
	GoalID      int64
	Type        string
	AmountCents int64
	CreatedAt   time.Time
}

type SourceExchangeRateSnapshot struct {
	SnapshotDate       string
	BaseCurrencyCode   string
	TargetCurrencyCode string
	RateE8             int64
	CreatedAt          time.Time
}

type SourceTransactionTemplate struct {
	ID           int64
	Name         string
	Type         string
	AmountCents  int64
	AmountFixed  bool
	CategoryID   int64
	AccountID    int64
	CurrencyCode string
	Note         string
	SortOrder    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Exporter struct {
	reader SourceReader
	now    func() time.Time
}

func NewExporter(reader SourceReader) *Exporter {
	return &Exporter{
		reader: reader,
		now:    time.Now,
	}
}

func (e *Exporter) Export(ctx context.Context, userID *int64) (MoneyTrackerBackup, error) {
	createdAt := epochMillis(e.now())
	backup := NewEmptyBackup(createdAt)

	users, err := e.reader.ListUsers(ctx, userID)
	if err != nil {
		return backup, fmt.Errorf("list export users: %w", err)
	}
	backup.Profiles = make([]BackupProfile, 0, len(users))
	for idx, user := range users {
		graph, err := e.reader.LoadProfile(ctx, user)
		if err != nil {
			return backup, fmt.Errorf("load profile %d: %w", idx+1, err)
		}
		profile, err := buildProfile(idx+1, graph)
		if err != nil {
			return backup, fmt.Errorf("build profile %d: %w", idx+1, err)
		}
		backup.Profiles = append(backup.Profiles, profile)
	}

	snapshots, err := e.reader.ListExchangeRateSnapshots(ctx)
	if err != nil {
		return backup, fmt.Errorf("list exchange rate snapshots: %w", err)
	}
	backup.ExchangeRateSnapshots = make([]BackupExchangeRateSnapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		backup.ExchangeRateSnapshots = append(backup.ExchangeRateSnapshots, BackupExchangeRateSnapshot{
			SnapshotDate:         snapshot.SnapshotDate,
			BaseCurrencyCode:     normalizeCurrency(snapshot.BaseCurrencyCode),
			TargetCurrencyCode:   normalizeCurrency(snapshot.TargetCurrencyCode),
			RateE8:               snapshot.RateE8,
			CreatedAtEpochMillis: epochMillis(snapshot.CreatedAt),
		})
	}

	if err := ValidateBackup(backup); err != nil {
		return backup, err
	}
	return backup, nil
}

func buildProfile(index int, graph SourceProfileGraph) (BackupProfile, error) {
	refs := newProfileRefs(index)
	accountRefsByID := make(map[int64]string, len(graph.Accounts))
	categoryRefsByID := make(map[int64]string, len(graph.Categories))
	transactionRefsByID := make(map[int64]string, len(graph.Transactions))
	goalRefsByID := make(map[int64]string, len(graph.SavingsGoals))
	statsChartStyle := normalizedDefault(graph.User.StatsChartStyle, "donut")
	if err := requireOneOf("stats chart style", statsChartStyle, "donut", "stacked_bar", "dual_bar", "profit_bars"); err != nil {
		return BackupProfile{}, err
	}
	theme := normalizedDefault(graph.User.Theme, "system")
	if err := requireOneOf("theme", theme, "system", "light", "dark"); err != nil {
		return BackupProfile{}, err
	}

	profile := BackupProfile{
		Ref:                  refs.nextProfile(),
		Label:                fmt.Sprintf("Profile %d", index),
		LanguageCode:         normalizedDefault(graph.User.Language, "en"),
		DisplayCurrencyCodes: parseDisplayCurrencies(graph.User.DisplayCurrencies),
		NotificationPreferences: BackupNotificationPreferences{
			NotifyBudgetAlerts:       graph.User.NotifyBudgetAlerts,
			NotifyRecurringReminders: graph.User.NotifyRecurringReminders,
			NotifyWeeklySummary:      graph.User.NotifyWeeklySummary,
			NotifyGoalMilestones:     graph.User.NotifyGoalMilestones,
		},
		UIPreferences: BackupUIPreferences{
			StatsChartStyle: statsChartStyle,
			AnimateNumbers:  graph.User.AnimateNumbers,
			Theme:           theme,
			HideAmounts:     graph.User.HideAmounts,
		},
		CreatedAtEpochMillis:  epochMillis(graph.User.CreatedAt),
		UpdatedAtEpochMillis:  epochMillis(graph.User.UpdatedAt),
		Accounts:              []BackupAccount{},
		Categories:            []BackupCategory{},
		Transactions:          []BackupTransaction{},
		Transfers:             []BackupTransfer{},
		Budgets:               []BackupBudget{},
		RecurringTransactions: []BackupRecurringTransaction{},
		SavingsGoals:          []BackupSavingsGoal{},
		GoalTransactions:      []BackupGoalTransaction{},
		ExchangeRateOverrides: []BackupExchangeRateOverride{},
		TransactionTemplates:  []BackupTransactionTemplate{},
	}

	for _, source := range graph.Accounts {
		if err := requireOneOf("account type", source.Type, "checking", "savings", "cash", "credit", "crypto"); err != nil {
			return BackupProfile{}, err
		}
		ref := refs.nextAccount()
		accountRefsByID[source.ID] = ref
		profile.Accounts = append(profile.Accounts, BackupAccount{
			Ref:                  ref,
			Name:                 source.Name,
			Icon:                 normalizedDefault(source.Icon, "wallet"),
			Color:                normalizedDefault(source.Color, "#6366f1"),
			Type:                 source.Type,
			CurrencyCode:         normalizeCurrency(source.CurrencyCode),
			IsDefault:            source.IsDefault,
			IncludeInTotal:       source.IncludeInTotal,
			CreatedAtEpochMillis: epochMillis(source.CreatedAt),
			UpdatedAtEpochMillis: epochMillis(source.UpdatedAt),
		})
	}

	for _, source := range graph.Categories {
		if err := requireOneOf("category type", source.Type, "expense", "income", "both", "savings", "transfer", "adjustment"); err != nil {
			return BackupProfile{}, err
		}
		ref := refs.nextCategory()
		categoryRefsByID[source.ID] = ref
		profile.Categories = append(profile.Categories, BackupCategory{
			Ref:                  ref,
			Name:                 source.Name,
			Icon:                 normalizedDefault(source.Icon, "tag"),
			Type:                 source.Type,
			Color:                normalizedDefault(source.Color, "#6366f1"),
			IsProtected:          source.IsProtected,
			UpdatedAtEpochMillis: epochMillis(source.UpdatedAt),
			DeletedAtEpochMillis: optionalEpochMillis(source.DeletedAt),
		})
	}

	for _, source := range graph.Transactions {
		if err := requireOneOf("transaction type", source.Type, "expense", "income"); err != nil {
			return BackupProfile{}, err
		}
		accountRef, err := lookupRef(accountRefsByID, source.AccountID, "transaction account")
		if err != nil {
			return BackupProfile{}, err
		}
		categoryRef, err := lookupRef(categoryRefsByID, source.CategoryID, "transaction category")
		if err != nil {
			return BackupProfile{}, err
		}
		ref := refs.nextTransaction()
		transactionRefsByID[source.ID] = ref
		profile.Transactions = append(profile.Transactions, BackupTransaction{
			Ref:                  ref,
			Type:                 source.Type,
			AmountCents:          source.AmountCents,
			CategoryRef:          categoryRef,
			AccountRef:           accountRef,
			Note:                 source.Note,
			CurrencyCode:         normalizeCurrency(source.CurrencyCode),
			SnapshotDate:         source.SnapshotDate,
			CreatedAtEpochMillis: epochMillis(source.CreatedAt),
			IsAdjustment:         source.IsAdjustment,
		})
	}

	for _, source := range graph.Transfers {
		fromAccountRef, err := lookupRef(accountRefsByID, source.FromAccountID, "transfer from account")
		if err != nil {
			return BackupProfile{}, err
		}
		toAccountRef, err := lookupRef(accountRefsByID, source.ToAccountID, "transfer to account")
		if err != nil {
			return BackupProfile{}, err
		}
		fromTransactionRef, err := optionalLookupRef(transactionRefsByID, source.FromTxID, "transfer from transaction")
		if err != nil {
			return BackupProfile{}, err
		}
		toTransactionRef, err := optionalLookupRef(transactionRefsByID, source.ToTxID, "transfer to transaction")
		if err != nil {
			return BackupProfile{}, err
		}
		profile.Transfers = append(profile.Transfers, BackupTransfer{
			Ref:                  refs.nextTransfer(),
			FromAccountRef:       fromAccountRef,
			ToAccountRef:         toAccountRef,
			AmountCents:          source.AmountCents,
			FromCurrencyCode:     normalizeCurrency(source.FromCurrencyCode),
			ToCurrencyCode:       normalizeCurrency(source.ToCurrencyCode),
			ExchangeRateE8:       source.ExchangeRateE8,
			Note:                 source.Note,
			FromTransactionRef:   fromTransactionRef,
			ToTransactionRef:     toTransactionRef,
			CreatedAtEpochMillis: epochMillis(source.CreatedAt),
		})
	}

	for _, source := range graph.Budgets {
		if err := requireOneOf("budget period", source.Period, "weekly", "monthly"); err != nil {
			return BackupProfile{}, err
		}
		categoryRef, err := lookupRef(categoryRefsByID, source.CategoryID, "budget category")
		if err != nil {
			return BackupProfile{}, err
		}
		profile.Budgets = append(profile.Budgets, BackupBudget{
			Ref:                       refs.nextBudget(),
			CategoryRef:               categoryRef,
			LimitCents:                source.LimitCents,
			Period:                    source.Period,
			CurrencyCode:              normalizeCurrency(source.CurrencyCode),
			NotifyAtPercent:           source.NotifyAtPercent,
			NotificationsEnabled:      source.NotificationsEnabled,
			LastNotifiedPercent:       source.LastNotifiedPercent,
			LastNotifiedAtEpochMillis: optionalEpochMillis(source.LastNotifiedAt),
			CreatedAtEpochMillis:      epochMillis(source.CreatedAt),
			UpdatedAtEpochMillis:      epochMillis(source.UpdatedAt),
		})
	}

	for _, source := range graph.RecurringTransactions {
		if err := requireOneOf("recurring transaction type", source.Type, "expense", "income"); err != nil {
			return BackupProfile{}, err
		}
		if err := requireOneOf("recurring frequency", source.Frequency, "daily", "weekly", "monthly", "yearly"); err != nil {
			return BackupProfile{}, err
		}
		accountRef, err := lookupRef(accountRefsByID, source.AccountID, "recurring account")
		if err != nil {
			return BackupProfile{}, err
		}
		categoryRef, err := lookupRef(categoryRefsByID, source.CategoryID, "recurring category")
		if err != nil {
			return BackupProfile{}, err
		}
		profile.RecurringTransactions = append(profile.RecurringTransactions, BackupRecurringTransaction{
			Ref:                  refs.nextRecurring(),
			AccountRef:           accountRef,
			CategoryRef:          categoryRef,
			Type:                 source.Type,
			AmountCents:          source.AmountCents,
			CurrencyCode:         normalizeCurrency(source.CurrencyCode),
			Note:                 source.Note,
			Frequency:            source.Frequency,
			NextRunAtEpochMillis: epochMillis(source.NextRunAt),
			IsActive:             source.IsActive,
			CreatedAtEpochMillis: epochMillis(source.CreatedAt),
			UpdatedAtEpochMillis: epochMillis(source.UpdatedAt),
		})
	}

	for _, source := range graph.SavingsGoals {
		accountRef, err := optionalLookupRef(accountRefsByID, source.AccountID, "savings goal account")
		if err != nil {
			return BackupProfile{}, err
		}
		ref := refs.nextGoal()
		goalRefsByID[source.ID] = ref
		profile.SavingsGoals = append(profile.SavingsGoals, BackupSavingsGoal{
			Ref:                  ref,
			Name:                 source.Name,
			TargetCents:          source.TargetCents,
			CurrentCents:         source.CurrentCents,
			CurrencyCode:         normalizeCurrency(source.CurrencyCode),
			DeadlineDate:         source.DeadlineDate,
			AccountRef:           accountRef,
			CreatedAtEpochMillis: epochMillis(source.CreatedAt),
			UpdatedAtEpochMillis: epochMillis(source.UpdatedAt),
		})
	}

	for _, source := range graph.GoalTransactions {
		if err := requireOneOf("goal transaction type", source.Type, "deposit", "withdraw"); err != nil {
			return BackupProfile{}, err
		}
		goalRef, err := lookupRef(goalRefsByID, source.GoalID, "goal transaction goal")
		if err != nil {
			return BackupProfile{}, err
		}
		profile.GoalTransactions = append(profile.GoalTransactions, BackupGoalTransaction{
			Ref:                  refs.nextGoalTransaction(),
			GoalRef:              goalRef,
			Type:                 source.Type,
			AmountCents:          source.AmountCents,
			CreatedAtEpochMillis: epochMillis(source.CreatedAt),
		})
	}

	for _, source := range graph.TransactionTemplates {
		if err := requireOneOf("transaction template type", source.Type, "expense", "income"); err != nil {
			return BackupProfile{}, err
		}
		accountRef, err := lookupRef(accountRefsByID, source.AccountID, "transaction template account")
		if err != nil {
			return BackupProfile{}, err
		}
		categoryRef, err := lookupRef(categoryRefsByID, source.CategoryID, "transaction template category")
		if err != nil {
			return BackupProfile{}, err
		}
		profile.TransactionTemplates = append(profile.TransactionTemplates, BackupTransactionTemplate{
			Ref:                  refs.nextTemplate(),
			Name:                 source.Name,
			Type:                 source.Type,
			AmountCents:          source.AmountCents,
			AmountFixed:          source.AmountFixed,
			CategoryRef:          categoryRef,
			AccountRef:           accountRef,
			CurrencyCode:         normalizeCurrency(source.CurrencyCode),
			Note:                 source.Note,
			SortOrder:            source.SortOrder,
			CreatedAtEpochMillis: epochMillis(source.CreatedAt),
			UpdatedAtEpochMillis: epochMillis(source.UpdatedAt),
		})
	}

	return profile, nil
}

func lookupRef(refs map[int64]string, id int64, scope string) (string, error) {
	ref, ok := refs[id]
	if !ok {
		return "", fmt.Errorf("missing %s ref for db row", scope)
	}
	return ref, nil
}

func optionalLookupRef(refs map[int64]string, id *int64, scope string) (*string, error) {
	if id == nil {
		return nil, nil
	}
	ref, err := lookupRef(refs, *id, scope)
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func epochMillis(t time.Time) int64 {
	return t.UTC().UnixNano() / int64(time.Millisecond)
}

func optionalEpochMillis(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	v := epochMillis(*t)
	return &v
}

func parseDisplayCurrencies(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	seen := map[string]struct{}{}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		code := normalizeCurrency(part)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	return result
}

func normalizeCurrency(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizedDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func requireOneOf(scope, value string, allowed ...string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return fmt.Errorf("unsupported %s %q", scope, value)
}

type profileRefsGenerator struct {
	profileIndex    int
	account         int
	category        int
	transaction     int
	transfer        int
	budget          int
	recurring       int
	goal            int
	goalTransaction int
	template        int
}

func newProfileRefs(profileIndex int) *profileRefsGenerator {
	return &profileRefsGenerator{profileIndex: profileIndex}
}

func (r *profileRefsGenerator) nextProfile() string {
	return fmt.Sprintf("profile:p%06d", r.profileIndex)
}

func (r *profileRefsGenerator) nextAccount() string {
	r.account++
	return fmt.Sprintf("account:a%06d", r.account)
}

func (r *profileRefsGenerator) nextCategory() string {
	r.category++
	return fmt.Sprintf("category:c%06d", r.category)
}

func (r *profileRefsGenerator) nextTransaction() string {
	r.transaction++
	return fmt.Sprintf("transaction:t%06d", r.transaction)
}

func (r *profileRefsGenerator) nextTransfer() string {
	r.transfer++
	return fmt.Sprintf("transfer:tr%06d", r.transfer)
}

func (r *profileRefsGenerator) nextBudget() string {
	r.budget++
	return fmt.Sprintf("budget:b%06d", r.budget)
}

func (r *profileRefsGenerator) nextRecurring() string {
	r.recurring++
	return fmt.Sprintf("recurring:r%06d", r.recurring)
}

func (r *profileRefsGenerator) nextGoal() string {
	r.goal++
	return fmt.Sprintf("goal:g%06d", r.goal)
}

func (r *profileRefsGenerator) nextGoalTransaction() string {
	r.goalTransaction++
	return fmt.Sprintf("goal-transaction:gt%06d", r.goalTransaction)
}

func (r *profileRefsGenerator) nextTemplate() string {
	r.template++
	return fmt.Sprintf("template:tp%06d", r.template)
}
