package dumpcheck

// Twin of languagetool-wikipedia/src/test/java/org/languagetool/dev/dumpcheck/TatoebaSentenceSourceTest.java
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of TatoebaSentenceSourceTest.testTatoebaSource
func TestTatoebaSentenceSource_TatoebaSource(t *testing.T) {
	// Fixture lines from tatoeba-en.txt (Java @Ignore lifted with pure-Go source).
	input := strings.Join([]string{
		"73986\teng\t\"What is your wish?\" asked the little white rabbit.",
		"462830\teng\tThe mother wakes up her daughter.",
		"62498\teng\tKen beat me at chess.",
	}, "\n") + "\n"
	src := NewTatoebaSentenceSource(strings.NewReader(input))
	// Tatoeba test data has short sentences; Java still used acceptSentence.
	// Enable length filter — all three sample lines pass min length/token.
	require.True(t, src.HasNext())
	s1, err := src.Next()
	require.NoError(t, err)
	require.Equal(t, `"What is your wish?" asked the little white rabbit.`, s1.GetText())
	s2, err := src.Next()
	require.NoError(t, err)
	require.Equal(t, "The mother wakes up her daughter.", s2.GetText())
	s3, err := src.Next()
	require.NoError(t, err)
	require.Equal(t, "Ken beat me at chess.", s3.GetText())
	require.False(t, src.HasNext())
}

// Port of TatoebaSentenceSourceTest.testTatoebaSourceInvalidInput
func TestTatoebaSentenceSource_TatoebaSourceInvalidInput(t *testing.T) {
	src := NewTatoebaSentenceSource(strings.NewReader("just a text"))
	// doesn't crash; no well-formed sentences
	require.False(t, src.HasNext())
}
