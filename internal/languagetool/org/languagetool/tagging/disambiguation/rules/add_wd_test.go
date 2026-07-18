package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDisambiguationLoader_AddWdPCT(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <rule id="UNKNOWN_PCT" name="pct">
    <pattern>
      <token regexp="yes">[\.,;:…!\?]</token>
    </pattern>
    <disambig action="add">
      <wd pos="PCT"/>
    </disambig>
  </rule>
</rules>`
	loader := NewDisambiguationRuleLoader()
	rules, err := loader.GetRulesFromString(xml, "en", "test")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, ActionAdd, rules[0].Action)
	require.Len(t, rules[0].NewTokenReadings, 1)
	require.NotNil(t, rules[0].NewTokenReadings[0].GetPOSTag())
	require.Equal(t, "PCT", *rules[0].NewTokenReadings[0].GetPOSTag())

	// Full sentence like Analyze: SENT_START + comma
	startTag := "SENT_START"
	start := languagetool.NewAnalyzedToken("", &startTag, nil)
	commaPos := ","
	comma := languagetool.NewAnalyzedToken(",", &commaPos, &commaPos)
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(start, 0),
		languagetool.NewAnalyzedTokenReadingsAt(comma, 0),
	}
	sent := languagetool.NewAnalyzedSentence(tokens)
	matcher := patterns.NewPatternRuleMatcher(rules[0].AbstractTokenBasedRule)
	ms, err := matcher.Match(sent)
	require.NoError(t, err)
	t.Logf("matches=%d", len(ms))
	for _, m := range ms {
		t.Logf("from=%d to=%d", m.FromPos, m.ToPos)
	}
	out := rules[0].Replace(sent)
	require.NotNil(t, out)
	for _, tok := range out.GetTokensWithoutWhitespace() {
		for _, rd := range tok.GetReadings() {
			pos := ""
			if rd.GetPOSTag() != nil {
				pos = *rd.GetPOSTag()
			}
			t.Logf("tok=%q pos=%q", tok.GetToken(), pos)
		}
	}
	tok := out.GetTokensWithoutWhitespace()
	// find comma
	var hasPCT bool
	for _, tk := range tok {
		if tk.GetToken() != "," {
			continue
		}
		for _, rd := range tk.GetReadings() {
			if rd.GetPOSTag() != nil && *rd.GetPOSTag() == "PCT" {
				hasPCT = true
			}
		}
	}
	require.True(t, hasPCT, "expected PCT reading on comma")
}
