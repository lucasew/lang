package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_ES_MISS_SCAN=1 go test -run TestDebugESMissScan -v
func TestDebugESMissScan(t *testing.T) {
	if os.Getenv("LANG_ES_MISS_SCAN") == "" {
		t.Skip("set LANG_ES_MISS_SCAN=1")
	}
	runDebugMissScan(t, "es")
}
