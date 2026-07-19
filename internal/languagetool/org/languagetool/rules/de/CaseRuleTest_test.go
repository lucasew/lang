package de

// Twin of CaseRuleTest — Java uses POS + tagger lookup (no surface invent).
import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCaseRule_Rule(t *testing.T) {
	rule := NewCaseRule(nil)
	// untagged AnalyzePlain must not invent capitalization hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Und das Neue Haus."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ich habe Heute keine Zeit."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ein einfacher Satz zum Testen."))))

	// Morph: "das" + lowercase adjective reading as noun error is complex (needs Lookup hook).
	// Smoke: matchMorph runs without panic on tagged sentence start.
	sent := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("einfacher", "ADJ:NOM:SIN:MAS:GRU:IND", "einfach"),
		atrWithPOS("Satz", "SUB:NOM:SIN:MAS", "Satz"),
		atrWithPOS(".", "PKT", "."),
	))
	require.NotPanics(t, func() { _ = rule.Match(sent) })
}

func TestCaseRuleCompareLists(t *testing.T) {
	s := languagetool.AnalyzePlain("Hier ein Test")
	toks := s.GetTokensWithoutWhitespace()
	// tokens: "", Hier, ein, Test, .
	require.True(t, CaseRuleCompareLists(toks, 1, 2,
		[]*regexp.Regexp{regexp.MustCompile("Hier"), regexp.MustCompile("ein")}))
	require.False(t, CaseRuleCompareLists(toks, 1, 2,
		[]*regexp.Regexp{regexp.MustCompile("Hier"), regexp.MustCompile("Test")}))
}

func TestCaseRule_EstimateContextAndURL(t *testing.T) {
	rule := NewCaseRule(nil)
	// Java: ANTI_PATTERNS.stream().mapToInt(List::size).max().orElse(0)
	require.Greater(t, rule.EstimateContextForSureMatch(), 0)
	require.Equal(t, "https://dict.leo.org/grammatik/deutsch/Rechtschreibung/Regeln/Gross-klein/index.html", rule.GetURL())
}

// Java: (prevTokenIsDas && …) || (i>1 && VER:AUX|MOD at i-2) continues the whole
// iteration — uppercase match is skipped even when prev is not "das".
// "Wird man Gehen nach …" with Gehen as VER:INF baseform must not flag DE_CASE.
func TestCaseRule_AuxModSkipIndependentOfDas(t *testing.T) {
	rule := NewCaseRule(nil)
	rule.IsMisspelled = func(string) bool { return false }
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("Wird", "VER:AUX:3:SIN:PRÄ:SFT", "werden"),
		atrWithPOS("man", "PRO:PER:NOM:SIN:3:MAS", "man"),
		atrWithPOS("Gehen", "VER:INF:NON", "Gehen"),
		atrWithPOS("weiter", "ADV", "weiter"),
		atrWithPOS(".", "PKT", "."),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	// matchMorph (no anti-pattern immunization) so we isolate the Java operator-precedence continue.
	require.Equal(t, 0, len(rule.matchMorph(toks, sent)), "VER:AUX at i-2 must skip uppercase path (Java precedence)")
}
