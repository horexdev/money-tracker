package fsm

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const ttl = 30 * time.Minute

// Store manages FSM state and intermediate data in Redis.
type Store struct {
	rdb *redis.Client
}

// NewStore creates a Store backed by the given Redis client.
func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

// GetState returns the current FSM state for a user.
// Returns StateNone if no state is set.
func (s *Store) GetState(ctx context.Context, userID int64) (State, error) {
	val, err := s.rdb.GetEx(ctx, stateKey(userID), ttl).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return StateNone, nil
		}
		return StateNone, err
	}
	return State(val), nil
}

// SetState stores the FSM state for a user with a sliding TTL.
func (s *Store) SetState(ctx context.Context, userID int64, state State) error {
	return s.rdb.Set(ctx, stateKey(userID), string(state), ttl).Err()
}

// ClearState removes the FSM state key for a user.
func (s *Store) ClearState(ctx context.Context, userID int64) error {
	return s.rdb.Del(ctx, stateKey(userID)).Err()
}

// SetData stores an intermediate value in the user's FSM data.
func (s *Store) SetData(ctx context.Context, userID int64, field, value string) error {
	return s.rdb.Set(ctx, dataKey(userID, field), value, ttl).Err()
}

// GetData retrieves an intermediate value from the user's FSM data.
// Returns an empty string if the key does not exist.
func (s *Store) GetData(ctx context.Context, userID int64, field string) (string, error) {
	val, err := s.rdb.GetEx(ctx, dataKey(userID, field), ttl).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

// ClearData removes all intermediate data keys for a user.
func (s *Store) ClearData(ctx context.Context, userID int64) error {
	keys, err := s.rdb.Keys(ctx, dataPattern(userID)).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.rdb.Del(ctx, keys...).Err()
}

// Clear removes both state and all data keys for a user.
func (s *Store) Clear(ctx context.Context, userID int64) error {
	if err := s.ClearState(ctx, userID); err != nil {
		return err
	}
	return s.ClearData(ctx, userID)
}
