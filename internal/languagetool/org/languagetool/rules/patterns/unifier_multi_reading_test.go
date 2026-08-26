package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Regression: Java uniMatched |= isSatisfied always evaluates RHS so every
// compatible reading is kept. Go `||` short-circuit used to drop later readings.
func TestUnifier_MultipleCompatibleReadingsOnLastToken(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("number", "sg", preparePOSElement(`VER:.*\+s|PRO\-.*\-S`))
	cfg.SetEquivalence("number", "pl", preparePOSElement(`VER:.*\+p|PRO\-.*\-P`))
	cfg.SetEquivalence("persona", "first", preparePOSElement(`.*[-\+]1[-\+].*`))
	cfg.SetEquivalence("persona", "second", preparePOSElement(`.*[-\+]2[-\+].*`))
	cfg.SetEquivalence("persona", "third", preparePOSElement(`.*[-\+]3[-\+].*`))
	uni := cfg.CreateUnifier()
	feats := map[string][]string{"number": nil, "persona": nil}
	p := func(s string) *string { return &s }

	require.True(t, uni.IsUnified(languagetool.NewAnalyzedToken("tu", p("PRO-PERS-2-F-S"), p("tu")), feats, false))
	require.True(t, uni.IsUnified(languagetool.NewAnalyzedToken("tu", p("PRO-PERS-2-M-S"), p("tu")), feats, true))
	require.True(t, uni.IsUnified(languagetool.NewAnalyzedToken("ami", p("VER:ind+pres+2+s"), p("amare")), feats, false))
	require.True(t, uni.IsUnified(languagetool.NewAnalyzedToken("ami", p("VER:sub+pres+2+s"), p("amare")), feats, true))

	fu := uni.GetFinalUnified()
	require.NotNil(t, fu)
	require.Len(t, fu, 2)
	var tags []string
	for _, rd := range fu[1].GetReadings() {
		if rd != nil && rd.GetPOSTag() != nil {
			tags = append(tags, *rd.GetPOSTag())
		}
	}
	require.ElementsMatch(t, []string{"VER:ind+pres+2+s", "VER:sub+pres+2+s"}, tags)
}
