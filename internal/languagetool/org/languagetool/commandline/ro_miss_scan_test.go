package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_RO_MISS_SCAN=1 go test -run TestDebugROMissScan -v
func TestDebugROMissScan(t *testing.T) {
	if os.Getenv("LANG_RO_MISS_SCAN") == "" {
		t.Skip("set LANG_RO_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ro")
}
