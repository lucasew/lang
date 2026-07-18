package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_TL_MISS_SCAN=1 go test -run TestDebugTLMissScan -v
func TestDebugTLMissScan(t *testing.T) {
	if os.Getenv("LANG_TL_MISS_SCAN") == "" {
		t.Skip("set LANG_TL_MISS_SCAN=1")
	}
	runDebugMissScan(t, "tl")
}
