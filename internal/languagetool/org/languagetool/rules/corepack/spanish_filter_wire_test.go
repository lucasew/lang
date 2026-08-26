package corepack_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/stretchr/testify/require"
)

func TestRegister_SpanishFilterRuleMatchesWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("es")
	corepack.Register(lt, "es")
	require.NotNil(t, lt.FilterRuleMatches, "language init must wire Spanish.filterRuleMatches")
	out := lt.FilterRuleMatches([]languagetool.LocalMatch{
		{RuleID: "AI_ES_GGEC_X", Suggestions: []string{"sólo"}},
		{RuleID: "OTHER", Suggestions: []string{"ok"}},
	})
	require.Len(t, out, 1)
	require.Equal(t, "OTHER", out[0].RuleID)
}

func TestRegister_SpanishPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("es")
	corepack.Register(lt, "es")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 20, lt.PriorityForId("TYPOGRAPHY"))
	require.Equal(t, -300, lt.PriorityForId("AI_ES_GGEC_REPLACEMENT_OTHER"))
	require.Equal(t, 0, lt.PriorityForId("AI_ES_GGEC_OTHER"))
}
