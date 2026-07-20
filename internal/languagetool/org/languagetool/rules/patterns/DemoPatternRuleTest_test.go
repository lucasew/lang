package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDemoPatternRule_Rules(t *testing.T) {
	// Demo language pattern: foo bar
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("foo", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("bar", nil, nil), 4),
	}
	sent := testSentence(toks...)
	rule := NewPatternRule("DEMO_RULE", "xx",
		[]*PatternToken{Token("foo"), Token("bar")},
		"demo", "found", "")
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

func TestDemoPatternRule_GrammarRulesFromXML2(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="DEMO_XML" name="from xml">
      <pattern><token>hello</token></pattern>
      <message>hi</message>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "demo.xml", "xx")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "DEMO_XML", rules[0].ID)
}
