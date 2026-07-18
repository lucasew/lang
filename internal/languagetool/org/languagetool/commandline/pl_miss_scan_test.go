package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_PL_MISS_SCAN=1 go test -run TestDebugPLMissScan -v
func TestDebugPLMissScan(t *testing.T) {
	if os.Getenv("LANG_PL_MISS_SCAN") == "" {
		t.Skip("set LANG_PL_MISS_SCAN=1")
	}
	runDebugMissScan(t, "pl")
}
