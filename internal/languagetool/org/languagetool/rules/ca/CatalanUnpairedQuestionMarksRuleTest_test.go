package ca

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestCatalanUnpairedQuestionMarksRule_Basic(t *testing.T) {
	rule := NewCatalanUnpairedQuestionMarksRule(nil)
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, "CA_UNPAIRED_QUESTION", rule.GetID())

	assertN := func(s string, n int, sugg string) {
		t.Helper()
		matches := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)})
		require.Equal(t, n, len(matches), "text %q got %v", s, formatCAQM(matches))
		if n >= 1 && sugg != "" {
			require.Equal(t, sugg, matches[0].GetSuggestedReplacements()[0], "text %q", s)
		}
	}
	assertNTagged := func(s string, tags map[string]string, n int, sugg string) {
		t.Helper()
		sent := languagetool.AnalyzePlain(s)
		injectCAFreeLing(sent, tags)
		matches := rule.MatchList([]*languagetool.AnalyzedSentence{sent})
		require.Equal(t, n, len(matches), "text %q got %v", s, formatCAQM(matches))
		if n >= 1 && sugg != "" {
			require.Equal(t, sugg, matches[0].GetSuggestedReplacements()[0], "text %q", s)
		}
	}

	assertN("¿Com estàs?", 0, "")
	assertN("Com estàs?", 1, "¿Com")
	// POS: què after comma
	assertNTagged("Hola, què vols?", map[string]string{"què": "PT000000"}, 1, "¿què")
	// de + què → SPS00 + PT
	assertNTagged("Hola, de què parles?", map[string]string{"de": "SPS00", "què": "PT000000"}, 1, "¿de")
	// surface no/oi/eh after comma (Java)
	assertN("Tens raó, oi?", 1, "¿oi")
	assertN("Tens raó, no?", 1, "¿no")
	assertN("Tens raó, eh?", 1, "¿eh")
}

func TestCatalanUnpairedQuestionMarksRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewCatalanUnpairedQuestionMarksRule(nil)
	// Without PT tag, firstToken stays sentence-initial content word (no invent).
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Hola, què vols?")})
	require.Equal(t, 1, len(matches))
	require.Equal(t, "¿Hola", matches[0].GetSuggestedReplacements()[0])
}

func TestCatalanUnpairedExclamationMarksRule(t *testing.T) {
	rule := NewCatalanUnpairedExclamationMarksRule(nil)
	require.Equal(t, "CA_UNPAIRED_EXCLAMATION", rule.GetID())
	require.True(t, rule.IsDefaultOff())
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Quina ràbia!")})
	require.Equal(t, 1, len(matches))
	require.Equal(t, "¡Quina", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("¡Quina ràbia!")})))
}

func injectCAFreeLing(sent *languagetool.AnalyzedSentence, tags map[string]string) {
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		for surface, pos := range tags {
			if !strings.EqualFold(tok.GetToken(), surface) {
				continue
			}
			p := pos
			tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &p, nil), "test")
		}
	}
}

func formatCAQM(matches []*rules.RuleMatch) string {
	if len(matches) == 0 {
		return "[]"
	}
	var b strings.Builder
	for i, m := range matches {
		if i > 0 {
			b.WriteString("; ")
		}
		if len(m.GetSuggestedReplacements()) > 0 {
			b.WriteString(m.GetSuggestedReplacements()[0])
		}
	}
	return b.String()
}
