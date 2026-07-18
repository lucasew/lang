package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_BR_MISS_SCAN=1 go test -run TestDebugBRMissScan -v
func TestDebugBRMissScan(t *testing.T) {
	if os.Getenv("LANG_BR_MISS_SCAN") == "" {
		t.Skip("set LANG_BR_MISS_SCAN=1")
	}
	runDebugMissScan(t, "br")
}
