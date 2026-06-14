package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const lockTTL = 5 * time.Minute

// LockService manages Redis distributed locks for seat reservations.
// Strategy: SET key value NX EX 300
// - NX: Only set if Not eXists (atomic check-and-set)
// - EX: Auto-expire after 5 minutes (safety net)
type LockService struct {
	rdb *redis.Client
}

// NewLockService creates a new LockService.
func NewLockService(rdb *redis.Client) *LockService {
	return &LockService{rdb: rdb}
}

// lockKey returns the Redis key for a seat lock.
// Format: seat:{showtimeId}:{seatLabel}
func lockKey(showtimeID, seatLabel string) string {
	return fmt.Sprintf("seat:%s:%s", showtimeID, seatLabel)
}

// AcquireLock attempts to lock a seat for the given user.
// Returns (true, nil) if lock was acquired, (false, nil) if already locked.
func (s *LockService) AcquireLock(ctx context.Context, showtimeID, seatLabel, userID string) (bool, error) {
	key := lockKey(showtimeID, seatLabel)
	// SET NX EX — atomic: only sets if key does not exist
	ok, err := s.rdb.SetNX(ctx, key, userID, lockTTL).Result()
	if err != nil {
		return false, fmt.Errorf("redis lock acquire: %w", err)
	}
	return ok, nil
}

// GetLockOwner returns the user ID that currently holds the lock.
// Returns empty string if lock does not exist or has expired.
func (s *LockService) GetLockOwner(ctx context.Context, showtimeID, seatLabel string) (string, error) {
	key := lockKey(showtimeID, seatLabel)
	val, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// ReleaseLock removes the lock. Only releases if the caller is the owner
// to prevent accidental release of another user's lock.
func (s *LockService) ReleaseLock(ctx context.Context, showtimeID, seatLabel, userID string) error {
	key := lockKey(showtimeID, seatLabel)

	// Use Lua script for atomic check-and-delete
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)
	return script.Run(ctx, s.rdb, []string{key}, userID).Err()
}

// ForceReleaseLock removes the lock regardless of owner (used by scheduler).
func (s *LockService) ForceReleaseLock(ctx context.Context, showtimeID, seatLabel string) error {
	key := lockKey(showtimeID, seatLabel)
	return s.rdb.Del(ctx, key).Err()
}

// GetLockTTL returns remaining TTL of a lock.
func (s *LockService) GetLockTTL(ctx context.Context, showtimeID, seatLabel string) (time.Duration, error) {
	key := lockKey(showtimeID, seatLabel)
	return s.rdb.TTL(ctx, key).Result()
}

// LockExpiry returns the absolute expiry time for a newly acquired lock.
func LockExpiry() time.Time {
	return time.Now().Add(lockTTL)
}
