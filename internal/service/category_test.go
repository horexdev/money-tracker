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

func TestCategoryService_Update_ProtectedSystemCategory(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	// UserID=0 means infrastructure category (IsSystem() == true).
	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 0, IsProtected: true}, nil)

	_, err := svc.Update(context.Background(), 99, 1, "New Name", "", "expense", "#fff")
	assert.ErrorIs(t, err, domain.ErrCategoryProtected)
}

func TestCategoryService_Update_ProtectedPersonalCategory(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	// IsProtected=true even with a valid UserID — should be blocked.
	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 99, IsProtected: true}, nil)

	_, err := svc.Update(context.Background(), 99, 1, "New Name", "", "expense", "#fff")
	assert.ErrorIs(t, err, domain.ErrCategoryProtected)
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

func TestCategoryService_Delete_ProtectedSystemCategory(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 0, IsProtected: true}, nil)

	err := svc.Delete(context.Background(), 99, 1)
	assert.ErrorIs(t, err, domain.ErrCategoryProtected)
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

func TestCategoryService_ListSorted_DefaultAsc(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	cats := []*domain.Category{
		{ID: 1, UserID: 99, Name: "Food"},
		{ID: 2, UserID: 99, Name: "Transport"},
	}
	repo.On("ListSorted", mock.Anything, int64(99), "", "asc").Return(cats, nil)

	got, err := svc.ListSorted(context.Background(), 99, "", "")
	require.NoError(t, err)
	assert.Len(t, got, 2)
	repo.AssertExpectations(t)
}

func TestCategoryService_ListSorted_DescWithTypeFilter(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	cats := []*domain.Category{{ID: 1, UserID: 99, Name: "Transport", Type: domain.CategoryTypeExpense}}
	repo.On("ListSorted", mock.Anything, int64(99), "expense", "desc").Return(cats, nil)

	got, err := svc.ListSorted(context.Background(), 99, "expense", "desc")
	require.NoError(t, err)
	assert.Len(t, got, 1)
	repo.AssertExpectations(t)
}

func TestCategoryService_ListSorted_InvalidOrder(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	_, err := svc.ListSorted(context.Background(), 99, "", "random")
	assert.ErrorIs(t, err, domain.ErrInvalidSortParam)
}

func TestCategoryService_ListSorted_InvalidType(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	_, err := svc.ListSorted(context.Background(), 99, "transfer", "asc")
	assert.ErrorIs(t, err, domain.ErrInvalidSortParam)
}

func TestCategoryService_InitDefaultForUser(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	svc := newCatService(repo)

	seeds := []domain.CategorySeed{
		{Name: "Food", Icon: "fork-knife", Type: domain.CategoryTypeExpense, Color: "#f97316"},
		{Name: "Salary", Icon: "briefcase", Type: domain.CategoryTypeIncome, Color: "#10b981"},
	}
	repo.On("BulkCreateForUser", mock.Anything, int64(42), seeds).Return(nil)

	err := svc.InitDefaultForUser(context.Background(), 42, seeds)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}
