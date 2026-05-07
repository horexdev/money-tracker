package mocks

import (
	"context"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockUserStorer is a testify mock for service.UserStorer.
type MockUserStorer struct {
	mock.Mock
}

func (m *MockUserStorer) Upsert(ctx context.Context, u *domain.User) (*domain.User, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorer) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorer) UpdateDisplayCurrencies(ctx context.Context, id int64, codes string) (*domain.User, error) {
	args := m.Called(ctx, id, codes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorer) UpdateLanguage(ctx context.Context, id int64, lang string) (*domain.User, error) {
	args := m.Called(ctx, id, lang)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorer) UpdateNotificationPreferences(ctx context.Context, id int64, prefs domain.NotificationPrefs) (*domain.User, error) {
	args := m.Called(ctx, id, prefs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorer) UpdateUIPreferences(ctx context.Context, id int64, style string, animateNumbers *bool) (*domain.User, error) {
	args := m.Called(ctx, id, style, animateNumbers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorer) ResetData(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
