package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/SimpleReplaceVerbsRuleTest.java
// Conjugation path with injected Tag/Synthesize (no invent of Spanish dict).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceVerbsRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceVerbsRule(nil)
	// Java: clickeaban → strip to clickear, tag amaban, synth clicar/cliquear/hacer clic
	rule.Tag = func(words []string) []*languagetool.AnalyzedTokenReadings {
		// words[0] is "am"+desinence e.g. "amaban"
		pos := "VMII3P0" // imperfect 3pl template from amaban
		tok := words[0]
		at := languagetool.NewAnalyzedToken(tok, &pos, &tok)
		return []*languagetool.AnalyzedTokenReadings{
			languagetool.NewAnalyzedTokenReadingsAt(at, 0),
		}
	}
	rule.Synthesize = func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
		// Minimal synthesizer for VMII3P0 (imperfect 3pl) used by clickeaban twin.
		lemma := ""
		if token.GetLemma() != nil {
			lemma = *token.GetLemma()
		}
		if posTag != "VMII3P0" {
			return []string{lemma}, nil
		}
		switch lemma {
		case "hacer":
			return []string{"hacían"}, nil
		default:
			if strings.HasSuffix(lemma, "ar") {
				return []string{strings.TrimSuffix(lemma, "ar") + "aban"}, nil
			}
			return []string{lemma}, nil
		}
	}

	matches := rule.Match(languagetool.AnalyzePlain("clickeaban"))
	require.Equal(t, 1, len(matches))
	// clickear → clicar|cliquear|hacer clic → synthesized first two; third has space
	require.Equal(t, []string{"clicaban", "cliqueaban", "hacían clic"}, matches[0].GetSuggestedReplacements())
}

func TestSimpleReplaceVerbsRule_FailClosedWithoutTagger(t *testing.T) {
	rule := NewSimpleReplaceVerbsRule(nil)
	// Conjugation path needs Tag; without it no invent surface-only hits.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("clickeaban"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Puede eruptar el volcán."))))
}

func TestSimpleReplaceVerbsRule_IgnoreTagged(t *testing.T) {
	rule := NewSimpleReplaceVerbsRule(nil)
	rule.Tag = func(words []string) []*languagetool.AnalyzedTokenReadings {
		pos := "VMII3P0"
		tok := words[0]
		at := languagetool.NewAnalyzedToken(tok, &pos, &tok)
		return []*languagetool.AnalyzedTokenReadings{languagetool.NewAnalyzedTokenReadingsAt(at, 0)}
	}
	rule.Synthesize = func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
		return []string{"x"}, nil
	}
	sent := languagetool.AnalyzePlain("clickeaban")
	// Mark token as tagged → skip (setIgnoreTaggedWords)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "clickeaban" {
			pos := "VMN0000"
			tok.AddReading(languagetool.NewAnalyzedToken("clickeaban", &pos, nil), "test")
		}
	}
	require.Equal(t, 0, len(rule.Match(sent)))
}
