package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestFalseFriendPatternRule(t *testing.T) {
	r := NewFalseFriendPatternRule("FF", "en", []*PatternToken{Token("gift")}, "desc", "msg", "short")
	require.Equal(t, "FF", r.GetID())
	require.True(t, r.HasTag(rules.TagPicky))
	require.Len(t, r.Tokens, 1)
}
