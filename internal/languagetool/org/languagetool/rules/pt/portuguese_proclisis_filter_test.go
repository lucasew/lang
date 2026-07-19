package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestPortugueseProclisisFilter(t *testing.T) {
	f := NewPortugueseProclisisFilter()
	got := f.Suggest([]struct{ Token, POS string }{
		{Token: "fazê-lo", POS: "VMN0000:PP3MSA00"},
	})
	require.Contains(t, got, "o fazê")
	// nos with plural verb ending
	got = f.Suggest([]struct{ Token, POS string }{
		{Token: "dizem-nos", POS: "VMIP3P0:PP1CPO00"},
	})
	require.Contains(t, got, "nos dizem")
	require.Contains(t, got, "os dizem")
}

func TestPortugueseProclisisFilter_AcceptRuleMatch(t *testing.T) {
	f := NewPortugueseProclisisFilter()
	pos := "VMIP1P0:PP3MSA00"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("bebemo-lo", &pos, nil), 0)
	m := rules.NewRuleMatch(nil, nil, 0, 9, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"foo": "bar"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "o bebemo")
}

func TestPortugueseProclisisFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.PortugueseProclisisFilter"))
}
