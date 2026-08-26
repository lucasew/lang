package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicTransVerbRuleTest.java
//
// Java setup:
//   rule = new ArabicTransVerbRule(TestTools.getEnglishMessages());
//   lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
// Go: NewArabicTransVerbRule(nil) + NewJLanguageTool("ar") with
// real ArabicTagger.Tag as SentenceTagger (lemma+POS path; no invent tags).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	arsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	artag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/stretchr/testify/require"
)

// wireARTransVerbLT ports JLanguageTool(Languages.getLanguageForShortCode("ar")).
// Returns nil skip when arabic.dict / arabic_synth.dict unavailable.
func wireARTransVerbLT(t *testing.T) *languagetool.JLanguageTool {
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
	// Ensure synthesizer Instance can load for generateNewForm.
	require.NotNil(t, arsynth.Instance())

	lt := languagetool.NewJLanguageTool("ar")
	// Java: language.getTagger().tag(tokens) — full-sentence ArabicTagger.Tag path.
	lt.SentenceTagger = artag.DefaultArabicTagger.Tag
	return lt
}

// Java @Before fixture.
func arabicTransVerbFixture(t *testing.T) (*ArabicTransVerbRule, *languagetool.JLanguageTool) {
	t.Helper()
	lt := wireARTransVerbLT(t)
	return NewArabicTransVerbRule(nil), lt
}

// TestArabicTransVerbRule_Rule twins Java testRule matrix.
func TestArabicTransVerbRule_Rule(t *testing.T) {
	rule, lt := arabicTransVerbFixture(t)

	assertCorrect := func(sentence string) {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(sentence))
		require.Equal(t, 0, len(matches), "correct: %q", sentence)
	}
	assertIncorrect := func(sentence string) {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(sentence))
		// Java: assertEquals(1, matches.length);
		require.Equal(t, 1, len(matches), "incorrect: %q (got %d)", sentence, len(matches))
	}

	// correct sentences:
	assertCorrect("كان أَفَاضَ في الحديث")

	// errors:
	assertIncorrect("كان أفاض من الحديث")
	assertIncorrect("لقد أفاضت من الحديث")
	assertIncorrect("لقد أفاضت الحديث")
	assertIncorrect("كان أفاضها الحديث")
	assertIncorrect("إذ استعجل الأمر")
}

// TestArabicTransVerbRule_Metadata twins Java ctor metadata + example pair.
func TestArabicTransVerbRule_Metadata(t *testing.T) {
	// Metadata does not need full dict analysis path.
	rule := NewArabicTransVerbRule(nil)

	require.Equal(t, ArabicTransVerbRuleID, rule.GetID())
	require.Equal(t, "AR_VERB_TRANSITIVE_IINDIRECT", rule.GetID())
	// Java getDescription(): leading fatha (U+064E) on T
	require.Equal(t, "َTransitive verbs corrected to indirect transitive", rule.GetDescription())
	require.Equal(t, "أفعال متعدية بحرف، يخطئ في تعديتها", rule.GetShort())
	require.Equal(t, "'$match' الفعل خاطئ في التعدية بحرف: $suggestions", rule.GetMessage())
	require.Equal(t, " أو ", rule.GetSuggestionsSeparator())
	require.Equal(t, "ar", rule.GetLocaleCode())
	require.Equal(t, ArabicTransVerbFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/ar/verb_trans_to_untrans2.txt"}, rule.GetFileNames())

	// Java: Categories.MISC; setLocQualityIssueType(ITSIssueType.Misspelling)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// example pair (English placeholder as in Java)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "The train arrived <marker>a hour</marker> ago.", inc[0].GetExample())
	require.Equal(t, []string{"an hour"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "The train arrived <marker>an hour</marker> ago.", corr[0].GetExample())

	// resource pairs (last wrongWords map = no-space) — official lemmas only, no invent keys
	ww := rule.wrongWords
	require.GreaterOrEqual(t, len(ww), 1)
	last := ww[len(ww)-1]
	require.Contains(t, last, "أَفَاضَ")
	require.Equal(t, "في", last["أَفَاضَ"].Suggestion)
	require.Contains(t, last, "أفاض")
	require.Equal(t, "في", last["أفاض"].Suggestion)
	require.Contains(t, last, "أَقَادَ")
	require.Equal(t, "إلى", last["أَقَادَ"].Suggestion)
	require.Contains(t, last, "اِسْتَعْجَلَ")
	require.Equal(t, "في", last["اِسْتَعْجَلَ"].Suggestion)
}

// TestArabicTransVerbRule_MatchMessage twins Java Match message/short/suggestion
// for an official incorrect sentence (real AR analysis).
func TestArabicTransVerbRule_MatchMessage(t *testing.T) {
	rule, lt := arabicTransVerbFixture(t)
	matches := rule.Match(lt.GetAnalyzedSentence("كان أفاض من الحديث"))
	require.Len(t, matches, 1)
	m := matches[0]
	// Java message embeds <suggestion>…</suggestion>
	require.Contains(t, m.GetMessage(), "<suggestion>")
	require.Contains(t, m.GetMessage(), "</suggestion>")
	require.Contains(t, m.GetMessage(), "بدلا من 'أفاض'")
	require.Contains(t, m.GetMessage(), "لأنّ الفعل ")
	require.Contains(t, m.GetMessage(), " متعد بحرف  .")
	require.Equal(t, "خطأ في الفعل المتعدي بحرف", m.ShortMessage)
	// span: offset covers prev verb only; pattern extends through next token
	toks := lt.GetAnalyzedSentence("كان أفاض من الحديث").GetTokensWithoutWhitespace()
	// tokens: SENT_START, كان, أفاض, من, الحديث
	require.GreaterOrEqual(t, len(toks), 4)
	verb := toks[2] // أفاض
	next := toks[3] // من
	require.Equal(t, "أفاض", verb.GetToken())
	require.Equal(t, verb.GetStartPos(), m.GetFromPos())
	require.Equal(t, verb.GetEndPos(), m.GetToPos())
	require.Equal(t, verb.GetStartPos(), m.GetPatternFromPos())
	require.Equal(t, next.GetEndPos(), m.GetPatternToPos())
	// suggestion: unattached verb + space + prep (Java synthesizer path)
	reps := m.GetSuggestedReplacements()
	require.NotEmpty(t, reps)
	require.Contains(t, reps[0], "في")
}
