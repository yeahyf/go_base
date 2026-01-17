package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
)

var testMiniredis *miniredis.Miniredis

func setupTestRedis(t *testing.T) {
	var err error
	testMiniredis, err = miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	t.Cleanup(func() {
		testMiniredis.Close()
	})
}

func getTestRedisAddr() string {
	if testMiniredis == nil {
		return "127.0.0.1:6379"
	}
	return testMiniredis.Addr()
}
