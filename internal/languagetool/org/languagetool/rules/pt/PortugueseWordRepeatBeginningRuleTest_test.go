package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseWordRepeatBeginningRuleTest.java
//
// Java uses JLanguageTool(Portuguese) + disableAllRulesExcept PORTUGUESE_WORD_REPEAT_BEGINNING_RULE.
// This twin exercises PortugueseWordRepeatBeginningRule alone via MatchList +
// SplitAndAnalyze (rule-only outcomes; same counts as Java for this textlevel rule).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func ptWRBMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning":           "Successive sentences beginning with the same word",
		"desc_repetition_beginning_adv":       "Três frases sucessivas começam com o mesmo advérbio.",
		"desc_repetition_beginning_word":      "Três frases sucessivas começam com a mesma palavra.",
		"desc_repetition_beginning_thesaurus": "Considere usar um dicionário de sinónimos.",
		"category_style":                      "Style",
	}
}

func ptWRBAtrSurface(tok string) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(tok, nil, nil))
}

// Java PortugueseWordRepeatBeginningRuleTest.testRule — full match-count matrix (0/0/0/1/1/0).
func TestPortugueseWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewPortugueseWordRepeatBeginningRule(ptWRBMessages())

	// correct sentences:
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Este exemplo está correto. Este exemplo também está."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("2011: Setembro já passou. 2011: Outubro também já passou. 2011: Novembro já se foi."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Certo, isto está bem. Este exemplo está correto. Certo que este também."))))
	// errors:
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Este exemplo está correto. Este segundo também. Este terceiro exemplo não."))))
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Então, este está correto. Então, este está errado, por causa da repetição."))))
	// this used to cause false alarms because reset() was not implemented
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Então, este deve ser considerado uma nova frase."))))
}

// Metadata from Java ctor: id, STYLE category (PT override), ITS Style, example pair.
func TestPortugueseWordRepeatBeginningRule_Metadata(t *testing.T) {
	r := NewPortugueseWordRepeatBeginningRule(ptWRBMessages())
	require.Equal(t, "PORTUGUESE_WORD_REPEAT_BEGINNING_RULE", r.GetID())
	require.Equal(t, 2, r.MinToCheckParagraph())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.NotNil(t, r.GetCategory())
	// Java: super.setCategory(Categories.STYLE.getCategory(messages)) — not REPETITIONS_STYLE
	require.Equal(t, rules.NewCategoryId("STYLE"), r.GetCategory().GetID())
	require.Equal(t, ptWRBMessages()["desc_repetition_beginning"], r.GetDescription())

	inc := r.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t,
		"Além disso, a rua é quase completamente residêncial. <marker>Além</marker> disso, foi chamada em nome de um poeta.",
		inc[0].GetExample(),
	)
	require.Equal(t, []string{"Foi"}, inc[0].GetCorrections())
	require.Len(t, r.GetCorrectExamples(), 1)
	require.Equal(t,
		"Além disso, a rua é quase completamente residêncial. <marker>Foi</marker> chamada em nome de um poeta.",
		r.GetCorrectExamples()[0].GetExample(),
	)
}

// Java ADVERBS set — sample positives, case-sensitive negatives, set size 129.
func TestPortugueseWordRepeatBeginningRule_AdverbList(t *testing.T) {
	r := NewPortugueseWordRepeatBeginningRule(ptWRBMessages())

	// Sample surfaces from Java ADVERBS (isAdverb true)
	for _, w := range []string{"Então", "Além", "Certo", "Abaixo", "Unicamente", "E", "Já", "Não"} {
		require.True(t, r.isAdverb(ptWRBAtrSurface(w)), "isAdverb(%q)", w)
	}
	// Not in set / case-sensitive (Java HashSet.contains on getToken surface)
	require.False(t, r.isAdverb(ptWRBAtrSurface("então")))
	require.False(t, r.isAdverb(ptWRBAtrSurface("ENTÃO")))
	require.False(t, r.isAdverb(ptWRBAtrSurface("random")))
	require.False(t, r.isAdverb(ptWRBAtrSurface("Este")))
	require.False(t, r.isAdverb(nil))
	// set size matches Java Arrays.asList count (129)
	require.Equal(t, 129, len(portugueseWordRepeatBeginningAdverbs))
}

// Word vs adverb shortMsg arms from core WordRepeatBeginningRule.match (PT isAdverb).
func TestPortugueseWordRepeatBeginningRule_WordVsAdverbShortMsg(t *testing.T) {
	rule := NewPortugueseWordRepeatBeginningRule(ptWRBMessages())

	// three successive "Este" → word short message (not adverb)
	ms := rule.MatchList(languagetool.SplitAndAnalyze("Este exemplo está correto. Este segundo também. Este terceiro exemplo não."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, ptWRBMessages()["desc_repetition_beginning_word"], ms[0].GetShortMessage())
	require.Contains(t, ms[0].GetMessage(), ptWRBMessages()["desc_repetition_beginning_thesaurus"])
	require.Equal(t, 4, ms[0].GetToPos()-ms[0].GetFromPos()) // "Este"

	// two successive "Então" → adverb short message (second sentence)
	ms = rule.MatchList(languagetool.SplitAndAnalyze("Então, este está correto. Então, este está errado, por causa da repetição."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, ptWRBMessages()["desc_repetition_beginning_adv"], ms[0].GetShortMessage())
	require.Contains(t, ms[0].GetMessage(), ptWRBMessages()["desc_repetition_beginning_thesaurus"])
	// "Então" UTF-16 length (BMP: ã is one code unit in UTF-16... wait ã is U+00E3 → 1 UTF-16 unit)
	// Java String.length of "Então" = 5
	require.Equal(t, 5, ms[0].GetToPos()-ms[0].GetFromPos())
}
