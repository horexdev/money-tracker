package service_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
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

func newExportService(repo *mocks.MockTransactionStorer) *service.ExportService {
	return service.NewExportService(repo, testutil.TestLogger())
}

func readCSV(t *testing.T, data []byte) [][]string {
	t.Helper()
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	require.NoError(t, err)
	return rows
}

func TestExportService_ExportCSV_HeaderOnly_WhenNoData(t *testing.T) {
	repo := &mocks.MockTransactionStorer{}
	svc := newExportService(repo)
	repo.On("List", mock.Anything, int64(1), 500, 0).Return([]*domain.Transaction{}, nil)

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)

	out, err := svc.ExportCSV(context.Background(), 1, from, to)
	require.NoError(t, err)

	rows := readCSV(t, out)
	require.Len(t, rows, 1)
	assert.Equal(t, []string{"Date", "Type", "Amount", "Currency", "Category", "Note"}, rows[0])
}

func TestExportService_ExportCSV_WritesAllRowsInRange(t *testing.T) {
	repo := &mocks.MockTransactionStorer{}
	svc := newExportService(repo)

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	txs := []*domain.Transaction{
		{
			CreatedAt: time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC),
			Type:      domain.TransactionTypeExpense,
			AmountCents: 1500, CurrencyCode: "USD", CategoryName: "Food", Note: "lunch",
		},
		{
			CreatedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
			Type:      domain.TransactionTypeIncome,
			AmountCents: 100000, CurrencyCode: "USD", CategoryName: "Salary",
		},
	}
	repo.On("List", mock.Anything, int64(1), 500, 0).Return(txs, nil)

	out, err := svc.ExportCSV(context.Background(), 1, from, to)
	require.NoError(t, err)

	rows := readCSV(t, out)
	require.Len(t, rows, 3, "header + 2 transactions")
	assert.Equal(t, "expense", rows[1][1])
	assert.Equal(t, "15.00", rows[1][2])
	assert.Equal(t, "USD", rows[1][3])
	assert.Equal(t, "Food", rows[1][4])
	assert.Equal(t, "lunch", rows[1][5])
	assert.Equal(t, "income", rows[2][1])
	assert.Equal(t, "1000.00", rows[2][2])
}

func TestExportService_ExportCSV_StopsAtTransactionsBeforeFrom(t *testing.T) {
	repo := &mocks.MockTransactionStorer{}
	svc := newExportService(repo)

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	txs := []*domain.Transaction{
		{
			CreatedAt: time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC),
			Type:      domain.TransactionTypeExpense,
			AmountCents: 100, CurrencyCode: "USD", CategoryName: "A",
		},
		{
			CreatedAt: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
			Type:      domain.TransactionTypeExpense,
			AmountCents: 200, CurrencyCode: "USD", CategoryName: "B",
		},
		{
			CreatedAt: time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC),
			Type:      domain.TransactionTypeExpense,
			AmountCents: 300, CurrencyCode: "USD", CategoryName: "C",
		},
	}
	repo.On("List", mock.Anything, int64(1), 500, 0).Return(txs, nil)

	out, err := svc.ExportCSV(context.Background(), 1, from, to)
	require.NoError(t, err)

	rows := readCSV(t, out)
	require.Len(t, rows, 3, "header + 2 in-range rows; pre-range tx must skip")
}

func TestExportService_ExportCSV_RepoErrorPropagates(t *testing.T) {
	repo := &mocks.MockTransactionStorer{}
	svc := newExportService(repo)
	repo.On("List", mock.Anything, int64(1), 500, 0).Return(nil, errors.New("db down"))

	out, err := svc.ExportCSV(context.Background(), 1, time.Now(), time.Now())
	assert.Nil(t, out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list transactions for export")
}
