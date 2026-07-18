package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_NL_MISS_SCAN=1 go test -run TestDebugNLMissScan -v
func TestDebugNLMissScan(t *testing.T) {
	if os.Getenv("LANG_NL_MISS_SCAN") == "" {
		t.Skip("set LANG_NL_MISS_SCAN=1")
	}
	runDebugMissScan(t, "nl")
}
