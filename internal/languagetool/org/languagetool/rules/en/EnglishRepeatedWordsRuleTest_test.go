package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishRepeatedWordsRuleTest.java
//
// Java setup:
//   rule = new EnglishRepeatedWordsRule(TestTools.getEnglishMessages());
//   lt = new JLanguageTool(Languages.getLanguageForShortCode("en"));
//   assertCorrectText / getRuleMatches → rule.match(lt.analyzeText(s), aText)
//
// Go analysis path (Java getAnalyzedSentence + English chunker):
//   SRXSentenceTokenizer("en") + tagging/en.AnalyzeEnglishSentence per part
//   + EnglishChunker.AddChunkTags (Java Language.createDefaultChunker).
// Synonyms: official en/synonyms.txt; synthesizer: english_synth.dict.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func engRepeatedWordsLT(t *testing.T) {
	t.Helper()
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree (third_party/english-pos-dict or inspiration)")
	}
	if chunking.DiscoverOpenNLPChunkerModel() == "" ||
		chunking.DiscoverOpenNLPPOSModel() == "" ||
		chunking.DiscoverOpenNLPTokenModel() == "" {
		t.Skip("OpenNLP en-token/en-pos-maxent/en-chunker models required for EN chunk tags")
	}
	tagen.EnsureDefaultEnglishTagger()
	require.NotNil(t, tagen.EnglishTagWord(), "EnglishTagWord must load from english.dict")
}

// analyzeENRepeatedWordsText ports JLanguageTool("en").analyzeText:
// sentence tokenize, getAnalyzedSentence per part, EnglishChunker.addChunkTags.
func analyzeENRepeatedWordsText(t *testing.T, text string) []*languagetool.AnalyzedSentence {
	t.Helper()
	parts := tokenizers.NewSRXSentenceTokenizer("en").Tokenize(text)
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		s := tagen.AnalyzeEnglishSentence(p)
		require.NotNil(t, s, "AnalyzeEnglishSentence(%q)", p)
		// Java: language.getChunker().addChunkTags on tagged tokens (pre-disambig path
		// still carries tags; chunker mutates token readings in place).
		chunking.NewEnglishChunker().AddChunkTags(s.GetTokens())
		out = append(out, s)
	}
	return out
}

// TestEnglishRepeatedWordsRule_Rule twins Java EnglishRepeatedWordsRuleTest.testRule.
func TestEnglishRepeatedWordsRule_Rule(t *testing.T) {
	engRepeatedWordsLT(t)
	rule := NewEnglishRepeatedWordsRule(nil)

	assertCorrectText := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.MatchList(analyzeENRepeatedWordsText(t, s))), "correct: %q", s)
	}
	getRuleMatches := func(s string) []*rules.RuleMatch {
		t.Helper()
		return rule.MatchList(analyzeENRepeatedWordsText(t, s))
	}

	// form
	assertCorrectText("They went on to form a new group. The bacteria causes a blood clot to form in the jugular vein.")

	// global
	assertCorrectText("Asia Global Crossing Ltd. Global Crossing and Asia Global Crossing.")
	assertCorrectText("It was a global effort. Announcing the participation of Enron Global Markets.")

	// great
	assertCorrectText("Matthew S. Anderson, Peter the Great. The Tomahawks were shipped from Great Britain.")
	assertCorrectText("It was great. The Tomahawks were shipped from Great Britain.")

	// interesting
	assertCorrectText("I found it very interesting. An interesting fact about me is that I have a twin.")

	// maybe
	assertCorrectText("Maybe it's because I used to be a FLAMING LIBERAL? Or maybe it's because with age, I developed EYES THAT SEE!")

	// need
	assertCorrectText("It needs to be done. That needs to be done.")
	assertCorrectText("This was needed. There is a need to do it.")
	assertCorrectText("It needs to be done. That also needed to be done.")
	assertCorrectText("I still need to sign in somewhere. You need to sign in too.")

	// new
	assertCorrectText("This is a new experience. Happy New Year!")

	// often
	assertCorrectText("It's often gloomy outside. More often than not, it's raining.")
	assertCorrectText("We have bad weather here often. Often times, it's raining.")

	// problem
	assertCorrectText("The students were given some problems. They needed help to solve the problems.")
	assertCorrectText("Then, there were numerous problems after that. His initial interest lay in an attempt to solve Hilbert's fifth problem.")
	assertCorrectText("There were some problems with the tests. No problem, I'm not in a rush.")
	assertCorrectText("The students were given some problems. They were math problems.")

	// several
	assertCorrectText("We noticed them several times. Several thousand people stormed the gate.")

	// suggest
	assertCorrectText("I suggested that, but he also suggests that.")
	assertCorrectText("He suggested that we review them again. What do these suggest about the transaction history?")
	assertCorrectText("I suggested he look it over again. This strongly suggests that Mr. Batt is guilty.")

	// whole
	assertCorrectText("In this example, persons is used instead of people because the law applies to the individuals and never to the group as a whole.")

	// ignore sentences without period at the end
	assertCorrectText("I suggested this. She suggests that")

	matches := getRuleMatches("I suggested this. She suggests that.")
	require.Equal(t, 1, len(matches))
	require.Equal(t, 22, matches[0].GetFromPos())
	require.Equal(t, 30, matches[0].GetToPos())
	require.Equal(t, "proposes", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "recommends", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "submits", matches[0].GetSuggestedReplacements()[2])

	matches = getRuleMatches("I suggested this. She suggests that. And they suggested that.")
	require.Equal(t, 2, len(matches))
	require.Equal(t, 22, matches[0].GetFromPos())
	require.Equal(t, 46, matches[1].GetFromPos())
	require.Equal(t, "proposes", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "proposed", matches[1].GetSuggestedReplacements()[0])

	matches = getRuleMatches("The problem was weird. And the solutions needed to be weird.")
	require.Equal(t, 1, len(matches))
	require.Equal(t, "odd", matches[0].GetSuggestedReplacements()[0])
}

// TestEnglishRepeatedWords_IsExceptionNNP ports isException NNP / mid-sentence capital.
func TestEnglishRepeatedWords_IsExceptionNNP(t *testing.T) {
	nnp := enAtr("Paris", "NNP", "Paris")
	require.True(t, englishRepeatedWordsIsException([]*languagetool.AnalyzedTokenReadings{nnp}, 0, true, true, false))
	// mid-sentence capital (not NNP) still exception
	need := enAtr("Need", "VB", "need")
	require.True(t, englishRepeatedWordsIsException([]*languagetool.AnalyzedTokenReadings{need}, 0, false, true, false))
}

// TestEnglishRepeatedWords_Messages ports getDescription / getShortMessage / minToCheckParagraph.
func TestEnglishRepeatedWords_Messages(t *testing.T) {
	r := NewEnglishRepeatedWordsRule(nil)
	require.Equal(t, "EN_REPEATEDWORDS", r.GetID())
	require.Equal(t, "Suggest synonyms for repeated words.", r.GetDescription())
	require.Equal(t, "Style: repeated word", r.ShortMsg)
	require.Equal(t, 1, r.MinToCheckParagraph())
}

// TestEnglishRepeatedWords_AntiPatternsCount ports ANTI_PATTERNS size (24/24).
func TestEnglishRepeatedWords_AntiPatternsCount(t *testing.T) {
	require.Equal(t, 24, len(EnglishRepeatedWordsAntiPatterns), "Java ANTI_PATTERNS 24/24")
	require.NotNil(t, NewEnglishRepeatedWordsRule(nil).SentenceWithImmunization)
}

// TestEnglishRepeatedWords_AntiPatternNeedTo: "need to" immunizes need (anti-pattern 1).
func TestEnglishRepeatedWords_AntiPatternNeedTo(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	need1 := enAtr("need", "VB", "need")
	need1.SetChunkTags([]string{"B-VP"})
	s1 := enSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		enAtr("I", "PRP", "I"),
		need1,
		enAtr("help", "NN", "help"),
		enAtr(".", ".", "."),
	)
	need2 := enAtr("need", "VB", "need")
	need2.SetChunkTags([]string{"B-VP"})
	s2 := enSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		enAtr("I", "PRP", "I"),
		need2,
		enAtr("to", "TO", "to"),
		enAtr("go", "VB", "go"),
		enAtr(".", ".", "."),
	)
	require.Equal(t, 0, len(NewEnglishRepeatedWordsRule(nil).MatchList([]*languagetool.AnalyzedSentence{s1, s2})))
}

// TestEnglishRepeatedWords_CategoryPicky ports ctor setCategory / setTags(picky) / Style.
func TestEnglishRepeatedWords_CategoryPicky(t *testing.T) {
	r := NewEnglishRepeatedWordsRule(nil)
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), r.GetCategory().GetID())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.True(t, r.HasTag(rules.TagPicky))
}

// Hand-built token helpers for unit-level exception / anti-pattern checks.
func enAtr(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	var p, l *string
	if pos != "" {
		pp := pos
		p = &pp
	}
	if lemma != "" {
		ll := lemma
		l = &ll
	}
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, p, l), 0)
}

func enSent(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	pos := 0
	for _, t := range toks {
		if t == nil {
			continue
		}
		t.SetStartPos(pos)
		pos += len([]rune(t.GetToken())) + 1
	}
	return languagetool.NewAnalyzedSentence(toks)
}
