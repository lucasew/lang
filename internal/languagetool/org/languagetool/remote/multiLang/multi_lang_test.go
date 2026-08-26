package multiLang

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiLangCorpora(t *testing.T) {
	c := NewMultiLangCorpora("en")
	c.AddSentence("Hello world.")
	c.InjectOtherSentence("de", "Guten Tag.")
	require.Equal(t, "en", c.GetLanguage())
	require.Equal(t, 2, c.GetSentencesInText())
	require.Contains(t, c.GetText(), "Hello")
	require.Contains(t, c.GetText(), "Guten")
	inj := c.GetInjectedSentences()
	require.Len(t, inj, 1)
	require.Equal(t, "de", inj[0].GetLanguage())
	require.Equal(t, "Guten Tag.", inj[0].GetText())

	s := NewInjectedSentence("fr", "  bonjour  ")
	require.Equal(t, "bonjour", s.GetText())
	require.True(t, s.Equal(NewInjectedSentence("fr", "bonjour")))

	ev := NewMultiLanguageTextCheckEval("en")
	require.Equal(t, "en", ev.MainLanguage)
}
