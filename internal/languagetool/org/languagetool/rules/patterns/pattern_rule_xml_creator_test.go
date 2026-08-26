package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternRuleXmlCreator(t *testing.T) {
	rule := NewPatternRule("DEMO", "en",
		[]*PatternToken{Token("foo"), TokenRegex("b.r")},
		"Demo rule", "msg <suggestion>x</suggestion>", "short")
	xml := NewPatternRuleXmlCreator().ToXMLFromRule(rule)
	require.Contains(t, xml, `id="DEMO"`)
	require.Contains(t, xml, "<token>foo</token>")
	require.Contains(t, xml, `regexp="yes"`)
	require.Contains(t, xml, "<message>")
	require.Contains(t, xml, "</rule>")
}
