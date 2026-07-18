package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_FR_MISS_SCAN=1 go test -run TestDebugFRMissScan -v
func TestDebugFRMissScan(t *testing.T) {
	if os.Getenv("LANG_FR_MISS_SCAN") == "" {
		t.Skip("set LANG_FR_MISS_SCAN=1")
	}
	runDebugMissScan(t, "fr")
}
