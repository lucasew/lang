package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_DE_MISS_SCAN=1 go test -run TestDebugDEMissScan -v
func TestDebugDEMissScan(t *testing.T) {
	if os.Getenv("LANG_DE_MISS_SCAN") == "" {
		t.Skip("set LANG_DE_MISS_SCAN=1")
	}
	runDebugMissScan(t, "de")
}
