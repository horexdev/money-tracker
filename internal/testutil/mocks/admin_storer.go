package mocks

import (
	"context"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockAdminStorer is a testify mock for service.AdminStorer.
type MockAdminStorer struct {
	mock.Mock
}

func (m *MockAdminStorer) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockAdminStorer) CountUsers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAdminStorer) CountNewUsers(ctx context.Context, from, to time.Time) (int64, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAdminStorer) CountActiveUsers(ctx context.Context, from, to time.Time) (int64, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAdminStorer) CountRetainedUsers(ctx context.Context, signupFrom, signupTo, activeFrom, activeTo time.Time) (int64, error) {
	args := m.Called(ctx, signupFrom, signupTo, activeFrom, activeTo)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAdminStorer) ListAllUserIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}
