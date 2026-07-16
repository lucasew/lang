package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternRuleXmlCreator_ToXML(t *testing.T) {
	rule := NewPatternRule("DEMO", "en",
		[]*PatternToken{Token("foo"), TokenRegex("b.r")},
		"Demo rule", "msg <suggestion>x</suggestion>", "short")
	xml := NewPatternRuleXmlCreator().ToXMLFromRule(rule)
	require.Contains(t, xml, `id="DEMO"`)
	require.Contains(t, xml, "<token>foo</token>")
	require.Contains(t, xml, `regexp="yes"`)
}

func TestPatternRuleXmlCreator_ToXMLWithRuleGroup(t *testing.T) {
	rule := NewPatternRule("G1", "en", []*PatternToken{Token("a")}, "n", "m", "")
	xml := NewPatternRuleXmlCreator().ToXMLFromRule(rule)
	require.Contains(t, xml, `id="G1"`)
	require.Contains(t, xml, "</rule>")
}

func TestPatternRuleXmlCreator_ToXMLWithRuleGroupAndSubId1(t *testing.T) {
	// SubId lives on AbstractPatternRule; PatternRule XML uses id only.
	apr := NewAbstractPatternRule("G1", "n", "en", []*PatternToken{Token("a")}, false)
	apr.SubID = "1"
	require.Equal(t, "G1[1]", apr.GetFullId())
}

func TestPatternRuleXmlCreator_ToXMLWithRuleGroupAndSubId2(t *testing.T) {
	apr := NewAbstractPatternRule("G1", "n", "en", []*PatternToken{Token("a")}, false)
	apr.SubID = "2"
	require.Equal(t, "G1[2]", apr.GetFullId())
}

func TestPatternRuleXmlCreator_ToXMLWithAntiPattern(t *testing.T) {
	rule := NewPatternRule("AP", "en", []*PatternToken{Token("bad")}, "n", "m", "")
	xml := NewPatternRuleXmlCreator().ToXMLFromRule(rule)
	require.Contains(t, xml, "bad")
}

func TestPatternRuleXmlCreator_ToXMLInvalidRuleId(t *testing.T) {
	require.Empty(t, NewPatternRuleXmlCreator().ToXMLFromRule(nil))
	rule := NewPatternRule("", "en", []*PatternToken{Token("x")}, "n", "m", "")
	xml := NewPatternRuleXmlCreator().ToXMLFromRule(rule)
	require.Contains(t, xml, `id=""`)
}
