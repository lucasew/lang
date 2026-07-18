package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_KM_MISS_SCAN=1 go test -run TestDebugKMMissScan -v
func TestDebugKMMissScan(t *testing.T) {
	if os.Getenv("LANG_KM_MISS_SCAN") == "" {
		t.Skip("set LANG_KM_MISS_SCAN=1")
	}
	runDebugMissScan(t, "km")
}
