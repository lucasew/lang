package synthesis

// Twin of GermanSynthesizerTest — implementation lives in synthesis/de.
// Full INSTANCE twins: de.TestGermanSynthesizer_Synthesize / Compounds / MorfologikBug.
import (
	"testing"
)

func TestGermanSynthesizer_PackagePointer(t *testing.T) {
	// Package-level name kept for twin audit discovery; real tests are in de/.
	_ = GermanSynthesizerPackage
}
