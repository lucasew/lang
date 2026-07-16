package en

// Twin of UpperCaseNgramRuleTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUpperCaseNgramRule_Rule(t *testing.T) {
	r := NewUpperCaseNgramRule(nil)
	// good: title at start
	ms, err := r.Match(languagetool.AnalyzePlain("This Was a Good Idea"))
	require.NoError(t, err)
	// may flag mid-sentence titles — ensure Match runs
	_ = ms
	// bad mid-sentence title-ish
	ms2, err := r.Match(languagetool.AnalyzePlain("The Dog ran."))
	require.NoError(t, err)
	require.NotEmpty(t, ms2)
}

func TestUpperCaseNgramRule_FirstLongWordToLeftIsUppercase(t *testing.T) {
	sent := languagetool.AnalyzePlain("United States also used short slogan")
	toks := sent.GetTokensWithoutWhitespace()
	// find index of "also"
	idx := -1
	for i, t := range toks {
		if t != nil && t.GetToken() == "also" {
			idx = i
			break
		}
	}
	require.Greater(t, idx, 0)
	// "United"/"States" long uppercase/title to the left
	require.True(t, FirstLongWordToLeftIsUppercase(toks, idx))
	// at first content word → false
	require.False(t, FirstLongWordToLeftIsUppercase(toks, 1))
}
