package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_JA_MISS_SCAN=1 go test -run TestDebugJAMissScan -v
func TestDebugJAMissScan(t *testing.T) {
	if os.Getenv("LANG_JA_MISS_SCAN") == "" {
		t.Skip("set LANG_JA_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ja")
}
