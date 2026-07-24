package ca

// Twin of EndOfParagraphPunctuationRuleTest using core PunctuationMarkAtParagraphEnd.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestEndOfParagraphPunctuationRule_Rule(t *testing.T) {
	rule := rules.NewPunctuationMarkAtParagraphEnd(map[string]string{
		"punctuation_mark_paragraph_end_msg": "Add a punctuation mark at paragraph end",
	})
	// Single sentence paragraph — Java with onlyOneSentence=true may skip; core differs.
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Això és un paràgraf amb una frase només"),
	})))
	// Two sentences in one paragraph, last without final punct
	sents := languagetool.SplitAndAnalyze("Això és un paràgraf amb una frase només. Això és la segona frase")
	matches := rule.MatchList(sents)
	// Expect at least one match when last sentence lacks terminal punctuation
	require.NotEmpty(t, matches)
}
