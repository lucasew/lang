package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_AR_MISS_SCAN=1 go test -run TestDebugARMissScan -v
func TestDebugARMissScan(t *testing.T) {
	if os.Getenv("LANG_AR_MISS_SCAN") == "" {
		t.Skip("set LANG_AR_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ar")
}
