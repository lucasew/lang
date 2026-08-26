package morfologik

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestTokenizingSegments_Hyphen(t *testing.T) {
	pat := regexp.MustCompile(`-`)
	segs := tokenizingSegments("well-known", pat)
	require.Len(t, segs, 2)
	require.Equal(t, "well", segs[0].word)
	require.Equal(t, 0, segs[0].utf16Off)
	require.Equal(t, "known", segs[1].word)
	require.Equal(t, 5, segs[1].utf16Off) // "well-" is 5 UTF-16 units

	segs = tokenizingSegments("a-b-c", pat)
	require.Len(t, segs, 3)
	require.Equal(t, []string{"a", "b", "c"}, []string{segs[0].word, segs[1].word, segs[2].word})
	require.Equal(t, 0, segs[0].utf16Off)
	require.Equal(t, 2, segs[1].utf16Off)
	require.Equal(t, 4, segs[2].utf16Off)

	// no hyphen → whole word
	segs = tokenizingSegments("hello", pat)
	require.Len(t, segs, 1)
	require.Equal(t, "hello", segs[0].word)
}

// Both hyphen parts known → no match (Java tokenizingPattern getRuleMatches each part).
func TestMatch_TokenizingPattern_BothPartsOK(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("well")
	sp.AddWord("known")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	r.TokenizingPattern = regexp.MustCompile(`-`)
	ms, err := r.Match(languagetool.AnalyzePlain("well-known"))
	require.NoError(t, err)
	require.Empty(t, ms, "both parts accepted → no spelling match")
}

// Second part misspelled → flag only that segment span.
func TestMatch_TokenizingPattern_SecondPartMisspelled(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("well")
	sp.AddWord("known")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	r.TokenizingPattern = regexp.MustCompile(`-`)
	ms, err := r.Match(languagetool.AnalyzePlain("well-knwn"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	// "well-" is 5 units from start of token at 0 (after SENT_START)
	// token "well-knwn" starts at 0 in AnalyzePlain content
	m := ms[0]
	// segment "knwn" starts at utf16 offset 5
	require.Equal(t, 5, m.GetFromPos(), "should flag only misspelled segment start")
	require.Equal(t, 5+4, m.GetToPos(), "knwn length 4")
}

// Without pattern, whole compound is one misspelling span.
func TestMatch_NoTokenizingPattern_WholeWord(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("well")
	sp.AddWord("known")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	// TokenizingPattern nil
	ms, err := r.Match(languagetool.AnalyzePlain("well-known"))
	require.NoError(t, err)
	// whole token not in dict as one form
	require.NotEmpty(t, ms)
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, utf16LenMF("well-known"), ms[0].GetToPos())
}
