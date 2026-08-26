package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/PolishWordRepeatRuleTest.java
//
// Java setup:
//   final PolishWordRepeatRule rule = new PolishWordRepeatRule(TestTools.getEnglishMessages());
//   JLanguageTool lt = new JLanguageTool(new Polish());
//   matches = rule.match(lt.getAnalyzedSentence(...));
//
// Go: NewPolishWordRepeatRule + NewJLanguageTool("pl") with PolishTagger (real polish.dict).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	"github.com/stretchr/testify/require"
)

// wirePLWordRepeatLT ports JLanguageTool(new Polish()) for this rule twin.
// Skips when polish.dict is unavailable.
func wirePLWordRepeatLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	if tagpl.DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not in tree")
	}
	tagpl.EnsureDefaultPolishTagger()
	require.NotNil(t, tagpl.DefaultPolishTagger)
	require.NotNil(t, tagpl.DefaultPolishTagger.GetWordTagger())
	require.NotEmpty(t, tagpl.PolishPOSDictPath(), "real polish.dict must load")

	lt := languagetool.NewJLanguageTool("pl")
	// Java: language.getTagger().tag(tokens) — full-sentence PolishTagger.Tag path.
	lt.SentenceTagger = tagpl.DefaultPolishTagger.Tag
	return lt
}

// TestPolishWordRepeatRule_Rule twins Java PolishWordRepeatRuleTest.testRule.
func TestPolishWordRepeatRule_Rule(t *testing.T) {
	// Java: new PolishWordRepeatRule(TestTools.getEnglishMessages())
	rule := NewPolishWordRepeatRule(nil)
	lt := wirePLWordRepeatLT(t)

	// correct
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("To jest zdanie próbne."))))
	// się twice: excluded via EXC_WORDS lemma "się"
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("On tak się bardzo nie martwił, bo przecież musiał się umyć."))))
	// repeated prepositions, don't count'em (prep:.* POS)
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Na dyskotece tańczył jeszcze, choć był na bani."))))
	// sf bug report:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Żadnych „ale”."))))

	// incorrect
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("Był on bowiem pięknym strzelcem bowiem."))))
	// Java assertEquals(2, ...): lemma "mówić" (Mówiła/mówić) + surface/lemma "długo"
	require.Equal(t, 2, len(rule.Match(lt.GetAnalyzedSentence("Mówiła długo, żeby tylko mówić długo."))))
}

// Metadata twin (Java getId / getDescription / getMessage / getShortMessage /
// AdvancedWordRepeatRule ctor: MISC, defaultOff, Style + example pair).
func TestPolishWordRepeatRule_Metadata(t *testing.T) {
	r := NewPolishWordRepeatRule(nil)
	require.Equal(t, "PL_WORD_REPEAT", r.GetID())
	require.Equal(t, "Powtórzenia wyrazów w zdaniu (monotonia stylistyczna)", r.GetDescription())
	require.Equal(t, "Powtórzony wyraz w zdaniu", r.Message)
	require.Equal(t, "Powtórzenie wyrazu", r.ShortMessage)
	require.True(t, r.IsDefaultOff())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(rules.CategoryMisc))
	require.NotEmpty(t, r.GetIncorrectExamples())
	// Java example: second długo → powoli (style synonym)
	require.Equal(t, []string{"powoli"}, r.GetIncorrectExamples()[0].GetCorrections())
}
