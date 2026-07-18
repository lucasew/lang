package commandline

import (
	"os"
	"testing"
)

// Debug-only: LANG_GL_MISS_SCAN=1 go test -run TestDebugGLMissScan -v
func TestDebugGLMissScan(t *testing.T) {
	if os.Getenv("LANG_GL_MISS_SCAN") == "" {
		t.Skip("set LANG_GL_MISS_SCAN=1")
	}
	runDebugMissScan(t, "gl")
}
