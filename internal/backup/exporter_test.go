package backup

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestExporterBuildsBackupWithLocalRefs(t *testing.T) {
	reader := &fakeSourceReader{
		users: []SourceUser{sampleSourceUser()},
		graph: sampleSourceGraph(),
		snapshots: []SourceExchangeRateSnapshot{
			{
				SnapshotDate:       "2026-07-01",
				BaseCurrencyCode:   "usd",
				TargetCurrencyCode: "eur",
				RateE8:             92000000,
				CreatedAt:          testTime(12),
			},
		},
	}
	userID := int64(987654321)

	exporter := NewExporter(reader)
	exporter.now = func() time.Time { return testTime(99) }

	backup, err := exporter.Export(context.Background(), &userID)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if reader.filter == nil || *reader.filter != userID {
		t.Fatalf("ListUsers() filter = %v, want %d", reader.filter, userID)
	}
	if len(backup.Profiles) != 1 {
		t.Fatalf("profiles len = %d, want 1", len(backup.Profiles))
	}

	profile := backup.Profiles[0]
	if profile.Ref != "profile:p000001" {
		t.Fatalf("profile ref = %q", profile.Ref)
	}
	if profile.Label != "Profile 1" {
		t.Fatalf("profile label = %q, want neutral label", profile.Label)
	}
	if got := profile.DisplayCurrencyCodes; len(got) != 2 || got[0] != "USD" || got[1] != "EUR" {
		t.Fatalf("display currencies = %#v", got)
	}
	if profile.Accounts[0].Ref != "account:a000001" || profile.Accounts[1].Ref != "account:a000002" {
		t.Fatalf("account refs = %#v", accountRefs(profile.Accounts))
	}
	if profile.Transactions[0].AccountRef != "account:a000001" {
		t.Fatalf("transaction account ref = %q", profile.Transactions[0].AccountRef)
	}
	if profile.Transactions[0].CategoryRef != "category:c000001" {
		t.Fatalf("transaction category ref = %q", profile.Transactions[0].CategoryRef)
	}
	if profile.Transfers[0].FromTransactionRef == nil || *profile.Transfers[0].FromTransactionRef != "transaction:t000001" {
		t.Fatalf("transfer from transaction ref = %#v", profile.Transfers[0].FromTransactionRef)
	}
	if profile.Transfers[0].ToTransactionRef == nil || *profile.Transfers[0].ToTransactionRef != "transaction:t000002" {
		t.Fatalf("transfer to transaction ref = %#v", profile.Transfers[0].ToTransactionRef)
	}
	if profile.SavingsGoals[0].AccountRef == nil || *profile.SavingsGoals[0].AccountRef != "account:a000002" {
		t.Fatalf("goal account ref = %#v", profile.SavingsGoals[0].AccountRef)
	}
	if profile.GoalTransactions[0].GoalRef != "goal:g000001" {
		t.Fatalf("goal transaction goal ref = %q", profile.GoalTransactions[0].GoalRef)
	}
	if profile.TransactionTemplates[0].AccountRef != "account:a000001" {
		t.Fatalf("template account ref = %q", profile.TransactionTemplates[0].AccountRef)
	}
	if backup.ExchangeRateSnapshots[0].BaseCurrencyCode != "USD" {
		t.Fatalf("snapshot base currency = %q", backup.ExchangeRateSnapshots[0].BaseCurrencyCode)
	}

	data, err := EncodeBackup(backup)
	if err != nil {
		t.Fatalf("EncodeBackup() error = %v", err)
	}
	keys := collectJSONKeys(t, data)
	for key := range keys {
		if key == "id" || strings.HasSuffix(key, "_id") {
			t.Fatalf("forbidden identifier key %q in backup JSON", key)
		}
		for _, forbidden := range forbiddenRefParts {
			if strings.Contains(key, forbidden) {
				t.Fatalf("forbidden key part %q in backup JSON key %q", forbidden, key)
			}
		}
	}
	encoded := string(data)
	for _, rawID := range []string{"987654321", "81001", "82001", "83001", "84001", "85001"} {
		if strings.Contains(encoded, rawID) {
			t.Fatalf("encoded backup leaks source id %s:\n%s", rawID, encoded)
		}
	}
}

func TestExporterFailsWhenSourceRelationshipCannotBeMapped(t *testing.T) {
	reader := &fakeSourceReader{
		users: []SourceUser{sampleSourceUser()},
		graph: sampleSourceGraph(),
	}
	reader.graph.Transfers[0].FromTxID = int64Ptr(999999)

	exporter := NewExporter(reader)
	_, err := exporter.Export(context.Background(), nil)
	if err == nil {
		t.Fatal("Export() error = nil, want missing relationship error")
	}
	if !strings.Contains(err.Error(), "transfer from transaction") {
		t.Fatalf("Export() error = %q, want transfer relationship context", err.Error())
	}
}

type fakeSourceReader struct {
	users     []SourceUser
	graph     SourceProfileGraph
	snapshots []SourceExchangeRateSnapshot
	filter    *int64
}

func (r *fakeSourceReader) ListUsers(_ context.Context, userID *int64) ([]SourceUser, error) {
	if userID != nil {
		value := *userID
		r.filter = &value
	}
	return r.users, nil
}

func (r *fakeSourceReader) LoadProfile(_ context.Context, user SourceUser) (SourceProfileGraph, error) {
	graph := r.graph
	graph.User = user
	return graph, nil
}

func (r *fakeSourceReader) ListExchangeRateSnapshots(context.Context) ([]SourceExchangeRateSnapshot, error) {
	return r.snapshots, nil
}

func sampleSourceUser() SourceUser {
	animate := true
	return SourceUser{
		ID:                       987654321,
		Language:                 "ru",
		DisplayCurrencies:        "usd, eur,USD",
		NotifyBudgetAlerts:       true,
		NotifyRecurringReminders: true,
		NotifyWeeklySummary:      false,
		NotifyGoalMilestones:     true,
		StatsChartStyle:          "donut",
		AnimateNumbers:           &animate,
		Theme:                    "dark",
		HideAmounts:              true,
		CreatedAt:                testTime(1),
		UpdatedAt:                testTime(2),
	}
}

func sampleSourceGraph() SourceProfileGraph {
	deadline := "2026-12-31"
	return SourceProfileGraph{
		Accounts: []SourceAccount{
			{
				ID:             81001,
				Name:           "Main",
				Icon:           "wallet",
				Color:          "#6366f1",
				Type:           "checking",
				CurrencyCode:   "usd",
				IsDefault:      true,
				IncludeInTotal: true,
				CreatedAt:      testTime(3),
				UpdatedAt:      testTime(4),
			},
			{
				ID:             81002,
				Name:           "Savings",
				Icon:           "piggy-bank",
				Color:          "#22c55e",
				Type:           "savings",
				CurrencyCode:   "eur",
				IsDefault:      false,
				IncludeInTotal: true,
				CreatedAt:      testTime(5),
				UpdatedAt:      testTime(6),
			},
		},
		Categories: []SourceCategory{
			{
				ID:          82001,
				Name:        "Food",
				Icon:        "fork-knife",
				Type:        "expense",
				Color:       "#ef4444",
				IsProtected: false,
				UpdatedAt:   testTime(7),
			},
			{
				ID:          82002,
				Name:        "Transfer",
				Icon:        "arrows-left-right",
				Type:        "transfer",
				Color:       "#6366f1",
				IsProtected: true,
				UpdatedAt:   testTime(8),
			},
		},
		Transactions: []SourceTransaction{
			{
				ID:           83001,
				Type:         "expense",
				AmountCents:  1250,
				CategoryID:   82001,
				AccountID:    81001,
				Note:         "Lunch",
				CurrencyCode: "usd",
				SnapshotDate: "2026-07-01",
				CreatedAt:    testTime(9),
			},
			{
				ID:           83002,
				Type:         "income",
				AmountCents:  2500,
				CategoryID:   82002,
				AccountID:    81002,
				Note:         "Transfer in",
				CurrencyCode: "eur",
				SnapshotDate: "2026-07-01",
				CreatedAt:    testTime(10),
			},
		},
		Transfers: []SourceTransfer{
			{
				ID:               84001,
				FromAccountID:    81001,
				ToAccountID:      81002,
				AmountCents:      2500,
				FromCurrencyCode: "usd",
				ToCurrencyCode:   "eur",
				ExchangeRateE8:   92000000,
				Note:             "Transfer",
				FromTxID:         int64Ptr(83001),
				ToTxID:           int64Ptr(83002),
				CreatedAt:        testTime(11),
			},
		},
		Budgets: []SourceBudget{
			{
				ID:                   85001,
				CategoryID:           82001,
				LimitCents:           50000,
				Period:               "monthly",
				CurrencyCode:         "usd",
				NotifyAtPercent:      80,
				NotificationsEnabled: true,
				LastNotifiedPercent:  50,
				LastNotifiedAt:       timePtr(testTime(13)),
				CreatedAt:            testTime(14),
				UpdatedAt:            testTime(15),
			},
		},
		RecurringTransactions: []SourceRecurringTransaction{
			{
				ID:           86001,
				AccountID:    81001,
				CategoryID:   82001,
				Type:         "expense",
				AmountCents:  90000,
				CurrencyCode: "usd",
				Note:         "Rent",
				Frequency:    "monthly",
				NextRunAt:    testTime(16),
				IsActive:     true,
				CreatedAt:    testTime(17),
				UpdatedAt:    testTime(18),
			},
		},
		SavingsGoals: []SourceSavingsGoal{
			{
				ID:           87001,
				Name:         "Trip",
				TargetCents:  200000,
				CurrentCents: 10000,
				CurrencyCode: "eur",
				DeadlineDate: &deadline,
				AccountID:    int64Ptr(81002),
				CreatedAt:    testTime(19),
				UpdatedAt:    testTime(20),
			},
		},
		GoalTransactions: []SourceGoalTransaction{
			{
				ID:          88001,
				GoalID:      87001,
				Type:        "deposit",
				AmountCents: 10000,
				CreatedAt:   testTime(21),
			},
		},
		TransactionTemplates: []SourceTransactionTemplate{
			{
				ID:           89001,
				Name:         "Lunch",
				Type:         "expense",
				AmountCents:  1250,
				AmountFixed:  true,
				CategoryID:   82001,
				AccountID:    81001,
				CurrencyCode: "usd",
				Note:         "Template",
				SortOrder:    1,
				CreatedAt:    testTime(22),
				UpdatedAt:    testTime(23),
			},
		},
	}
}

func testTime(day int) time.Time {
	return time.Date(2026, time.July, day, 12, 0, 0, 0, time.UTC)
}

func int64Ptr(value int64) *int64 {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
