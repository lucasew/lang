package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAndGroup_PrepareAndGroupResolvesRefs(t *testing.T) {
	// And-member with setpos match from first match token (TokenRef 0).
	andRef := NewPatternToken("", false, false, false)
	andRef.SetPosToken(PosToken{Regexp: true})
	mm := NewMatch("NN", "NN", true, "", "", CaseNone, true, false, IncludeNone)
	mm.SetTokenRef(0)
	andRef.SetMatch(mm)

	am := NewPatternTokenMatcher(andRef)
	require.True(t, andRef.IsReferenceElement())
	pos, lem := "NN", "dog"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("dog", &pos, &lem), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("cat", &pos, &lem), 4),
	}
	am.PrepareAndGroup(0, toks, "en") // no and-group on andRef itself
	am.ResolveReference(0, toks, "en")
	pt := am.GetPatternToken()
	require.NotNil(t, pt.Pos)
	require.Equal(t, "NN", pt.Pos.PosTag)
	require.False(t, pt.IsReferenceElement(), "compiled clears TokenMatch")
	require.True(t, am.IsMatched(toks[1].GetAnalyzedToken(0)))
}

func TestRawPos_UsesPreDisambigWhenPresent(t *testing.T) {
	// Post-disambig: second token stripped to empty POS; pre keeps NN.
	posNN, lem := "NN", "word"
	pre1 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("The", nil, nil), 0)
	pre2 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("word", &posNN, &lem), 4)
	post1 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("The", nil, nil), 0)
	post2 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("word", nil, nil), 4) // POS removed by "disambig"
	sent := languagetool.NewAnalyzedSentenceFull(
		[]*languagetool.AnalyzedTokenReadings{post1, post2},
		[]*languagetool.AnalyzedTokenReadings{pre1, pre2},
	)

	tok := NewPatternToken("word", false, false, false)
	tok.SetPosToken(PosToken{PosTag: "NN"})
	pr := NewPatternRule("RAW", "en", []*PatternToken{tok}, "d", "m", "")
	pr.InterpretPreDisambig = true
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1, "raw_pos should see pre-disambig NN")

	pr2 := NewPatternRule("POST", "en", []*PatternToken{tok}, "d", "m", "")
	pr2.InterpretPreDisambig = false
	ms2, err := pr2.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms2, "post-disambig has no NN")
}
