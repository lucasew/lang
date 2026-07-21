package morfologik

import (
	"fmt"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func newUSWrongSplitRule(words ...string) *MorfologikSpellerRule {
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	for _, w := range words {
		sp.AddWord(w)
		// Map inject has no frequency tags (Java Speller → 0). Wrong-split gates require
		// freq(sugg parts) > freq(prev); give known words a modest positive frequency
		// so join beats unknown misspellings at freq 0 (real en_US.dict has tags).
		sp.SetFrequency(w, 10)
	}
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/en/hunspell/en_US.dict", sp)
	// Java AbstractEnglishSpellerRule: ignoreWordsWithLength = 1
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	return r
}

// Twin of MorfologikAmericanSpellerRuleTest.testRuleWithWrongSplit core cases (map inject).
func TestMorfologikSpellerRule_WrongSplit(t *testing.T) {
	r := newUSWrongSplitRule(
		"thank", "you", "the", "feedback", "But", "for", "going",
		"Additionally", "LanguageTool", "offers", "spell", "checking",
		"show", "throw", "tank", "LanguageTol",
	)

	// "than kyou" -> "thank you" (sugg2)
	ms, err := r.Match(languagetool.AnalyzePlain("But than kyou for the feedback"))
	require.NoError(t, err)
	m := firstWithSuggestion(ms, "thank you")
	require.NotNil(t, m, summary(ms))
	require.Equal(t, 4, m.GetFromPos())
	require.Equal(t, 13, m.GetToPos())

	// "thanky ou" -> "thank you" (sugg1)
	ms, err = r.Match(languagetool.AnalyzePlain("But thanky ou for the feedback"))
	require.NoError(t, err)
	m = firstWithSuggestion(ms, "thank you")
	require.NotNil(t, m, summary(ms))
	require.Equal(t, 4, m.GetFromPos())
	require.Equal(t, 13, m.GetToPos())

	// "th efeedback" -> "the feedback"
	ms, err = r.Match(languagetool.AnalyzePlain("But thank you for th efeedback"))
	require.NoError(t, err)
	m = firstWithSuggestion(ms, "the feedback")
	require.NotNil(t, m, summary(ms))
	require.Equal(t, 18, m.GetFromPos())
	require.Equal(t, 30, m.GetToPos())

	// "thef eedback" -> "the feedback"
	ms, err = r.Match(languagetool.AnalyzePlain("But thank you for thef eedback"))
	require.NoError(t, err)
	require.NotNil(t, firstWithSuggestion(ms, "the feedback"), summary(ms))

	// "g oing" / "go ing" -> "going"
	// AnalyzePlain splits I'm → I ' m; ignoreWordsWithLength=1 covers I/m.
	ms, err = r.Match(languagetool.AnalyzePlain("I'm g oing"))
	require.NoError(t, err)
	m = firstWithSuggestion(ms, "going")
	require.NotNil(t, m, summary(ms))
	require.Equal(t, 4, m.GetFromPos())
	require.Equal(t, 10, m.GetToPos())

	ms, err = r.Match(languagetool.AnalyzePlain("I'm go ing"))
	require.NoError(t, err)
	m = firstWithSuggestion(ms, "going")
	require.NotNil(t, m, summary(ms))
	require.Equal(t, 4, m.GetFromPos())
	require.Equal(t, 10, m.GetToPos())

	// next-word join: "offer sspell" → "offers spell"
	// LanguageTol left misspelled (dict has LanguageTool for contrast).
	ms, err = r.Match(languagetool.AnalyzePlain("LanguageTol offer sspell checking"))
	require.NoError(t, err)
	require.NotNil(t, firstWithSuggestion(ms, "offers spell"), summary(ms))
}

func TestMorfologikSpeller_GetFrequency(t *testing.T) {
	sp := NewMorfologikSpeller("/x.dict", 1)
	require.Equal(t, 0, sp.GetFrequency("unknown"))
	sp.AddWord("table")
	// Java Speller.getFrequency without frequency tags → 0 (not invent 1)
	require.Equal(t, 0, sp.GetFrequency("table"))
	sp.SetFrequency("table", 15)
	require.Equal(t, 15, sp.GetFrequency("table"))
	require.Equal(t, 15, sp.GetFrequency("TABLE"))
}

func hasSuggestion(ms []*rules.RuleMatch, want string) bool {
	return firstWithSuggestion(ms, want) != nil
}

func firstWithSuggestion(ms []*rules.RuleMatch, want string) *rules.RuleMatch {
	for _, m := range ms {
		if m == nil {
			continue
		}
		for _, s := range m.GetSuggestedReplacements() {
			if s == want {
				return m
			}
		}
	}
	return nil
}

func summary(ms []*rules.RuleMatch) string {
	var b string
	for i, m := range ms {
		if m == nil {
			continue
		}
		b += fmt.Sprintf("[%d] %d-%d %v; ", i, m.GetFromPos(), m.GetToPos(), m.GetSuggestedReplacements())
	}
	if b == "" {
		return "(no matches)"
	}
	return b
}
