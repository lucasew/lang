package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_RU_MISS_SCAN=1 go test -run TestDebugRUMissScan -v
func TestDebugRUMissScan(t *testing.T) {
	if os.Getenv("LANG_RU_MISS_SCAN") == "" {
		t.Skip("set LANG_RU_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ru")
}
