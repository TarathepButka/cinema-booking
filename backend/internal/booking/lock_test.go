package booking

import (
	"testing"
	"time"
)

func TestLockKey(t *testing.T) {
	if got, want := lockKey("showtime-1", "A10"), "seat:showtime-1:A10"; got != want {
		t.Fatalf("lockKey() = %q, want %q", got, want)
	}
}

func TestLockExpiryUsesConfiguredTTL(t *testing.T) {
	before := time.Now().Add(lockTTL)
	expiry := LockExpiry()
	after := time.Now().Add(lockTTL)

	if expiry.Before(before) || expiry.After(after) {
		t.Fatalf("LockExpiry() = %v, want between %v and %v", expiry, before, after)
	}
}
