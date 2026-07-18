package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_FA_MISS_SCAN=1 go test -run TestDebugFAMissScan -v
func TestDebugFAMissScan(t *testing.T) {
	if os.Getenv("LANG_FA_MISS_SCAN") == "" {
		t.Skip("set LANG_FA_MISS_SCAN=1")
	}
	runDebugMissScan(t, "fa")
}
