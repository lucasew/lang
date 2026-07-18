package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_EL_MISS_SCAN=1 go test -run TestDebugELMissScan -v
func TestDebugELMissScan(t *testing.T) {
	if os.Getenv("LANG_EL_MISS_SCAN") == "" {
		t.Skip("set LANG_EL_MISS_SCAN=1")
	}
	runDebugMissScan(t, "el")
}
