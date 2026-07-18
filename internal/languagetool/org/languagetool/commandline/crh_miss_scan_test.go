package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_CRH_MISS_SCAN=1 go test -run TestDebugCRHMissScan -v
func TestDebugCRHMissScan(t *testing.T) {
	if os.Getenv("LANG_CRH_MISS_SCAN") == "" {
		t.Skip("set LANG_CRH_MISS_SCAN=1")
	}
	runDebugMissScan(t, "crh")
}
