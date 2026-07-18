package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_PT_MISS_SCAN=1 go test -run TestDebugPTMissScan -v
func TestDebugPTMissScan(t *testing.T) {
	if os.Getenv("LANG_PT_MISS_SCAN") == "" {
		t.Skip("set LANG_PT_MISS_SCAN=1")
	}
	runDebugMissScan(t, "pt")
}
