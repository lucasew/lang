package pt

// Twin of org.languagetool.rules.pt.PortugueseWrongWordInContextRule (no dedicated Java *Test;
// metadata + Java example pair + official wrongWordInContext.txt outcomes).
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START); this rule matches surface tokens
// and context regexes without needing POS tags (MatchLemmas=false).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Java PortugueseWrongWordInContextRule: id, description, SEMANTICS+Grammar, messages, example.
func TestPortugueseWrongWordInContextRule_Metadata(t *testing.T) {
	rule := NewPortugueseWrongWordInContextRule(nil)
	require.Equal(t, "PORTUGUESE_WRONG_WORD_IN_CONTEXT", rule.GetID())
	require.Equal(t, "Confusão de palavra dentro do contexto (p.ex. infligir/infringir, etc.)", rule.GetDescription())
	// Final category after Java ctor override (not CONFUSED_WORDS from base).
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.CategorySemantics, rule.GetCategory().GetID())
	require.Equal(t, "SEMANTICS", rule.GetCategory().GetName()) // nil messages → id as name
	require.Equal(t, rules.ITSGrammar, rule.GetLocQualityIssueType())
	require.Equal(t, "pt", rule.LanguageCode)
	require.Equal(t, PortugueseWrongWordInContextFilePath, "/pt/wrongWordInContext.txt")
	require.Equal(t, "Pretende dizer <suggestion>$SUGGESTION</suggestion> em vez de $WRONGWORD?", rule.MessageString)
	require.Equal(t, "Possível confusão de termos. Verifique.", rule.ShortMessageString)
	require.Equal(t, "Considere <suggestion>$SUGGESTION</suggestion>, i.e. $EXPLANATION_SUGGESTION, em vez de '$WRONGWORD', i.e. $EXPLANATION_WRONGWORD?", rule.LongMessageString)

	// Official /pt/wrongWordInContext.txt (embed twin of Java resource).
	require.NotEmpty(t, rule.Entries)
	require.GreaterOrEqual(t, len(rule.Entries), 30) // official file has 31 active pairs

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "O acidente <marker>infringiu</marker> grandes danos.", inc[0].GetExample())
	require.Equal(t, []string{"infligiu"}, inc[0].GetCorrections())
	cor := rule.GetCorrectExamples()
	require.Len(t, cor, 1)
	require.Equal(t, "O acidente <marker>infligiu</marker> grandes danos.", cor[0].GetExample())
}

func TestPortugueseWrongWordInContextRule_Rule(t *testing.T) {
	rule := NewPortugueseWrongWordInContextRule(nil)
	assertGood := func(sentence string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(sentence))), "good: %q", sentence)
	}
	assertBad := func(sentence string) {
		t.Helper()
		require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain(sentence))), "bad: %q", sentence)
	}

	// Java example pair: infligir/infringir (context danos → infligiu)
	assertBad("O acidente infringiu grandes danos.")
	assertGood("O acidente infligiu grandes danos.")
	matches := rule.Match(languagetool.AnalyzePlain("O acidente infringiu grandes danos."))
	require.Len(t, matches, 1)
	require.Equal(t, "infligiu", matches[0].GetSuggestedReplacements()[0])
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "inflig")
	require.Equal(t, strings.Index("O acidente infringiu grandes danos.", "infringiu"), matches[0].GetFromPos())
	require.Equal(t, matches[0].GetFromPos()+len("infringiu"), matches[0].GetToPos())
	idRule, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "Java always wraps match in SpecificIdRule")
	require.Equal(t, tools.ToId("PORTUGUESE_WRONG_WORD_IN_CONTEXT_infringiu_infligiu", "pt"), idRule.GetID())
	// Description has no $match placeholder in PT (unlike EN); SpecificIdRule gets unreplaced desc.
	require.Equal(t, rule.GetDescription(), idRule.GetDescription())
	// Long message (explanations present in data): aplicar / transgredir
	require.Contains(t, matches[0].GetMessage(), "infligiu")
	require.Contains(t, matches[0].GetMessage(), "infringiu")
	require.Contains(t, matches[0].GetMessage(), "aplicar")
	require.Contains(t, matches[0].GetMessage(), "transgredir")
	require.Equal(t, "Possível confusão de termos. Verifique.", matches[0].GetShortMessage())
	// SpecificIdRule inherits final SEMANTICS + Grammar from rule meta
	require.NotNil(t, idRule.Category)
	require.Equal(t, rules.CategorySemantics, idRule.Category.GetID())
	require.Equal(t, rules.ITSGrammar, idRule.IssueType)

	// Second pair from official data: tráfego/tráfico (context drogas → tráfico)
	assertBad("O tráfego de drogas aumenta.")
	assertGood("O tráfego rodoviário aumenta.")
	assertGood("O tráfico de drogas aumenta.")
	traf := rule.Match(languagetool.AnalyzePlain("O tráfego de drogas aumenta."))
	require.Len(t, traf, 1)
	require.Equal(t, "tráfico", traf[0].GetSuggestedReplacements()[0])
	require.Contains(t, traf[0].GetMessage(), "negócio ilícito")
	require.Contains(t, traf[0].GetMessage(), "trânsito")
}
