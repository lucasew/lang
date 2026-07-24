package en

// Twin of languagetool-standalone/src/test/java/org/languagetool/rules/en/EnglishForL2SpeakersFalseFriendRuleTest.java
// Scope: testRule only (testMessageDetailData is @Ignore in Java).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/languagemodel"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/stretchr/testify/require"
)

// l2FakeCountMap ports FakeLanguageModel map keys (space-joined tokens).
// Java: FakeLanguageModel.getCount(List) → map.get(String.join(" ", tokens)).
type l2FakeCountMap map[string]int

func (m l2FakeCountMap) GetCountToken(token string) int64 {
	if n, ok := m[token]; ok {
		return int64(n)
	}
	return 0
}

func (m l2FakeCountMap) GetCount(tokens []string) int64 {
	if len(tokens) == 0 {
		return 0
	}
	if len(tokens) == 1 {
		return m.GetCountToken(tokens[0])
	}
	return m.GetCountToken(strings.Join(tokens, " "))
}

func (m l2FakeCountMap) GetTotalTokenCount() int64 {
	var s int64
	for _, v := range m {
		s += int64(v)
	}
	return s
}

// newL2FakeLanguageModel ports new FakeLanguageModel(map) via BaseLanguageModel.getPseudoProbability.
func newL2FakeLanguageModel(counts map[string]int) ngrams.LanguageModel {
	base := languagemodel.NewBaseLanguageModel(l2FakeCountMap(counts))
	return ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		return base.GetPseudoProbability(tokens)
	})
}

// Port of EnglishForL2SpeakersFalseFriendRuleTest.testRule
func TestEnglishForL2SpeakersFalseFriendRule_Rule(t *testing.T) {
	// Official confusion_sets_l2_fr.txt (Java CPR constructor path)
	require.NotEmpty(t, discoverENL2ConfusionSet("confusion_sets_l2_fr.txt"),
		"official confusion_sets_l2_fr.txt required")
	// Official false-friends.xml for L2 message rewrite
	require.NotEmpty(t, discoverFalseFriendsXML(), "official false-friends.xml required")
	// English tagger for isBaseformMatch on "achieved" → lemma "achieve" (Java lang.getTagger())
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree (third_party/english-pos-dict or inspiration)")
	}
	tagen.EnsureDefaultEnglishTagger()
	require.NotNil(t, tagen.EnglishTagWord(), "EnglishTagWord must load from english.dict")

	// Java FakeLanguageModel map from testRule
	mapCounts := map[string]int{
		// first test:
		"will achieve":      1,
		"will complete":     10,
		"achieve her":       1,
		"complete her":      5,
		"achieve her task":  1,
		"complete her task": 10,
		// second test:
		"was completed":      100,
		"was completed .":    100,
		"form was completed": 100,
	}
	lm := newL2FakeLanguageModel(mapCounts)

	// Java: new EnglishForFrenchFalseFriendRule(messages, lm, fr, en)
	// loads confusion_sets_l2_fr.txt + false-friends.xml; TagWord = English tagger
	ClearL2FalseFriendRuleCache()
	fr := NewEnglishForFrenchFalseFriendRuleWithLM(lm)
	require.Equal(t, "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS", fr.GetID())
	require.Equal(t, "fr", fr.MotherTongue)
	require.Equal(t, []string{"confusion_sets_l2_fr.txt"}, fr.GetFilenames())
	// Constructor must load official pairs (no hand-injected WordToPairs invent)
	require.NotEmpty(t, fr.WordToPairs, "WordToPairs must load from confusion_sets_l2_fr.txt")
	require.NotEmpty(t, fr.WordToPairs["achieve"], "achieve pair from official L2 FR set")
	require.NotEmpty(t, fr.WordToPairs["achieved"], "achieved pair from official L2 FR set")

	// Java: lt.getAnalyzedSentence — EN path (EnglishWordTokenizer + EnglishTagger)
	sent1 := tagen.AnalyzeEnglishSentence("She will achieve her task.")
	matches1 := fr.Match(sent1)
	require.Len(t, matches1, 1)
	require.Contains(t, matches1[0].GetMessage(), `"achieve" (English) means "réaliser" (French)`)

	sent2 := tagen.AnalyzeEnglishSentence("The code only worked if the form was achieved.")
	matches2 := fr.Match(sent2)
	require.Len(t, matches2, 1)
	require.Contains(t, matches2[0].GetMessage(), `"achieve" (English) means "réaliser" (French)`)

	// Optional non-invent control: preferred form should not match (Java does not assert)
	matches3 := fr.Match(tagen.AnalyzeEnglishSentence("She will complete her task."))
	require.Empty(t, matches3)
}
