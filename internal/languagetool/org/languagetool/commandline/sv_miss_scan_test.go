package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_SV_MISS_SCAN=1 go test -run TestDebugSVMissScan -v
func TestDebugSVMissScan(t *testing.T) {
	if os.Getenv("LANG_SV_MISS_SCAN") == "" {
		t.Skip("set LANG_SV_MISS_SCAN=1")
	}
	runDebugMissScan(t, "sv")
}
