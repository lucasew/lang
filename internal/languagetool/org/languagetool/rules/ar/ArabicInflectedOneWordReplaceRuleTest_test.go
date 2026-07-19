package ar

// Twin of ArabicInflectedOneWordReplaceRuleTest — lemma+POS path (no surface clitic invent).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicInflectedOneWordReplaceRule_Rule(t *testing.T) {
	rule := NewArabicInflectedOneWordReplaceRule(nil)
	// Correct (no wrong lemma)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("أجريت بحوثا في المخبر"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("وجعل لكم من أزواجكم بنين وحفدة"))))

	// Errors: inject lemma matching dictionary (Java tagger provides lemma+POS)
	// أبحاثا → lemma أبحاث if in file
	require.NotEqual(t, 0, len(rule.Match(analyzeARInflected("أجريت أبحاثا في المخبر", map[string]struct{ Lemma, POS string }{
		"أبحاثا": {Lemma: "أبحاث", POS: "N"},
	}))))

	// without POS/lemma: fail closed
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("أجريت أبحاثا في المخبر"))))
}

func TestArabicInflectedOneWordReplaceRule_FailClosedUntagged(t *testing.T) {
	rule := NewArabicInflectedOneWordReplaceRule(nil)
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("وجعل لكم من أزواجكم بنين وأحفاد")))
}

func analyzeARInflected(text string, tags map[string]struct{ Lemma, POS string }) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(text, func(tok string) []languagetool.TokenTag {
		if tg, ok := tags[tok]; ok {
			return []languagetool.TokenTag{{POS: tg.POS, Lemma: tg.Lemma}}
		}
		// try without diacritics-ish: exact only
		_ = strings.TrimSpace
		return nil
	})
}
