package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_CA_MISS_SCAN=1 go test -run TestDebugCAMissScan -v
func TestDebugCAMissScan(t *testing.T) {
	if os.Getenv("LANG_CA_MISS_SCAN") == "" {
		t.Skip("set LANG_CA_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ca")
}
