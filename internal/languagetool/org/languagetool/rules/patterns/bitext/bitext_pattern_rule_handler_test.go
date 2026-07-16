package bitext

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitextPatternRuleHandler(t *testing.T) {
	xml := `<?xml version="1.0"?>
	<rules targetLang="de">
	  <rule id="H1" name="handler rule">
	    <source lang="en">
	      <pattern><token>gift</token></pattern>
	    </source>
	    <target>
	      <pattern><token>Gift</token></pattern>
	    </target>
	    <message>false friend</message>
	  </rule>
	</rules>`
	h := NewBitextPatternRuleHandler()
	require.NoError(t, h.Parse(strings.NewReader(xml)))
	require.Equal(t, "de", h.TargetLang)
	require.Len(t, h.GetBitextRules(), 1)
	require.Equal(t, "H1", h.GetBitextRules()[0].GetID())
	require.Equal(t, "en", h.GetBitextRules()[0].GetSourceLanguage())
}
