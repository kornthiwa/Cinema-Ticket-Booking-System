package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "seat_lock:"

type Manager struct {
	client *redis.Client
	ttl    time.Duration
}

func NewManager(client *redis.Client, ttlSeconds int) *Manager {
	return &Manager{client: client, ttl: time.Duration(ttlSeconds) * time.Second}
}

func (m *Manager) key(screeningID string, row, col int) string {
	return fmt.Sprintf("%s%s:%d:%d", keyPrefix, screeningID, row, col)
}

// Acquire tries to acquire a distributed lock for a seat. Returns lockID if successful, empty string if already locked.
func (m *Manager) Acquire(ctx context.Context, screeningID string, row, col int) (lockID string, err error) {
	lockID = uuid.New().String()
	k := m.key(screeningID, row, col)
	ok, err := m.client.SetNX(ctx, k, lockID, m.ttl).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return lockID, nil
}

// Release releases the lock only if it's still held by the same lockID (ownership check).
func (m *Manager) Release(ctx context.Context, screeningID string, row, col int, lockID string) error {
	k := m.key(screeningID, row, col)
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)
	return script.Run(ctx, m.client, []string{k}, lockID).Err()
}

// Extend extends TTL only if lock is still held by lockID.
func (m *Manager) Extend(ctx context.Context, screeningID string, row, col int, lockID string) (bool, error) {
	k := m.key(screeningID, row, col)
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)
	ttlMs := m.ttl.Milliseconds()
	result, err := script.Run(ctx, m.client, []string{k}, lockID, ttlMs).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// GetLockID returns current lock holder ID for the seat, or empty if not locked.
func (m *Manager) GetLockID(ctx context.Context, screeningID string, row, col int) (string, error) {
	k := m.key(screeningID, row, col)
	val, err := m.client.Get(ctx, k).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}
