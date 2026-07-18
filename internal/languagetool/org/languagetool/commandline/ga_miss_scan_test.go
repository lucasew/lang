package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_GA_MISS_SCAN=1 go test -run TestDebugGAMissScan -v
func TestDebugGAMissScan(t *testing.T) {
	if os.Getenv("LANG_GA_MISS_SCAN") == "" {
		t.Skip("set LANG_GA_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ga")
}
