package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/SpanishWordRepeatRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpanishWordRepeatRule_Rule(t *testing.T) {
	rule := NewSpanishWordRepeatRule(map[string]string{"repetition": "Repetición"})
	// Java JLanguageTool + Spanish tagger/disambig assigns _allow_repeat.
	// AnalyzePlain has no tagger — inject for twin good cases.
	require.Equal(t, 0, len(rule.Match(withAllowRepeat("Bienvenido/a a LanguageTool."))))
	require.Equal(t, 0, len(rule.Match(withAllowRepeat("HUCHA-GANGA.ES es la web de referencia."))))
	// real error still fires (no invent surface skip)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Esto es es un error."))))
}

func TestSpanishWordRepeatRule_FailClosedWithoutTags(t *testing.T) {
	rule := NewSpanishWordRepeatRule(map[string]string{"repetition": "Repetición"})
	// Without _allow_repeat, "a a" after slash is a repetition (no surface invent).
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Bienvenido/a a LanguageTool."))))
}

// withAllowRepeat injects _allow_repeat on the second token of the first equal-fold pair.
func withAllowRepeat(s string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(s)
	nws := sent.GetTokensWithoutWhitespace()
	for i := 2; i < len(nws); i++ {
		if nws[i] == nil || nws[i-1] == nil {
			continue
		}
		if !strings.EqualFold(nws[i-1].GetToken(), nws[i].GetToken()) {
			continue
		}
		tag := "_allow_repeat"
		nws[i].AddReading(languagetool.NewAnalyzedToken(nws[i].GetToken(), &tag, nil), "test")
		return sent
	}
	return sent
}
