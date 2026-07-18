package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_IT_MISS_SCAN=1 go test -run TestDebugITMissScan -v
func TestDebugITMissScan(t *testing.T) {
	if os.Getenv("LANG_IT_MISS_SCAN") == "" {
		t.Skip("set LANG_IT_MISS_SCAN=1")
	}
	runDebugMissScan(t, "it")
}
