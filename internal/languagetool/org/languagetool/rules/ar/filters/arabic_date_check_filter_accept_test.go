package filters

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestArabicDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewArabicDateCheckFilter()
	// 27 أغسطس 2014 was Wednesday; claim الجمعة (Friday) → mismatch
	m := rules.NewRuleMatch(nil, nil, 0, 20, "اليوم {realDay} وليس {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "أغسطس", "day": "27", "weekDay": "الجمعة",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "الأربعاء")
	require.Contains(t, out.GetMessage(), "الجمعة")

	// 27 Aug 2014 is Wednesday
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "أغسطس", "day": "27", "weekDay": "الأربعاء",
	}, 0, nil, nil)
	require.Nil(t, ok)
}

func TestArabicDateCheckFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ar.filters.ArabicDateCheckFilter"))
}
