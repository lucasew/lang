package patterns

import (
	"testing"
	"runtime"
	"path/filepath"
	"os"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPatternRuleLoader_GetRules(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="R1" name="Rule one">
      <pattern>
        <token>foo</token>
        <token regexp="yes">b.r</token>
      </pattern>
      <message>bad</message>
      <short>s</short>
    </rule>
    <rule id="R2" name="Rule two">
      <pattern><token>x</token></pattern>
      <message>m2</message>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "test.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	require.Equal(t, "R1", rules[0].ID)
	require.Equal(t, "R2", rules[1].ID)
	require.Len(t, rules[0].PatternTokens, 2)
	require.True(t, rules[0].PatternTokens[1].Regexp)
}

// Twin of PatternRuleLoaderTest.testPremiumXmlFlag (official xx fixtures).
func TestPatternRuleLoader_PremiumXmlFlag(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// .../rules/patterns → repo root 6 levels up
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..", ".."))
	nonPrem := filepath.Join(root, "inspiration/languagetool/languagetool-core/src/test/resources/org/languagetool/rules/xx/grammar-nonPremium.xml")
	prem := filepath.Join(root, "inspiration/languagetool/languagetool-core/src/test/resources/org/languagetool/rules/xx/grammar-premium.xml")
	if _, err := os.Stat(nonPrem); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	load := func(path string) map[string]*AbstractPatternRule {
		b, err := os.ReadFile(path)
		require.NoError(t, err)
		loaded, err := NewPatternRuleLoader().GetRulesFromString(string(b), path, "xx")
		require.NoError(t, err)
		m := map[string]*AbstractPatternRule{}
		for _, r := range loaded {
			m[r.ID] = r
		}
		return m
	}
	by := load(nonPrem)
	require.False(t, by["F-NP_C-NP_RG-NP_R-NP"].Premium)
	require.True(t, by["F-NP_C-NP_RG-NP_R-P"].Premium)
	require.False(t, by["F-NP_C-NP_RG-P_R-NP"].Premium)
	require.True(t, by["F-NP_C-NP_RG-P_R-P"].Premium)
	require.False(t, by["F-NP_C-P_RG-NP_R-NP"].Premium)
	require.False(t, by["F-NP_C-P_RG-NP_R-P"].Premium)
	require.False(t, by["F-NP_C-P_RG-P_R-NP"].Premium)
	require.True(t, by["F-NP_C-P_RG-P_R-P"].Premium)

	byP := load(prem)
	require.True(t, byP["F-P_C-P_RG-P_R-P"].Premium)
	require.False(t, byP["F-P_C-P_RG-P_R-NP"].Premium)
	require.True(t, byP["F-P_C-P_RG-NP_R-P"].Premium)
	require.False(t, byP["F-P_C-P_RG-NP_R-NP"].Premium)
	require.False(t, byP["F-P_C-NP_RG-P_R-NP"].Premium)
	require.True(t, byP["F-P_C-NP_RG-P_R-P"].Premium)
	require.False(t, byP["F-P_C-NP_RG-NP_R-NP"].Premium)
	require.True(t, byP["F-P_C-NP_RG-NP_R-P"].Premium)
}

// Twin of PatternRuleLoaderTest.testToneTagsAttribute (subset of official style.xml cases).
func TestPatternRuleLoader_ToneTagsAttribute(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="xx">
  <category id="C" name="Style">
    <rule id="Formal_Clarity_TONE_RULE" name="n" tone_tags="formal clarity">
      <pattern><token>a</token></pattern><message>m</message>
    </rule>
    <rule id="NO_TONE_RULE" name="n">
      <pattern><token>b</token></pattern><message>m</message>
    </rule>
    <rule id="CONFIDENT_ACADEMIC_SCIENTIFIC_TONE_RULE" name="n" tone_tags="confident academic scientific">
      <pattern><token>c</token></pattern><message>m</message>
    </rule>
    <rule id="PERSUASIVE_GOAL_SPECIFIC_TONE_RULE" name="n" tone_tags="persuasive" is_goal_specific="yes">
      <pattern><token>d</token></pattern><message>m</message>
    </rule>
    <rule id="PERSUASIVE_NOT_GOAL_SPECIFIC_TONE_RULE" name="n" tone_tags="persuasive" is_goal_specific="no">
      <pattern><token>e</token></pattern><message>m</message>
    </rule>
    <rule id="PICKY-CLARITY_CONFIDENT_ACADEMIC_TONE_RULE" name="n" tone_tags="clarity confident academic" tags="picky">
      <pattern><token>f</token></pattern><message>m</message>
    </rule>
  </category>
</rules>`
	loaded, err := NewPatternRuleLoader().GetRulesFromString(xml, "style.xml", "xx")
	require.NoError(t, err)
	byID := map[string]*AbstractPatternRule{}
	for _, r := range loaded {
		byID[r.ID] = r
	}
	fc := byID["Formal_Clarity_TONE_RULE"]
	require.Contains(t, fc.ToneTags, languagetool.ToneFormal)
	require.Contains(t, fc.ToneTags, languagetool.ToneClarity)
	require.Len(t, fc.ToneTags, 2)
	require.False(t, fc.GoalSpecific)

	require.Empty(t, byID["NO_TONE_RULE"].ToneTags)

	cas := byID["CONFIDENT_ACADEMIC_SCIENTIFIC_TONE_RULE"]
	require.Len(t, cas.ToneTags, 3)
	require.Contains(t, cas.ToneTags, languagetool.ToneConfident)
	require.Contains(t, cas.ToneTags, languagetool.ToneAcademic)
	require.Contains(t, cas.ToneTags, languagetool.ToneScientific)

	require.True(t, byID["PERSUASIVE_GOAL_SPECIFIC_TONE_RULE"].GoalSpecific)
	require.False(t, byID["PERSUASIVE_NOT_GOAL_SPECIFIC_TONE_RULE"].GoalSpecific)

	picky := byID["PICKY-CLARITY_CONFIDENT_ACADEMIC_TONE_RULE"]
	require.Len(t, picky.ToneTags, 3)
	require.Contains(t, picky.Tags, rules.TagPicky)
}

// Twin of PatternRuleLoaderTest.testPrioAttribute (category < group < rule).
func TestPatternRuleLoader_PrioAttribute(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="xx">
  <category id="C5" name="C" prio="5">
    <rulegroup id="RG10" prio="10">
      <rule id="CAT-PRIO-5-RG-PRIO-10-R-PRIO-15" name="n" prio="15">
        <pattern><token>a</token></pattern><message>m</message>
      </rule>
      <rule id="CAT-PRIO-5-RG-PRIO-10-R-PRIO-0" name="n">
        <pattern><token>b</token></pattern><message>m</message>
      </rule>
    </rulegroup>
    <rulegroup id="RG0">
      <rule id="CAT-PRIO-5-RG-PRIO-0-R-PRIO-0" name="n">
        <pattern><token>c</token></pattern><message>m</message>
      </rule>
    </rulegroup>
  </category>
  <category id="C0" name="C0">
    <rulegroup id="RG00">
      <rule id="CAT-PRIO-0-RG-PRIO-0-R-PRIO-0" name="n">
        <pattern><token>d</token></pattern><message>m</message>
      </rule>
    </rulegroup>
    <rule id="CAT-PRIO-0-R-PRIO-0" name="n">
      <pattern><token>e</token></pattern><message>m</message>
    </rule>
  </category>
</rules>`
	loaded, err := NewPatternRuleLoader().GetRulesFromString(xml, "grammar-withPrio.xml", "xx")
	require.NoError(t, err)
	byID := map[string]*AbstractPatternRule{}
	for _, r := range loaded {
		byID[r.ID] = r
	}
	require.Equal(t, 15, byID["CAT-PRIO-5-RG-PRIO-10-R-PRIO-15"].Priority)
	require.Equal(t, 10, byID["CAT-PRIO-5-RG-PRIO-10-R-PRIO-0"].Priority)
	require.Equal(t, 5, byID["CAT-PRIO-5-RG-PRIO-0-R-PRIO-0"].Priority)
	require.Equal(t, 0, byID["CAT-PRIO-0-RG-PRIO-0-R-PRIO-0"].Priority)
	require.Equal(t, 0, byID["CAT-PRIO-0-R-PRIO-0"].Priority)
}
