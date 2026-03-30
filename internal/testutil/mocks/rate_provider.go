package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockRateProvider is a testify mock for service.RateProvider.
type MockRateProvider struct {
	mock.Mock
}

func (m *MockRateProvider) FetchRates(ctx context.Context, base string) (map[string]float64, error) {
	args := m.Called(ctx, base)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}
