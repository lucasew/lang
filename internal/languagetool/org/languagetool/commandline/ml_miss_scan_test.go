package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_ML_MISS_SCAN=1 go test -run TestDebugMLMissScan -v
func TestDebugMLMissScan(t *testing.T) {
	if os.Getenv("LANG_ML_MISS_SCAN") == "" {
		t.Skip("set LANG_ML_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ml")
}
