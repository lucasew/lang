package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_AST_MISS_SCAN=1 go test -run TestDebugASTMissScan -v
func TestDebugASTMissScan(t *testing.T) {
	if os.Getenv("LANG_AST_MISS_SCAN") == "" {
		t.Skip("set LANG_AST_MISS_SCAN=1")
	}
	runDebugMissScan(t, "ast")
}
