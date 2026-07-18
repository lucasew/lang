package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_SK_MISS_SCAN=1 go test -run TestDebugSKMissScan -v
func TestDebugSKMissScan(t *testing.T) {
	if os.Getenv("LANG_SK_MISS_SCAN") == "" {
		t.Skip("set LANG_SK_MISS_SCAN=1")
	}
	runDebugMissScan(t, "sk")
}
