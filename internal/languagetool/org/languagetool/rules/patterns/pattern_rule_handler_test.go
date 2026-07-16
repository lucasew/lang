package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternRuleHandlerDemoRules(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="xx">
  <unification feature="case_sensitivity">
    <equivalence type="lowercase">
      <token regexp="yes">\p{Ll}+</token>
    </equivalence>
  </unification>
  <category id="MISC" name="misc">
    <rule id="DEMO_RULE" name="Find foo bar">
      <pattern case_sensitive="no">
        <token>foo</token>
        <token>bar</token>
      </pattern>
      <message>Did you mean something?</message>
    </rule>
    <rule id="REGEX_DEMO" name="regex">
      <regexp>(fo[ou]) bar</regexp>
      <message>msg <suggestion>\1 baz</suggestion></message>
    </rule>
  </category>
</rules>`
	h := NewPatternRuleHandler("grammar.xml", "xx")
	require.NoError(t, h.ParseString(xml))
	require.Contains(t, h.Categories, "MISC")
	require.Len(t, h.LoadedPatternRules, 1)
	require.Equal(t, "DEMO_RULE", h.LoadedPatternRules[0].ID)
	require.Len(t, h.LoadedRegexRules, 1)
	require.Equal(t, "REGEX_DEMO", h.LoadedRegexRules[0].ID)
	// unifier config
	types := h.UnifierConfiguration.GetEquivalenceTypes()
	require.NotEmpty(t, types)
}

func TestPatternRuleCheck(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <category id="MISC" name="misc">
    <rule id="DEMO_RULE" name="Find foo bar">
      <pattern case_sensitive="no">
        <token>foo</token>
        <token>bar</token>
      </pattern>
      <message>found foo bar</message>
    </rule>
  </category>
</rules>`
	h := NewPatternRuleHandler("g.xml", "en")
	require.NoError(t, h.ParseString(xml))
	chk := NewPatternRuleCheck().FromHandler(h)
	// word tokenizer may split "foo bar" into foo, space, bar
	matches, err := chk.Check("This is foo bar today")
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	require.Equal(t, "DEMO_RULE", matches[0].GetRule().(interface{ GetID() string }).GetID())
}
