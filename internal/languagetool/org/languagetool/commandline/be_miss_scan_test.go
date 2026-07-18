package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_BE_MISS_SCAN=1 go test -run TestDebugBEMissScan -v
func TestDebugBEMissScan(t *testing.T) {
	if os.Getenv("LANG_BE_MISS_SCAN") == "" {
		t.Skip("set LANG_BE_MISS_SCAN=1")
	}
	runDebugMissScan(t, "be")
}
