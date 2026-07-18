package commandline

import (
	"os"
	"testing"
)

// Debug-only optional soft goldens: LANG_{LANG}_OPT_MISS_SCAN=1 go test -run TestDebugXXOptMissScan -v
func TestDebugENOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "en") }
func TestDebugDEOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "de") }
func TestDebugFROptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "fr") }
func TestDebugPTOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "pt") }
func TestDebugCAOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "ca") }
func TestDebugGAOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "ga") }
func TestDebugNLOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "nl") }
func TestDebugITOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "it") }
func TestDebugPLOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "pl") }
func TestDebugRUOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "ru") }
func TestDebugESOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "es") }
func TestDebugGLOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "gl") }
func TestDebugSVOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "sv") }
func TestDebugDAOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "da") }
func TestDebugROOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "ro") }
func TestDebugELOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "el") }
func TestDebugFAOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "fa") }
func TestDebugKMOptMissScan(t *testing.T)  { runDebugOptionalMissScan(t, "km") }

// silence unused import if os only used via helper
var _ = os.Getenv
