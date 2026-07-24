package dumpcheck

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestCSVHandler_MatchAndNoMatch(t *testing.T) {
	var buf strings.Builder
	h := NewCSVHandler(&buf, 0, 0)
	sent := NewSentence("Hello wrong word here.", "plain", "T", "", 1)
	m := rules.NewRuleMatch(idRule{"DEMO"}, nil, 6, 11, "bad")
	require.NoError(t, h.HandleResult(sent, []*rules.RuleMatch{m}, "en"))
	require.Contains(t, buf.String(), "MATCH\tDEMO\t")
	require.Contains(t, buf.String(), "__wrong__")

	buf.Reset()
	require.NoError(t, h.HandleResult(sent, nil, "en"))
	require.Contains(t, buf.String(), "NOMATCH\t\t")
}

func TestNoTabs(t *testing.T) {
	require.Equal(t, `a\tb`, noTabs("a\tb"))
}
