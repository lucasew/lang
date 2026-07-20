package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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

func TestPatternRuleHandler_ToneTagsAndGoalSpecific(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <category id="C" name="Cat" tone_tags="clarity formal">
    <rule id="R1" name="n" is_goal_specific="yes" tone_tags="professional">
      <pattern><token>foo</token></pattern>
      <message>m</message>
    </rule>
    <rulegroup id="G" tone_tags="informal">
      <rule id="R2" name="n2">
        <pattern><token>bar</token></pattern>
        <message>m2</message>
      </rule>
    </rulegroup>
  </category>
</rules>`
	h := NewPatternRuleHandler("test.xml", "en")
	require.NoError(t, h.ParseString(xml))
	require.Len(t, h.LoadedPatternRules, 2)
	r1 := h.LoadedPatternRules[0]
	require.True(t, r1.GoalSpecific)
	// rule + category tones (order: rule, group, category — R1 has no group)
	require.Contains(t, r1.ToneTags, languagetool.ToneProfessional)
	require.Contains(t, r1.ToneTags, languagetool.ToneClarity)
	require.Contains(t, r1.ToneTags, languagetool.ToneFormal)
	r2 := h.LoadedPatternRules[1]
	require.False(t, r2.GoalSpecific)
	require.Contains(t, r2.ToneTags, languagetool.ToneInformal)
	require.Contains(t, r2.ToneTags, languagetool.ToneClarity)
}
