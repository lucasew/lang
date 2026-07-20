package languagetool

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiThreadedJLanguageTool(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 2)
	require.Equal(t, 2, lt.GetThreadPoolSize())
	var n atomic.Int32
	lt.Matchers = []SentenceMatcherFunc{
		func(s *AnalyzedSentence) error {
			n.Add(1)
			return nil
		},
	}
	sents := []*AnalyzedSentence{
		NewAnalyzedSentence([]*AnalyzedTokenReadings{NewAnalyzedTokenReadings(NewAnalyzedToken("a", nil, nil))}),
		NewAnalyzedSentence([]*AnalyzedTokenReadings{NewAnalyzedTokenReadings(NewAnalyzedToken("b", nil, nil))}),
		NewAnalyzedSentence([]*AnalyzedTokenReadings{NewAnalyzedTokenReadings(NewAnalyzedToken("c", nil, nil))}),
	}
	require.NoError(t, lt.CheckSentences(sents))
	require.Equal(t, int32(3), n.Load())
	lt.Shutdown()
}

func TestMultiThreadedJLanguageTool_ParallelCheck(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 4)
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	// multi-sentence document
	src := "This is is one. And an test two. Fine here."
	m := lt.Check(src)
	require.NotEmpty(t, m)
	// both rule types may fire
	ids := map[string]bool{}
	for _, x := range m {
		ids[x.RuleID] = true
		require.GreaterOrEqual(t, x.FromPos, 0)
		require.LessOrEqual(t, x.ToPos, len(src))
	}
	require.True(t, ids["WORD_REPEAT_RULE"] || ids["EN_A_VS_AN"])
	lt.Shutdown()
	require.True(t, lt.IsShutdown())
}

func TestMultiThreadedJLanguageTool_TextLevel(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 4)
	// text-level only: successive sentence starts
	lt.AddTextLevelRuleChecker("WORD_REPEAT_BEGINNING_RULE", func(sents []*AnalyzedSentence) []LocalMatch {
		if len(sents) < 3 {
			return nil
		}
		// soft inject: flag if three sentences start with same surface token
		var starts []string
		for _, s := range sents {
			toks := s.GetTokensWithoutWhitespace()
			if len(toks) > 1 {
				starts = append(starts, toks[1].GetToken())
			}
		}
		if len(starts) >= 3 && starts[0] == starts[1] && starts[1] == starts[2] {
			return []LocalMatch{{
				FromPos: 0, ToPos: 1, RuleID: "WORD_REPEAT_BEGINNING_RULE",
				Message: "repeated beginning",
			}}
		}
		return nil
	})
	m := lt.Check("Also one. Also two. Also three.")
	found := false
	for _, x := range m {
		if x.RuleID == "WORD_REPEAT_BEGINNING_RULE" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)

	lt.SetMode(ModeAllButTextLevel)
	m2 := lt.Check("Also one. Also two. Also three.")
	for _, x := range m2 {
		require.NotEqual(t, "WORD_REPEAT_BEGINNING_RULE", x.RuleID)
	}
}

func TestMultiThreadedJLanguageTool_AnalyzeParallel(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 2)
	out := lt.AnalyzeSentencesParallel([]string{"Hello.", "World."})
	require.Len(t, out, 2)
	require.NotNil(t, out[0])
	require.NotNil(t, out[1])
	// last sentence paragraph end
	toks := out[1].GetTokens()
	require.NotEmpty(t, toks)
	require.True(t, toks[len(toks)-1].IsParagraphEnd())
	lt.ShutdownWhenDone()
	require.True(t, lt.IsShutdown())
}

// Twin of AnalyzeSentenceCallable → getAnalyzedSentence (no SRX re-split on the part).
func TestAnalyzeSentenceCallable_GetAnalyzedSentence(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 1)
	// Multi-clause string already chosen as one SRX unit by the caller in Java;
	// getAnalyzedSentence must keep it as one AnalyzedSentence.
	text := "A sentence. This is not actual."
	c := AnalyzeSentenceCallable{LT: lt, Sentence: text}
	s, err := c.Call()
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, text, s.GetText())
	lt.ShutdownWhenDone()
}

func TestSentenceData(t *testing.T) {
	a := AnalyzePlain("hi")
	sd := NewSentenceData(a, "hi", 0, 1, 1)
	require.Equal(t, "hi", sd.Text)
	require.Greater(t, sd.WordCount, 0)
}

func TestCleanTokenAlias(t *testing.T) {
	c := CleanToken{Orig: "a\u00ADb", Clean: "ab"}
	require.Equal(t, "a\u00ADb", c.GetOrigToken())
	require.Equal(t, "ab", c.GetCleanToken())
}

func TestJLanguageToolModeParagraphEnums(t *testing.T) {
	require.Equal(t, Mode("ALL"), ModeAll)
	require.Equal(t, ParagraphHandling("NORMAL"), ParagraphNormal)
	var cancelled CheckCancelledCallback = func() bool { return true }
	require.True(t, cancelled())
}
