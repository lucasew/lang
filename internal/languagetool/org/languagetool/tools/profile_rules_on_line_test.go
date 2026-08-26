package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of Tools.profileRulesOnLine control flow (tokenize + sum match lengths).
func TestProfileRulesOnLine(t *testing.T) {
	// Fake: two sentences; each "hits" once if contains "x"
	tok := func(s string) []string {
		parts := strings.Split(s, ". ")
		return parts
	}
	match := func(sentence string) int {
		if strings.Contains(sentence, "x") {
			return 1
		}
		return 0
	}
	n := ProfileRulesOnLine("a x. b. c x", tok, match)
	require.Equal(t, 2, n)
	require.Equal(t, 0, ProfileRulesOnLine("none", tok, match))
	require.Equal(t, 0, ProfileRulesOnLine("x", nil, match))
}
