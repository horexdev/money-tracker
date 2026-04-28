package api_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/testutil"
)

// fakeAdjustmentApplier is a testify mock for the unexported adjustmentApplier
// interface, exposed via api.AdjustmentApplier.
type fakeAdjustmentApplier struct {
	mock.Mock
}

func (f *fakeAdjustmentApplier) Apply(ctx context.Context, userID, accountID, deltaCents int64, note string) (*domain.Transaction, error) {
	args := f.Called(ctx, userID, accountID, deltaCents, note)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func newAdjustmentRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/1/adjust", bytes.NewBufferString(body))
	return r.WithContext(api.WithUserID(r.Context(), 1))
}

func TestAdjustAccountHandler_POST_Applies_PositiveDelta(t *testing.T) {
	applier := &fakeAdjustmentApplier{}
	applier.On("Apply", mock.Anything, int64(1), int64(7), int64(500), "").
		Return(&domain.Transaction{ID: 100, AmountCents: 500, AccountID: 7, Type: domain.TransactionTypeIncome, IsAdjustment: true}, nil)

	h := api.AdjustAccountHandlerForTest(applier, testutil.TestLogger(), 7)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, newAdjustmentRequest(t, `{"delta_cents":500}`))

	require.Equal(t, http.StatusCreated, w.Code)
	applier.AssertExpectations(t)
}

func TestAdjustAccountHandler_POST_Applies_NegativeDelta(t *testing.T) {
	applier := &fakeAdjustmentApplier{}
	applier.On("Apply", mock.Anything, int64(1), int64(7), int64(-300), "correction").
		Return(&domain.Transaction{ID: 101, AmountCents: 300, AccountID: 7, Type: domain.TransactionTypeExpense, IsAdjustment: true}, nil)

	h := api.AdjustAccountHandlerForTest(applier, testutil.TestLogger(), 7)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, newAdjustmentRequest(t, `{"delta_cents":-300,"note":"correction"}`))

	require.Equal(t, http.StatusCreated, w.Code)
	applier.AssertExpectations(t)
}

func TestAdjustAccountHandler_POST_400_OnInvalidJSON(t *testing.T) {
	applier := &fakeAdjustmentApplier{}
	h := api.AdjustAccountHandlerForTest(applier, testutil.TestLogger(), 7)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, newAdjustmentRequest(t, `{not-json}`))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	applier.AssertNotCalled(t, "Apply", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAdjustAccountHandler_POST_400_OnZeroDelta(t *testing.T) {
	applier := &fakeAdjustmentApplier{}
	h := api.AdjustAccountHandlerForTest(applier, testutil.TestLogger(), 7)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, newAdjustmentRequest(t, `{"delta_cents":0}`))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	applier.AssertNotCalled(t, "Apply", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAdjustAccountHandler_NonPOST_Returns405(t *testing.T) {
	applier := &fakeAdjustmentApplier{}
	h := api.AdjustAccountHandlerForTest(applier, testutil.TestLogger(), 7)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/1/adjust", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestAdjustAccountHandler_PropagatesServiceError(t *testing.T) {
	applier := &fakeAdjustmentApplier{}
	applier.On("Apply", mock.Anything, int64(1), int64(7), int64(500), "").
		Return(nil, errors.New("db down"))

	h := api.AdjustAccountHandlerForTest(applier, testutil.TestLogger(), 7)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, newAdjustmentRequest(t, `{"delta_cents":500}`))

	assert.GreaterOrEqual(t, w.Code, 400, "service error should map to a 4xx/5xx response")
}
