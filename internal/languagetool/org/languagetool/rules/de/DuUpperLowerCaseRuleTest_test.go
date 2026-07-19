package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/DuUpperLowerCaseRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDuUpperLowerCaseRule_Rule(t *testing.T) {
	rule := NewDuUpperLowerCaseRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	// Java twin samples
	require.Equal(t, 0, matchN("Aber du bist noch jung, sagt euer Vater oft."))
	require.Equal(t, 0, matchN("Aber Du bist noch jung, sagt Euer Vater oft."))
	require.Equal(t, 1, matchN("Aber Du bist noch jung, sagt euer Vater oft."))
	require.Equal(t, 1, matchN("Aber du bist noch jung, sagt Euer Vater oft."))
	require.Equal(t, 0, matchN("Könnt Ihr Euch das vorstellen???"))
	require.Equal(t, 0, matchN("Könnt ihr euch das vorstellen???"))
	// Java example from class: Dir then du
	require.Equal(t, 1, matchN("Wie geht es Dir? Bist du wieder gesund?"))
	// multi-sentence via AnalyzePlain single sentence may differ; use MatchList of two
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Wie geht es Dir?"),
		languagetool.AnalyzePlain("Bist du wieder gesund?"),
	}
	require.Equal(t, 1, len(rule.MatchList(sents)))
	require.Equal(t, "https://languagetool.org/insights/de/beitrag/duzen-grossgeschrieben/", rule.GetURL())
	require.Equal(t, -1, rule.MinToCheckParagraph())
}
