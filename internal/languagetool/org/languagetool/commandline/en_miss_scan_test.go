package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_EN_MISS_SCAN=1 go test -run TestDebugENMissScan -v
func TestDebugENMissScan(t *testing.T) {
	if os.Getenv("LANG_EN_MISS_SCAN") == "" {
		t.Skip("set LANG_EN_MISS_SCAN=1")
	}
	runDebugMissScan(t, "en")
}
