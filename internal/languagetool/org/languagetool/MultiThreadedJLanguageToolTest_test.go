package languagetool

// Twin of MultiThreadedJLanguageToolTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiThreadedJLanguageTool_Check(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 2)
	require.Equal(t, 2, lt.GetThreadPoolSize())
	var n int
	lt.Matchers = []SentenceMatcherFunc{
		func(s *AnalyzedSentence) error { n++; return nil },
	}
	s1 := AnalyzePlain("Hello world.")
	s2 := AnalyzePlain("Another sentence.")
	require.NoError(t, lt.CheckSentences([]*AnalyzedSentence{s1, s2}))
	require.Equal(t, 2, n)
}

func TestMultiThreadedJLanguageTool_ShutdownException(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 1)
	lt.Shutdown()
	require.True(t, lt.IsShutdown())
	require.Panics(t, func() {
		_ = lt.CheckSentences([]*AnalyzedSentence{AnalyzePlain("x")})
	})
}

func TestMultiThreadedJLanguageTool_ConfigurableThreadPoolSize(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 4)
	require.Equal(t, 4, lt.GetThreadPoolSize())
	lt0 := NewMultiThreadedJLanguageTool("en", 0)
	require.GreaterOrEqual(t, lt0.GetThreadPoolSize(), 1)
}

// Twin of MultiThreadedJLanguageToolTest.testTextAnalysis
func TestMultiThreadedJLanguageTool_TextAnalysis(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 2)
	sents := lt.AnalyzeSentencesParallel([]string{"Hello world.", "Another one."})
	require.Len(t, sents, 2)
	require.NotNil(t, sents[0])
	require.NotNil(t, sents[1])
}

// Twin of MultiThreadedJLanguageToolTest.testTwoRulesOnly
func TestMultiThreadedJLanguageTool_TwoRulesOnly(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 2)
	var hits int
	lt.Matchers = []SentenceMatcherFunc{
		func(s *AnalyzedSentence) error { hits++; return nil },
		func(s *AnalyzedSentence) error { hits++; return nil },
	}
	require.NoError(t, lt.CheckSentences([]*AnalyzedSentence{AnalyzePlain("One."), AnalyzePlain("Two.")}))
	require.Equal(t, 4, hits) // 2 matchers × 2 sentences
}
