package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/pkg/money"
)

// ExportService handles data export in various formats.
type ExportService struct {
	txRepo TransactionStorer
	log    *slog.Logger
}

func NewExportService(txRepo TransactionStorer, log *slog.Logger) *ExportService {
	return &ExportService{txRepo: txRepo, log: log}
}

// ExportCSV generates a CSV file with transactions for the given date range.
func (s *ExportService) ExportCSV(ctx context.Context, userID int64, from, to time.Time) ([]byte, error) {
	txs, err := s.fetchAll(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header row.
	if err := w.Write([]string{"Date", "Type", "Amount", "Currency", "Category", "Note"}); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, tx := range txs {
		record := []string{
			tx.CreatedAt.Format("2006-01-02 15:04"),
			string(tx.Type),
			money.FormatCents(tx.AmountCents),
			tx.CurrencyCode,
			tx.CategoryName,
			tx.Note,
		}
		if err := w.Write(record); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	s.log.InfoContext(ctx, "csv exported",
		slog.Int64("user_id", userID),
		slog.Int("rows", len(txs)),
	)
	return buf.Bytes(), nil
}

// fetchAll retrieves all transactions in the date range by paginating through them.
func (s *ExportService) fetchAll(ctx context.Context, userID int64, from, to time.Time) ([]*domain.Transaction, error) {
	// Use stats date range to get all transactions in that window.
	// We paginate in batches of 500 to avoid memory issues.
	const batchSize = 500
	var all []*domain.Transaction

	for offset := 0; ; offset += batchSize {
		batch, err := s.txRepo.List(ctx, userID, batchSize, offset)
		if err != nil {
			return nil, fmt.Errorf("list transactions for export: %w", err)
		}

		for _, tx := range batch {
			if tx.CreatedAt.Before(from) {
				// Transactions are ordered DESC, so once we pass the range, stop.
				return all, nil
			}
			if tx.CreatedAt.Before(to) {
				all = append(all, tx)
			}
		}

		if len(batch) < batchSize {
			break
		}
	}
	return all, nil
}
