package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_SL_MISS_SCAN=1 go test -run TestDebugSLMissScan -v
func TestDebugSLMissScan(t *testing.T) {
	if os.Getenv("LANG_SL_MISS_SCAN") == "" {
		t.Skip("set LANG_SL_MISS_SCAN=1")
	}
	runDebugMissScan(t, "sl")
}
