package detector

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCharNGramDetector(t *testing.T) {
	d := NewCharNGramDetector(3)
	d.TrainFromText("en", "the quick brown fox jumps over the lazy dog the the the")
	d.TrainFromText("de", "der die das und ist nicht mit von auf den den den und")
	scores := d.DetectLanguages("the dog and the fox")
	require.NotEmpty(t, scores)
	require.Greater(t, scores["en"], scores["de"])
}
