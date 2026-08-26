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

func TestNormalizeNGramText_NFKCAndLower(t *testing.T) {
	// Java NGramDetector.encode: NFKC + toLowerCase (keeps precomposed é)
	require.Equal(t, "café", normalizeNGramText("CAFÉ"))
	// fullwidth digit → ASCII digit via NFKC then stripped (not letter/space)
	require.Equal(t, "", normalizeNGramText("\uFF11"))
	// soft hyphen removed by letter filter; letters kept lowercased
	require.Equal(t, "foobar", normalizeNGramText("Foo\u00ADBar"))
	// spaces collapsed
	require.Equal(t, "a b", normalizeNGramText("a \t\n b"))
}
