package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegexRuleFilterCreator(t *testing.T) {
	c := NewRegexRuleFilterCreator()
	f := c.GetFilter("org.languagetool.rules.patterns.RegexAntiPatternFilter")
	require.NotNil(t, f)
	require.Panics(t, func() { c.GetFilter("org.example.Missing") })
}
