package language

// Twin of SimpleLanguageIdentifierTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier"
	"github.com/stretchr/testify/require"
)

func TestSimpleLanguageIdentifier_Detection(t *testing.T) {
	s := identifier.NewSimpleLanguageIdentifier(1000)
	s.RegisterSpeller("en", func(w string) bool { return w != "the" && w != "cat" && w != "is" })
	s.RegisterSpeller("de", func(w string) bool { return w != "der" && w != "hund" && w != "ist" })
	d := s.Detect("the cat is the cat", nil, nil)
	require.NotNil(t, d)
	require.Equal(t, "en", d.GetDetectedLanguageCode())
	d = s.Detect("der hund ist der hund", nil, nil)
	require.NotNil(t, d)
	require.Equal(t, "de", d.GetDetectedLanguageCode())
}

func TestSimpleLanguageIdentifier_ShortTexts(t *testing.T) {
	s := identifier.NewSimpleLanguageIdentifier(1000)
	s.RegisterSpeller("en", func(w string) bool { return true }) // all misspelled
	// very short / empty
	require.Nil(t, s.Detect("", nil, nil))
	require.Nil(t, s.Detect("   ", nil, nil))
}
