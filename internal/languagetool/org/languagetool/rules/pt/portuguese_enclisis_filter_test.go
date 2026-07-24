package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestPortugueseEnclisisFilter_PronounTags(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	tags := f.PronounTags([]PronounTagReading{{Token: "nos", POS: "PP1CPO00"}}, "dizem", false)
	require.Equal(t, []string{"PP1CPO00", "PP3MPA00"}, tags)

	tags = f.PronounTags([]PronounTagReading{{Token: "eles", POS: "PP3MPN00"}}, "ver", true)
	require.Equal(t, []string{"PP3MPA00"}, tags)
}

func TestPortugueseEnclisisFilter_Suggest(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SynthesizeEnclisis = func(verb, pos, ptag string) []string {
		return []string{verb + "-" + ptag}
	}
	got := f.Suggest(VerbReading{Token: "ver", POS: "VMN0000"}, []string{"PP3MSA00"})
	require.Equal(t, []string{"ver-PP3MSA00"}, got)
}

func TestPortugueseEnclisisFilter_AcceptRuleMatch(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SynthesizeEnclisis = func(verb, pos, ptag string) []string {
		return []string{verb + "-no"}
	}
	vpos := "VMIP3P0"
	ppos := "PP3MSA00"
	verb := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("tinham", &vpos, nil), 0)
	// middle token unused (pronounPos:2)
	mid := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("-", nil, nil), 6)
	pron := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("o", &ppos, nil), 7)
	m := rules.NewRuleMatch(nil, nil, 0, 8, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"verbPos": "0", "pronounPos": "2", "convertToAccusative": "False",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, mid, pron}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"tinham-no"}, out.GetSuggestedReplacements())

	// no PP tags → null
	out = f.AcceptRuleMatch(m, map[string]string{
		"verbPos": "0", "pronounPos": "2", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, mid,
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("xyz", nil, nil), 7)}, nil)
	require.Nil(t, out)
}

func TestPortugueseEnclisisFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.PortugueseEnclisisFilter"))
}
