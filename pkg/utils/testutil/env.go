package testutil

import (
	"os"
	"testing"
)

func LoadEnv(t *testing.T, key string) string {
	t.Helper()
	value, ok := os.LookupEnv(key)
	if !ok {
		t.Skipf("Environment variable %s is not set", key)
	}
	return value
}
