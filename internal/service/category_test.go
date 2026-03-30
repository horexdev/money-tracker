package service_test

import (
	"context"
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCatService(repo *mocks.MockCategoryStorer) *service.CategoryService {
	return service.NewCategoryService(repo, testutil.TestLogger())
}

func TestCategoryService_Create_EmptyName(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	_, err := svc.Create(context.Background(), 1, "", "🍔", "expense", "#fff")
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateForUser")
}

func TestCategoryService_Update_SystemCategory(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 0}, nil)

	_, err := svc.Update(context.Background(), 99, 1, "New Name", "", "expense", "#fff")
	assert.ErrorIs(t, err, domain.ErrCategorySystemReadOnly)
}

func TestCategoryService_Update_WrongUser(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 2}, nil)

	_, err := svc.Update(context.Background(), 99, 1, "Name", "", "expense", "#fff")
	assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
}

func TestCategoryService_Update_Success(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	existing := &domain.Category{ID: 1, UserID: 99, Color: "#old"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)
	updated := &domain.Category{ID: 1, UserID: 99, Name: "Groceries"}
	repo.On("Update", mock.Anything, int64(99), int64(1), "Groceries", "🛒", "expense", "#new").Return(updated, nil)

	got, err := svc.Update(context.Background(), 99, 1, "Groceries", "🛒", "expense", "#new")
	require.NoError(t, err)
	assert.Equal(t, "Groceries", got.Name)
}

func TestCategoryService_Delete_SystemCategory(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 0}, nil)

	err := svc.Delete(context.Background(), 99, 1)
	assert.ErrorIs(t, err, domain.ErrCategorySystemReadOnly)
}

func TestCategoryService_Delete_WrongUser(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 2}, nil)

	err := svc.Delete(context.Background(), 99, 1)
	assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
}

func TestCategoryService_Delete_HasTransactions(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 99}, nil)
	repo.On("CountTransactions", mock.Anything, int64(1)).Return(int64(3), nil)

	err := svc.Delete(context.Background(), 99, 1)
	assert.ErrorIs(t, err, domain.ErrCategoryInUse)
}

func TestCategoryService_Delete_NoTransactions(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 99}, nil)
	repo.On("CountTransactions", mock.Anything, int64(1)).Return(int64(0), nil)
	repo.On("SoftDelete", mock.Anything, int64(1), int64(99)).Return(nil)

	err := svc.Delete(context.Background(), 99, 1)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}
