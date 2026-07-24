package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// EnglishPartialPosTagFilter (with disambiguator) is out of claim scope for sector
// 3.C.3-NoDisambiguationEnglishPartialPosTagFilter. Keep a light registration check
// and wire-path smoke in english_partial_pos_tag_filter_wire_test.go.

func TestEnglishPartialPosTagFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.EnglishPartialPosTagFilter"))
}
