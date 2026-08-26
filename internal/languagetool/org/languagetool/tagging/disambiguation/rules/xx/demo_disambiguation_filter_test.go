package xx

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func TestDemoDisambiguationFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.tagging.disambiguation.rules.xx.DemoDisambiguationFilter"))
}

func TestDemoDisambiguationFilter_AcceptX9(t *testing.T) {
	f := NewDemoDisambiguationFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 1, "internal")
	x9 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("X9", nil, nil))
	other := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("nope", nil, nil))
	require.Equal(t, m, f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{x9}, nil))
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{other}, nil))
}

func TestDemoDisambiguationFilter_KeepDespiteFilter(t *testing.T) {
	// Filter-gated immunize: only "X9" token keeps the action.
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="DEMO_F" name="demo filter">
    <pattern>
      <token regexp="yes">.+</token>
    </pattern>
    <filter class="org.languagetool.tagging.disambiguation.rules.xx.DemoDisambiguationFilter" args="formme:foo"/>
    <disambig action="immunize"/>
  </rule>
</rules>`
	ars, err := disambigrules.NewDisambiguationRuleLoader().GetRulesFromString(xml, "xx", "t.xml")
	require.NoError(t, err)
	require.Len(t, ars, 1)
	require.NotNil(t, ars[0].Filter)

	sentX9 := languagetool.AnalyzePlain("X9")
	out := ars[0].Replace(sentX9)
	// X9 should be immunized
	foundImm := false
	for _, tok := range out.GetTokens() {
		if tok != nil && tok.GetToken() == "X9" && tok.IsImmunized() {
			foundImm = true
		}
	}
	require.True(t, foundImm, "X9 must pass filter and immunize")

	sentNo := languagetool.AnalyzePlain("hello")
	out2 := ars[0].Replace(sentNo)
	for _, tok := range out2.GetTokens() {
		if tok != nil && tok.GetToken() == "hello" {
			require.False(t, tok.IsImmunized(), "non-X9 must be filtered out")
		}
	}
}
