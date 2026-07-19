package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

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

func TestEnglishRepeatedWordsRule(t *testing.T) {
	rule := NewEnglishRepeatedWordsRule(nil)
	// synonyms: need/VB.*/B-VP=require — Java uses lemma + POS + chunk
	ss := languagetool.SentenceStartTagName
	need1 := enAtr("need", "VB", "need")
	need1.SetChunkTags([]string{"B-VP"})
	need2 := enAtr("need", "VB", "need")
	need2.SetChunkTags([]string{"B-VP"})
	s1 := enSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		enAtr("I", "PRP", "I"),
		need1,
		enAtr("help", "NN", "help"),
		enAtr(".", ".", "."),
	)
	s2 := enSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		enAtr("I", "PRP", "I"),
		enAtr("still", "RB", "still"),
		need2,
		enAtr("time", "NN", "time"),
		enAtr(".", ".", "."),
	)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "require")
}

func TestEnglishRepeatedWords_IsExceptionNNP(t *testing.T) {
	nnp := enAtr("Paris", "NNP", "Paris")
	require.True(t, englishRepeatedWordsIsException([]*languagetool.AnalyzedTokenReadings{nnp}, 0, true, true, false))
	// mid-sentence capital (not NNP) still exception
	need := enAtr("Need", "VB", "need")
	require.True(t, englishRepeatedWordsIsException([]*languagetool.AnalyzedTokenReadings{need}, 0, false, true, false))
}

func TestEnglishRepeatedWords_Messages(t *testing.T) {
	r := NewEnglishRepeatedWordsRule(nil)
	require.Equal(t, "Suggest synonyms for repeated words.", r.GetDescription())
	require.Equal(t, "Style: repeated word", r.ShortMsg)
	require.Equal(t, 1, r.MinToCheckParagraph())
}

func TestEnglishRepeatedWords_AntiPatternsCount(t *testing.T) {
	require.Equal(t, 24, len(EnglishRepeatedWordsAntiPatterns), "Java ANTI_PATTERNS 24/24")
	require.NotNil(t, NewEnglishRepeatedWordsRule(nil).SentenceWithImmunization)
}

// Java: "need to" immunizes "need" so repeated need before "to" is not matched as synonym target.
func TestEnglishRepeatedWords_AntiPatternNeedTo(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	// Sent1: I need help.
	need1 := enAtr("need", "VB", "need")
	need1.SetChunkTags([]string{"B-VP"})
	s1 := enSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		enAtr("I", "PRP", "I"),
		need1,
		enAtr("help", "NN", "help"),
		enAtr(".", ".", "."),
	)
	// Sent2: I need to go. — "need" before "to" should be immunized
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
	// Without immunization this would fire (need/VB.*/B-VP in synonyms); with anti-pattern → 0
	require.Equal(t, 0, len(NewEnglishRepeatedWordsRule(nil).MatchList([]*languagetool.AnalyzedSentence{s1, s2})))
}

func TestEnglishRepeatedWords_CategoryPicky(t *testing.T) {
	r := NewEnglishRepeatedWordsRule(nil)
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), r.GetCategory().GetID())
	require.Equal(t, rules.ITSStyle, r.GetLocQualityIssueType())
	require.True(t, r.HasTag(rules.TagPicky))
}
