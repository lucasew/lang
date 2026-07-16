package languagetool

import (
	"runtime"
	"testing"
)

// Port of org.languagetool.VersionTest.printVersion — diagnostic only.
func TestVersion_PrintVersion(t *testing.T) {
	// Java logs JVM/OS properties; we log Go runtime equivalents so the twin
	// exists and runs without panicking.
	t.Logf("Go version: %s", runtime.Version())
	t.Logf("OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	t.Logf("NumCPU: %d", runtime.NumCPU())
}
