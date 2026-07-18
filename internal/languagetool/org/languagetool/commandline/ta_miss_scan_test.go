package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_TA_MISS_SCAN=1 go test -run TestDebugTAMissScan -v
func TestDebugTAMissScan(t *testing.T) {
	if os.Getenv("LANG_TA_MISS_SCAN") == "" {
		t.Skip("set LANG_TA_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ta")
}
