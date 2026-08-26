package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/MissingHyphenRuleTest.java
//
// Java setup:
//   MissingHyphenRule rule = new MissingHyphenRule(TestTools.getEnglishMessages(),
//       new UkrainianTagger().getWordTagger());
//   JLanguageTool lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis (UkrainianTagger + official ukrainian.dict WordTagger)
// via wireUKWordRepeatLT — not invent POS.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
	"github.com/stretchr/testify/require"
)

// TestMissingHyphenRule_Rule twins MissingHyphenRuleTest.testRule — every Java case.
func TestMissingHyphenRule_Rule(t *testing.T) {
	// Java: new UkrainianTagger().getWordTagger() + JLanguageTool(Ukrainian.DEFAULT_VARIANT)
	wt, lt := wireMissingHyphen(t)
	// Java: new MissingHyphenRule(TestTools.getEnglishMessages(), wordTagger)
	rule := NewMissingHyphenRule(nil, wt)

	// incorrect sentences:
	matches := rule.Match(lt.GetAnalyzedSentence("Поїхали у штаб квартиру."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"штаб-квартиру"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("Роблю «тайм аут»"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"тайм-аут"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("Такий компакт диск."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"компакт-диск"}, matches[0].GetSuggestedReplacements())

	// spelling 2019 (:alt → join without hyphen)
	matches = rule.Match(lt.GetAnalyzedSentence("Такий міні автомобіль."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"мініавтомобіль"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("Арт проект вийшов провальним."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Артпроект"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("Шоу бізнес - не моє."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Шоубізнес"}, matches[0].GetSuggestedReplacements())

	// en-dash U+2013 in Java source
	matches = rule.Match(lt.GetAnalyzedSentence("на шоу поп–діви"))
	require.Equal(t, 0, len(matches))

	// TODO in Java: hard - two errors, should be "ексвіцепрезидент"
	matches = rule.Match(lt.GetAnalyzedSentence("екс віце-президент"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"ексвіце-президент"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("всі медіа півострова."))
	require.Equal(t, 0, len(matches))

	// correct sentences:
	matches = rule.Match(lt.GetAnalyzedSentence("Тут все гаразд."))
	require.Equal(t, 0, len(matches))

	matches = rule.Match(lt.GetAnalyzedSentence("Такий блок схемі не потрібен."))
	require.Equal(t, 0, len(matches))

	matches = rule.Match(lt.GetAnalyzedSentence("конгрес Трансваалю."))
	require.Equal(t, 0, len(matches))

	matches = rule.Match(lt.GetAnalyzedSentence("конгрес профспілок."))
	require.Equal(t, 0, len(matches))

	matches = rule.Match(lt.GetAnalyzedSentence("рентген п'яти."))
	require.Equal(t, 0, len(matches))
}

// TestMissingHyphenRule_Metadata checks id / description / ITS Misspelling.
func TestMissingHyphenRule_Metadata(t *testing.T) {
	r := NewMissingHyphenRule(nil, nil)
	require.Equal(t, "UK_MISSING_HYPHEN", r.GetID())
	require.Equal(t, "Пропущений дефіс", r.GetDescription())
	require.Equal(t, rules.ITSMisspelling, r.GetLocQualityIssueType())
}

// wireMissingHyphen ports JLanguageTool(Ukrainian.DEFAULT_VARIANT) + UkrainianTagger.getWordTagger().
func wireMissingHyphen(t *testing.T) (tagging.WordTagger, *languagetool.JLanguageTool) {
	t.Helper()
	lt := wireUKWordRepeatLT(t)
	p := discoverUkrainianPOSDict()
	mt := tagging.OpenMorfologikTagger(p)
	require.NotNil(t, mt, "OpenMorfologikTagger")
	tg := taguk.NewUkrainianTagger(mt)
	// Ensure analysis uses the same tagger instance as WordTagger source.
	lt.SentenceTagger = tg.Tag
	return tg.GetWordTagger(), lt
}
