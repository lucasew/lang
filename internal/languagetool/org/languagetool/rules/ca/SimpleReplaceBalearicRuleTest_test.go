package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/SimpleReplaceBalearicRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceBalearicRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceBalearicRule(nil)

	// correct sentences
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Això està força bé."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Joan Navarro no és de Navarra ni de Jerez."))))
	// Prosper / Index: Java multiword proper names tagged NP — inject NP (no title-case invent).
	require.Equal(t, 0, len(rule.Match(withNPTag("Prosper Mérimée.", "Prosper"))))
	require.Equal(t, 0, len(rule.Match(withNPTag("Index Librorum Prohibitorum", "Index"))))

	// incorrect sentences
	matches := rule.Match(languagetool.AnalyzePlain("El calcul del telefon."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "càlcul", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "telèfon", matches[1].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("EL CALCUL DEL TELEFON."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "CÀLCUL", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "TELÈFON", matches[1].GetSuggestedReplacements()[0])
}

func TestSimpleReplaceBalearicRule_FailClosedWithoutNP(t *testing.T) {
	rule := NewSimpleReplaceBalearicRule(nil)
	// Without NP tag, Prosper is in replace_balearic.txt → match (no capital-surface invent).
	matches := rule.Match(languagetool.AnalyzePlain("Prosper Mérimée."))
	require.GreaterOrEqual(t, len(matches), 1)
	// Case preserved from surface (Java AbstractSimpleReplaceRule).
	require.Equal(t, "Pròsper", matches[0].GetSuggestedReplacements()[0])
}

func withNPTag(text, surface string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	nws := sent.GetTokensWithoutWhitespace()
	for _, tok := range nws {
		if tok == nil {
			continue
		}
		if !strings.EqualFold(tok.GetToken(), surface) {
			continue
		}
		// NP… starting with NP (Java hasPosTagStartingWith("NP"))
		pos := "NP00SP0"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
		break
	}
	return sent
}
