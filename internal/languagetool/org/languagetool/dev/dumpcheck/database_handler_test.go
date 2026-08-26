package dumpcheck

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestDatabaseHandler_StoresMatches(t *testing.T) {
	h := NewDatabaseHandler(0, 0)
	h.CategoryOf = func(rule any) string { return "Grammar" }
	h.DescriptionOf = func(rule any) string { return "desc" }
	sent := NewSentence("Hello wrong word here.", "wikipedia", "Title", "http://example.org", 1)
	m := rules.NewRuleMatch(idRule{"R1"}, nil, 6, 11, "bad word")
	require.NoError(t, h.HandleResult(sent, []*rules.RuleMatch{m}, "en"))
	require.Len(t, h.Matches, 1)
	row := h.Matches[0]
	require.Equal(t, "en", row.LanguageCode)
	require.Equal(t, "R1", row.RuleID)
	require.Equal(t, "Grammar", row.RuleCategory)
	require.Equal(t, "desc", row.RuleDescription)
	require.Equal(t, "wikipedia", row.SourceType)
	require.Equal(t, "http://example.org", row.SourceURI)
	require.Contains(t, row.ErrorContext, MarkerStart)
	require.Contains(t, row.ErrorContext, "wrong")
	require.NoError(t, h.Close())
}
