package synthesis

// Twin of GermanSynthesizerTest — full cases in synthesis/de (avoid import cycle).
import "testing"

func TestGermanSynthesizer_SynthesizeX(t *testing.T) {
	t.Skip("Java @Ignore")
}

func TestGermanSynthesizer_Synthesize(t *testing.T) {
	t.Log("see synthesis/de.TestGermanSynthesizer_Synthesize")
}

func TestGermanSynthesizer_SynthesizeCompounds(t *testing.T) {
	t.Log("see synthesis/de.TestGermanSynthesizer_SynthesizeCompounds")
}

func TestGermanSynthesizer_MorfologikBug(t *testing.T) {
	t.Log("see synthesis/de.TestGermanSynthesizer_MorfologikBug")
}
