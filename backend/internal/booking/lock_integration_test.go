package booking

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedisLockAllowsOnlyOneConcurrentOwner(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set TEST_REDIS_ADDR to run Redis concurrency integration test")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("TEST_REDIS_PASSWORD"),
	})
	t.Cleanup(func() {
		_ = rdb.Del(context.Background(), lockKey("integration-showtime", "A1")).Err()
		_ = rdb.Close()
	})

	lockSvc := NewLockService(rdb)
	var winners atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(user int) {
			defer wg.Done()
			acquired, err := lockSvc.AcquireLock(
				context.Background(),
				"integration-showtime",
				"A1",
				fmt.Sprintf("user-%d", user),
			)
			if err != nil {
				t.Errorf("AcquireLock() error = %v", err)
				return
			}
			if acquired {
				winners.Add(1)
			}
		}(i)
	}
	wg.Wait()

	if got := winners.Load(); got != 1 {
		t.Fatalf("concurrent lock winners = %d, want 1", got)
	}
}
