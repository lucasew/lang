package filters

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestArabicDMYDateCheckFilter_RejectsYearMonthDayKeys(t *testing.T) {
	// Java: containsKey("year"|"month"|"day") → RuntimeException with simple class name.
	f := NewArabicDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 1, "msg")
	const want = "Set only 'weekDay' and 'date' for ArabicDMYDateCheckFilter"
	require.PanicsWithValue(t, want, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "12-03-2022", "year": "2022", "weekDay": "السبت"}, 0, nil, nil)
	})
	require.PanicsWithValue(t, want, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "12-03-2022", "month": "3", "weekDay": "السبت"}, 0, nil, nil)
	})
	require.PanicsWithValue(t, want, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "12-03-2022", "day": "12", "weekDay": "السبت"}, 0, nil, nil)
	})
}

func TestArabicDMYDateCheckFilter_MissingDatePanics(t *testing.T) {
	// Java getRequired("date") → IllegalArgumentException "Missing key 'date'"
	f := NewArabicDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 1, "msg")
	require.PanicsWithValue(t, "Missing key 'date'", func() {
		f.AcceptRuleMatch(m, map[string]string{"weekDay": "السبت"}, 0, nil, nil)
	})
}

func TestArabicDMYDateCheckFilter_BadFormatPanics(t *testing.T) {
	// Java: parts.length != 3 → RuntimeException with exact wording + quoted dateString.
	f := NewArabicDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 1, "msg")
	require.PanicsWithValue(t, "Expected date in format 'dd-mm-yyyy': 'bad'", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "bad", "weekDay": "السبت"}, 0, nil, nil)
	})
	require.PanicsWithValue(t, "Expected date in format 'dd-mm-yyyy': '12-03'", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "12-03", "weekDay": "السبت"}, 0, nil, nil)
	})
	// Java split discards trailing empty: "12-03-" → ["12","03"] length 2
	require.PanicsWithValue(t, "Expected date in format 'dd-mm-yyyy': '12-03-'", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "12-03-", "weekDay": "السبت"}, 0, nil, nil)
	})
	// Four parts
	require.PanicsWithValue(t, "Expected date in format 'dd-mm-yyyy': '12-03-2022-1'", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "12-03-2022-1", "weekDay": "السبت"}, 0, nil, nil)
	})
	// Empty date present: getRequired returns ""; split → {""} length 1
	require.PanicsWithValue(t, "Expected date in format 'dd-mm-yyyy': ''", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "", "weekDay": "السبت"}, 0, nil, nil)
	})
}

func TestArabicDMYDateCheckFilter_CorrectWeekdayNil(t *testing.T) {
	// 12-03-2022 is Saturday → السبت; correct weekDay drops the match (null).
	f := NewArabicDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "12-03-2022",
		"weekDay": "السبت",
	}, 0, nil, nil)
	require.Nil(t, out)
}

func TestArabicDMYDateCheckFilter_IncorrectWeekdayKeepsMatch(t *testing.T) {
	// 12-03-2022 Saturday; claim الأحد (Sunday) → keep match; message uses ar LONG names.
	f := NewArabicDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "{realDay}/{day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "12-03-2022",
		"weekDay": "الأحد",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "السبت") // real
	require.Contains(t, out.GetMessage(), "الأحد") // claimed
}

func TestArabicDMYDateCheckFilter_MutatesArgsForSuper(t *testing.T) {
	// Java args.put mutates the same map so super sees day/month/year.
	f := NewArabicDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	args := map[string]string{"date": "12-03-2022", "weekDay": "السبت"}
	_ = f.AcceptRuleMatch(m, args, 0, nil, nil)
	require.Equal(t, "12", args["day"])
	require.Equal(t, "03", args["month"])
	require.Equal(t, "2022", args["year"])
}

func TestArabicDMYDateCheckFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ar.filters.ArabicDMYDateCheckFilter"))
}
