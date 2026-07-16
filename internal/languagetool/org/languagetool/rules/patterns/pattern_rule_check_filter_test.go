package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternRuleCheckAntiPatternFilter(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <category id="MISC" name="misc">
    <rule id="R1" name="r">
      <regexp mark="1">(fo.) (bar)</regexp>
      <filter class="org.languagetool.rules.patterns.RegexAntiPatternFilter" args="antipatterns:fou"/>
      <message>msg</message>
    </rule>
  </category>
</rules>`
	h := NewPatternRuleHandler("g.xml", "en")
	require.NoError(t, h.ParseString(xml))
	chk := NewPatternRuleCheck().FromHandler(h)
	// fou bar should be filtered; fox bar should match
	m1, err := chk.Check("This is fou bar stuff")
	require.NoError(t, err)
	m2, err := chk.Check("This is fox bar stuff")
	require.NoError(t, err)
	require.Empty(t, m1)
	require.NotEmpty(t, m2)
}
