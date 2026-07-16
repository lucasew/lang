package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisambiguationRuleHandler(t *testing.T) {
	xml := `<?xml version="1.0"?>
	<rules>
	  <rulegroup id="G1" name="group">
	    <rule id="R1" name="rule one">
	      <pattern>
	        <token>foo</token>
	        <token regexp="yes">bar.*</token>
	      </pattern>
	      <disambig action="add" postag="NN">
	        <wd lemma="foo" pos="NN">foo</wd>
	      </disambig>
	      <example type="untouched">keep this</example>
	      <example>see &lt;marker&gt;foo&lt;/marker&gt;</example>
	    </rule>
	  </rulegroup>
	  <rule id="R2" name="solo">
	    <pattern><token>x</token></pattern>
	    <disambig action="filter" postag="VB"/>
	  </rule>
	</rules>`
	h := NewDisambiguationRuleHandler("en", "disambiguation.xml")
	require.NoError(t, h.Parse(strings.NewReader(xml)))
	require.Len(t, h.GetRules(), 2)
	r1 := h.GetRules()[0]
	require.Equal(t, "R1", r1.GetID())
	require.Equal(t, ActionAdd, r1.Action)
	require.NotEmpty(t, r1.NewTokenReadings)
	require.NotEmpty(t, r1.GetUntouchedExamples())
	require.Equal(t, "R2", h.GetRules()[1].GetID())
}
