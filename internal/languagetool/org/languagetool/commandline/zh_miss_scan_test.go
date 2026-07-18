package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_ZH_MISS_SCAN=1 go test -run TestDebugZHMissScan -v
func TestDebugZHMissScan(t *testing.T) {
	if os.Getenv("LANG_ZH_MISS_SCAN") == "" {
		t.Skip("set LANG_ZH_MISS_SCAN=1")
	}
	runDebugMissScan(t, "zh")
}
