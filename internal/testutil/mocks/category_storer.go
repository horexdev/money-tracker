package mocks

import (
	"context"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockCategoryStorer is a testify mock for service.CategoryStorer.
type MockCategoryStorer struct {
	mock.Mock
}

func (m *MockCategoryStorer) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryStorer) ListForUser(ctx context.Context, userID int64) ([]*domain.Category, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryStorer) ListForUserByType(ctx context.Context, userID int64, catType string) ([]*domain.Category, error) {
	args := m.Called(ctx, userID, catType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryStorer) CreateForUser(ctx context.Context, userID int64, name, emoji, catType, color string) (*domain.Category, error) {
	args := m.Called(ctx, userID, name, emoji, catType, color)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryStorer) Update(ctx context.Context, userID, id int64, name, emoji, catType, color string) (*domain.Category, error) {
	args := m.Called(ctx, userID, id, name, emoji, catType, color)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryStorer) SoftDelete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockCategoryStorer) CountTransactions(ctx context.Context, categoryID int64) (int64, error) {
	args := m.Called(ctx, categoryID)
	return args.Get(0).(int64), args.Error(1)
}
