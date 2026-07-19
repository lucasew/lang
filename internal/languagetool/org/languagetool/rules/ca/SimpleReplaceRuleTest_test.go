package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/SimpleReplaceRuleTest.java
// Without gender/number filter; assertions use surface replacements from replace.txt.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Això està força bé."))))
	// Java ignoreTaggedWords: proper names tagged → skip. Inject POS so IsTagged().
	// Java Catalan tagger tags proper names → ignoreTaggedWords. Inject POS for all list hits.
	require.Equal(t, 0, len(rule.Match(withAnyTag("Joan Navarro no és de Navarra ni de Jerez.", "Navarro", "Navarra", "Jerez"))))

	matches := rule.Match(languagetool.AnalyzePlain("El recader fa huelga."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "ordinari", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "transportista", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "vaga", matches[1].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Aconteixements"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Esdeveniments", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Els desencontres."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "desavinences")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "desacords")

	matches = rule.Match(languagetool.AnalyzePlain("La seguent solució."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "següent")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "seient")

	matches = rule.Match(languagetool.AnalyzePlain("Un caminet poc ciclable baixa uns metres."))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "pedalable")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "ciclista")

	matches = rule.Match(languagetool.AnalyzePlain("La seva escola transformada pq les seves filles encaixen molt bé."))
	require.Equal(t, "perquè", matches[0].GetSuggestedReplacements()[0])
}

func TestSimpleReplaceRule_FailClosedUntaggedProperName(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	// Without tags, capitalized "Navarro" still matches replace list (no capital invent).
	matches := rule.Match(languagetool.AnalyzePlain("Joan Navarro no és de Navarra ni de Jerez."))
	require.GreaterOrEqual(t, len(matches), 1)
}

// withAnyTag marks surfaces with a non-empty POS so IsTagged() is true (ignoreTaggedWords).
func withAnyTag(text string, surfaces ...string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	want := map[string]bool{}
	for _, s := range surfaces {
		want[strings.ToLower(s)] = true
	}
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		if !want[strings.ToLower(tok.GetToken())] {
			continue
		}
		pos := "NP00SP0"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
	}
	return sent
}
