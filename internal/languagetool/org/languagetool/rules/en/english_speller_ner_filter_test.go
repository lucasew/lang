package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ner"
	"github.com/stretchr/testify/require"
)

type mapCounts map[string]int64

func (m mapCounts) GetCountToken(token string) int64 { return m[token] }
func (m mapCounts) GetCount(tokens []string) int64 {
	return m[stringsJoin(tokens)]
}

func stringsJoin(tokens []string) string {
	out := ""
	for i, t := range tokens {
		if i > 0 {
			out += " "
		}
		out += t
	}
	return out
}

func TestFilterNERMatches_DropsRareName(t *testing.T) {
	// Covered "Fastow" has count 0; suggestions also 0 → filter out
	lm := mapCounts{}
	m := rules.NewRuleMatch(nil, languagetool.AnalyzePlain("Fastow said so."), 0, 6, "msg")
	m.SetSuggestedReplacements([]string{"Fa stow", "Fast ow"})
	// "Fa stow" is 2 tokens with 0 counts; nonZero=0, lookupFailures=0 → filter
	out := filterNERMatches([]*rules.RuleMatch{m}, "Fastow said so.", []ner.Span{ner.NewSpan(0, 6)}, lm)
	require.Empty(t, out)
}

func TestFilterNERMatches_KeepsWhenCommonReplClose(t *testing.T) {
	// covered "Colour" count 1; "Color" count 100, dist=1 → keep
	lm := mapCounts{"Colour": 1, "Color": 100}
	m := rules.NewRuleMatch(nil, languagetool.AnalyzePlain("Colour is fine."), 0, 6, "msg")
	m.SetSuggestedReplacements([]string{"Color"})
	out := filterNERMatches([]*rules.RuleMatch{m}, "Colour is fine.", []ner.Span{ner.NewSpan(0, 6)}, lm)
	require.Len(t, out, 1)
}

func TestFilterNERMatches_DropsDistantCommonRepl(t *testing.T) {
	// mostCommon far from covered → drop
	lm := mapCounts{"Xyzabc": 1, "CompletelyDifferent": 100}
	m := rules.NewRuleMatch(nil, languagetool.AnalyzePlain("Xyzabc here."), 0, 6, "msg")
	m.SetSuggestedReplacements([]string{"CompletelyDifferent"})
	out := filterNERMatches([]*rules.RuleMatch{m}, "Xyzabc here.", []ner.Span{ner.NewSpan(0, 6)}, lm)
	require.Empty(t, out)
}

func TestEnLevenshtein(t *testing.T) {
	require.Equal(t, 0, enLevenshtein("a", "a"))
	require.Equal(t, 1, enLevenshtein("Color", "Colour"))
	require.True(t, enLevenshtein("abc", "xyz") > 2)
}

// Twin: covered = sentenceText.substring(from,to) with UTF-16 positions.
func TestFilterNERMatches_CoveredUTF16Span(t *testing.T) {
	// "café" is 4 UTF-16 units; byte length 5. Positions [0,4) must yield "café".
	lm := mapCounts{} // zero counts → filter when NER covers
	m := rules.NewRuleMatch(nil, languagetool.AnalyzePlain("café here"), 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"caf e"}) // 2 tokens, 0 counts
	out := filterNERMatches([]*rules.RuleMatch{m}, "café here", []ner.Span{ner.NewSpan(0, 4)}, lm)
	// Uppercase required for NER path; café starts lower → not filtered (continue)
	require.Len(t, out, 1)

	// Title case multi-byte: "Café" UTF-16 len 4
	lm2 := mapCounts{}
	m2 := rules.NewRuleMatch(nil, languagetool.AnalyzePlain("Café said"), 0, 4, "msg")
	m2.SetSuggestedReplacements([]string{"Ca fe"})
	out2 := filterNERMatches([]*rules.RuleMatch{m2}, "Café said", []ner.Span{ner.NewSpan(0, 4)}, lm2)
	require.Empty(t, out2, "title-case covered with zero-count multi-token sugs should drop")
}

func TestEnVariantBlogURL_Colour(t *testing.T) {
	u := enVariantBlogURL("colour")
	require.Contains(t, u, "our-or")
}
