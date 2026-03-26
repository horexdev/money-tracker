package fsm_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/fsm"
)

func newTestStore(t *testing.T) *fsm.Store {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	return fsm.NewStore(rdb)
}

func TestStore_StateLifecycle(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	const userID int64 = 42

	// default is StateNone
	state, err := store.GetState(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, fsm.StateNone, state)

	// set a state
	require.NoError(t, store.SetState(ctx, userID, fsm.StateExpenseWaitAmount))
	state, err = store.GetState(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, fsm.StateExpenseWaitAmount, state)

	// clear state
	require.NoError(t, store.ClearState(ctx, userID))
	state, err = store.GetState(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, fsm.StateNone, state)
}

func TestStore_DataLifecycle(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	const userID int64 = 99

	// missing key returns empty string
	val, err := store.GetData(ctx, userID, "amount")
	require.NoError(t, err)
	assert.Equal(t, "", val)

	// set and retrieve
	require.NoError(t, store.SetData(ctx, userID, "amount", "1250"))
	require.NoError(t, store.SetData(ctx, userID, "category", "3"))

	val, err = store.GetData(ctx, userID, "amount")
	require.NoError(t, err)
	assert.Equal(t, "1250", val)

	// clear all data
	require.NoError(t, store.ClearData(ctx, userID))
	val, err = store.GetData(ctx, userID, "amount")
	require.NoError(t, err)
	assert.Equal(t, "", val)
}

func TestStore_Clear(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	const userID int64 = 7

	require.NoError(t, store.SetState(ctx, userID, fsm.StateIncomeWaitNote))
	require.NoError(t, store.SetData(ctx, userID, "amount", "500"))

	require.NoError(t, store.Clear(ctx, userID))

	state, err := store.GetState(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, fsm.StateNone, state)

	val, err := store.GetData(ctx, userID, "amount")
	require.NoError(t, err)
	assert.Equal(t, "", val)
}
