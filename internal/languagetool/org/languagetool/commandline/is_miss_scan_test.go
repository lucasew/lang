package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_IS_MISS_SCAN=1 go test -run TestDebugISMissScan -v
func TestDebugISMissScan(t *testing.T) {
	if os.Getenv("LANG_IS_MISS_SCAN") == "" {
		t.Skip("set LANG_IS_MISS_SCAN=1")
	}
	runDebugMissScan(t, "is")
}
