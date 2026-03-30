package api_test

import (
	"testing"
	"time"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBotToken = "1234567890:ABCDefghIJKLMnopQRSTuvwxyz"

func TestValidateInitData_EmptyString(t *testing.T) {
	_, err := api.ValidateInitData(testBotToken, "")
	assert.ErrorIs(t, err, api.ErrInvalidInitData)
}

func TestValidateInitData_MissingHash(t *testing.T) {
	_, err := api.ValidateInitData(testBotToken, "auth_date=1234567890&user=%7B%22id%22%3A1%7D")
	assert.ErrorIs(t, err, api.ErrInvalidInitData)
}

func TestValidateInitData_TamperedHash(t *testing.T) {
	// Build valid initData, then replace hash with garbage.
	valid := testutil.BuildValidInitData(testBotToken, 42, "Alice", time.Now())
	tampered := valid[:len(valid)-5] + "ZZZZZ"
	_, err := api.ValidateInitData(testBotToken, tampered)
	assert.ErrorIs(t, err, api.ErrInvalidInitData)
}

func TestValidateInitData_ExpiredAuthDate(t *testing.T) {
	expired := time.Now().Add(-25 * time.Hour)
	initData := testutil.BuildValidInitData(testBotToken, 42, "Alice", expired)
	_, err := api.ValidateInitData(testBotToken, initData)
	assert.ErrorIs(t, err, api.ErrInvalidInitData)
}

func TestValidateInitData_Valid(t *testing.T) {
	now := time.Now()
	initData := testutil.BuildValidInitData(testBotToken, 12345, "Alice", now)
	user, err := api.ValidateInitData(testBotToken, initData)
	require.NoError(t, err)
	assert.Equal(t, int64(12345), user.ID)
	assert.Equal(t, "Alice", user.FirstName)
}

func TestValidateInitData_WrongBotToken(t *testing.T) {
	initData := testutil.BuildValidInitData(testBotToken, 42, "Alice", time.Now())
	_, err := api.ValidateInitData("wrong:token", initData)
	assert.ErrorIs(t, err, api.ErrInvalidInitData)
}

func TestValidateInitData_InvalidQueryString(t *testing.T) {
	_, err := api.ValidateInitData(testBotToken, "%%%invalid%%%")
	assert.ErrorIs(t, err, api.ErrInvalidInitData)
}
