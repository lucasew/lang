package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_DA_MISS_SCAN=1 go test -run TestDebugDAMissScan -v
func TestDebugDAMissScan(t *testing.T) {
	if os.Getenv("LANG_DA_MISS_SCAN") == "" {
		t.Skip("set LANG_DA_MISS_SCAN=1")
	}
	runDebugMissScan(t, "da")
}
