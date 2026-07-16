package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleWithLanguage(t *testing.T) {
	rule := "RULE"
	r := NewRuleWithLanguage(rule, "en-US")
	require.Equal(t, rule, r.GetRule())
	require.Equal(t, "en-US", r.GetLanguageCode())
	require.True(t, r.Equal(NewRuleWithLanguage(rule, "en-US")))
	require.Panics(t, func() { NewRuleWithLanguage(nil, "en") })
}
