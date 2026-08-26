package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicInflectedOneWordReplaceRuleTest.java
//
// Java setup:
//   rule = new ArabicInflectedOneWordReplaceRule(TestTools.getEnglishMessages());
//   lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
// Go: NewArabicInflectedOneWordReplaceRule(nil) + NewJLanguageTool("ar") with
// real ArabicTagger.Tag as SentenceTagger (lemma+POS path; no invent tags).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	arsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	artag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/stretchr/testify/require"
)

// wireARInflectedLT ports JLanguageTool(Languages.getLanguageForShortCode("ar")).
// Returns nil skip when arabic.dict is unavailable.
func wireARInflectedLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	if artag.DiscoverArabicPOSDict() == "" {
		t.Skip("arabic.dict not in tree")
	}
	if arsynth.DiscoverArabicSynthDict() == "" {
		t.Skip("arabic_synth.dict not in tree")
	}
	artag.EnsureDefaultArabicTagger()
	require.NotNil(t, artag.DefaultArabicTagger)
	require.NotNil(t, artag.DefaultArabicTagger.GetWordTagger())
	// Ensure synthesizer Instance can load for InflectLemmaLike.
	require.NotNil(t, arsynth.Instance())

	lt := languagetool.NewJLanguageTool("ar")
	// Java: language.getTagger().tag(tokens) — full-sentence ArabicTagger.Tag path.
	lt.SentenceTagger = artag.DefaultArabicTagger.Tag
	return lt
}

// Java @Before fixture.
func arabicInflectedFixture(t *testing.T) (*ArabicInflectedOneWordReplaceRule, *languagetool.JLanguageTool) {
	t.Helper()
	lt := wireARInflectedLT(t)
	return NewArabicInflectedOneWordReplaceRule(nil), lt
}

// TestArabicInflectedOneWordReplaceRule_Rule twins Java testRule matrix.
func TestArabicInflectedOneWordReplaceRule_Rule(t *testing.T) {
	rule, lt := arabicInflectedFixture(t)

	assertCorrect := func(sentence string) {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(sentence))
		require.Equal(t, 0, len(matches), "correct: %q", sentence)
	}
	assertIncorrect := func(sentence string) {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(sentence))
		require.NotEqual(t, 0, len(matches), "incorrect: %q", sentence)
	}

	// Correct sentences:
	assertCorrect("أجريت بحوثا في المخبر")
	assertCorrect("وجعل لكم من أزواجكم بنين وحفدة")

	// errors:
	assertIncorrect("أجريت أبحاثا في المخبر")
	assertIncorrect("وجعل لكم من أزواجكم بنين وأحفاد")
}

// TestArabicInflectedOneWordReplaceRule_Metadata twins Java ctor metadata + example pair.
func TestArabicInflectedOneWordReplaceRule_Metadata(t *testing.T) {
	// Metadata does not need full dict analysis path.
	rule := NewArabicInflectedOneWordReplaceRule(nil)

	require.Equal(t, ArabicInflectedOneWordReplaceRuleID, rule.GetID())
	require.Equal(t, "AR_INFLECTED_ONE_WORD", rule.GetID())
	require.Equal(t, "قاعدة تطابق الكلمات التي يجب تجنبها وتقترح تصويبا لها", rule.GetDescription())
	// Java getShort(): double space after أن
	require.Equal(t, "خطأ، يفضل أن  يقال:", rule.GetShort())
	require.Equal(t, " لا تقل '$match' بل قل: $suggestions", rule.GetMessage())
	require.Equal(t, " أو ", rule.GetSuggestionsSeparator())
	require.Equal(t, "ar", rule.GetLocaleCode())
	require.Equal(t, ArabicInflectedOneWordFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/ar/inflected_one_word.txt"}, rule.GetFileNames())

	// Java: Categories.MISC; setLocQualityIssueType(ITSIssueType.Inconsistency)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSInconsistency, rule.GetLocQualityIssueType())

	// example pair
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "أجريت <marker>أبحاثا</marker> في المخبر", inc[0].GetExample())
	require.Equal(t, []string{"بحوثا"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "أجريت <marker>بحوثا</marker> في المخبر.", corr[0].GetExample())

	// resource pairs (last wrongWords map = no-space)
	ww := rule.wrongWords
	require.GreaterOrEqual(t, len(ww), 1)
	last := ww[len(ww)-1]
	require.Contains(t, last, "أبحاث")
	require.Equal(t, "بحوث", last["أبحاث"].Suggestion)
	require.Equal(t, "بحوث أفصح وأصحّ", last["أبحاث"].Message)
	require.Contains(t, last, "أحفاد")
	require.Equal(t, "حفدة", last["أحفاد"].Suggestion)
}

// TestArabicInflectedOneWordReplaceRule_MatchMessage twins Java Match message/short
// for the official incorrect example (real AR analysis).
func TestArabicInflectedOneWordReplaceRule_MatchMessage(t *testing.T) {
	rule, lt := arabicInflectedFixture(t)
	matches := rule.Match(lt.GetAnalyzedSentence("أجريت أبحاثا في المخبر"))
	require.NotEmpty(t, matches)
	m := matches[0]
	// short: "خطأ في استعمال كلمة:" + sugMsg
	require.Contains(t, m.GetShortMessage(), "خطأ في استعمال كلمة:")
	require.Contains(t, m.GetShortMessage(), "بحوث أفصح وأصحّ")
	// message embeds <suggestion>…</suggestion>&nbsp;
	require.Contains(t, m.GetMessage(), "' الكلمة خاطئة ")
	require.Contains(t, m.GetMessage(), "أبحاثا")
	require.Contains(t, m.GetMessage(), " ' ،")
	require.Contains(t, m.GetMessage(), ". استعمل  ")
	require.Contains(t, m.GetMessage(), "<suggestion>")
	require.Contains(t, m.GetMessage(), "</suggestion>&nbsp;")
	// span on أبحاثا token
	require.Greater(t, m.GetToPos(), m.GetFromPos())
	require.NotEmpty(t, m.GetSuggestedReplacements())
}
