package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_EO_MISS_SCAN=1 go test -run TestDebugEOMissScan -v
func TestDebugEOMissScan(t *testing.T) {
	if os.Getenv("LANG_EO_MISS_SCAN") == "" {
		t.Skip("set LANG_EO_MISS_SCAN=1")
	}
	runDebugMissScan(t, "eo")
}
