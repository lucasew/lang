package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFalseFriendRuleHandler(t *testing.T) {
	h := NewFalseFriendRuleHandler("en", "de", "''{0}'' ({1}) means {2} ({3})")
	require.True(t, h.ShouldEmitRule("en", "de", true))
	require.False(t, h.ShouldEmitRule("de", "de", true))
	require.False(t, h.ShouldEmitRule("en", "fr", true))
	require.False(t, h.ShouldEmitRule("en", "de", false))

	require.Equal(t, `"Haus", "Heim"`, FormatTranslations([]string{"Haus", "Heim"}))
	msg := h.FormatHint("gift", "English", `"Gift"`, "German")
	require.Contains(t, msg, "gift")
	require.Contains(t, msg, "English")

	h.AddSuggestions("FF_1", []string{"present", "present", "gift"})
	require.Equal(t, []string{"present", "gift"}, h.SuggestionMap["FF_1"])
}
