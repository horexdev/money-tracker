package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockSnapshotStorer is a testify mock for service.SnapshotStorer.
type MockSnapshotStorer struct {
	mock.Mock
}

func (m *MockSnapshotStorer) Upsert(ctx context.Context, date time.Time, base, target string, rate float64) error {
	args := m.Called(ctx, date, base, target, rate)
	return args.Error(0)
}

func (m *MockSnapshotStorer) GetRate(ctx context.Context, date time.Time, base, target string) (float64, error) {
	args := m.Called(ctx, date, base, target)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockSnapshotStorer) GetRateOrLatest(ctx context.Context, date time.Time, base, target string) (float64, error) {
	args := m.Called(ctx, date, base, target)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockSnapshotStorer) ListDistinctBaseCurrencies(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockSnapshotStorer) GetLatestSnapshotDate(ctx context.Context) (time.Time, error) {
	args := m.Called(ctx)
	return args.Get(0).(time.Time), args.Error(1)
}
