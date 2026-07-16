package detector

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNGramDetectorScripts(t *testing.T) {
	d := NewNGramDetector(500)
	d.TrainFromText("en", "the cat sat on the mat hello world")
	d.TrainFromText("de", "der hund liegt auf der matte hallo welt")
	require.Equal(t, "en", d.TopLanguage("hello the world the cat"))
	// Chinese characters boost zh
	scores := d.DetectLanguages("你好世界")
	require.Contains(t, scores, "zh")
	require.Greater(t, scores["zh"], 0.0)
}
