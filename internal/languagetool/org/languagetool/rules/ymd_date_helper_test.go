package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYMDDateHelper_ParseDate(t *testing.T) {
	h := NewYMDDateHelper()
	args := map[string]string{"date": "2014-08-23", "weekDay": "Samstag"}
	out, err := h.ParseDate(args)
	require.NoError(t, err)
	// Java mutates the same map and returns it.
	require.Equal(t, args, out)
	require.Equal(t, "2014", out["year"])
	require.Equal(t, "08", out["month"])
	require.Equal(t, "23", out["day"])
	require.Equal(t, "Samstag", out["weekDay"])
	require.Equal(t, "2014", args["year"], "parseDate mutates caller's map in place")
}

func TestYMDDateHelper_ParseDate_MissingDate(t *testing.T) {
	h := NewYMDDateHelper()
	_, err := h.ParseDate(map[string]string{})
	require.Error(t, err)
	require.Equal(t, "Missing key 'date'", err.Error())
	// Key present with empty value: Java does not treat as missing; split yields != 3 parts.
	_, err = h.ParseDate(map[string]string{"date": ""})
	require.Error(t, err)
	require.Equal(t, "Expected date in format 'yyyy-mm-dd': ''", err.Error())
}

func TestYMDDateHelper_ParseDate_BadFormat(t *testing.T) {
	h := NewYMDDateHelper()
	_, err := h.ParseDate(map[string]string{"date": "2014/08/23"})
	require.Error(t, err)
	require.Equal(t, "Expected date in format 'yyyy-mm-dd': '2014/08/23'", err.Error())
	_, err = h.ParseDate(map[string]string{"date": "2014-08"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Expected date in format 'yyyy-mm-dd'")
}

func TestYMDDateHelper_CorrectDate(t *testing.T) {
	h := NewYMDDateHelper()
	m := NewRuleMatch(NewFakeRule("Y"), nil, 0, 4, "use {realDate} instead")
	m.SetShortMessage("short")
	m.SetType(RuleMatchTypeHint)
	out := h.CorrectDate(m, map[string]string{"year": "2013", "month": "03", "day": "15"})
	require.NotNil(t, out)
	// Java: String.format("%d-%s-%s", year+1, month, day)
	require.Equal(t, "use 2014-03-15 instead", out.GetMessage())
	require.Equal(t, "short", out.GetShortMessage())
	require.Equal(t, RuleMatchTypeHint, out.GetType())
}

func TestYMDDateHelper_CorrectDate_BadYearPanics(t *testing.T) {
	h := NewYMDDateHelper()
	m := NewRuleMatch(NewFakeRule("Y"), nil, 0, 1, "{realDate}")
	require.Panics(t, func() {
		h.CorrectDate(m, map[string]string{"year": "x", "month": "01", "day": "01"})
	})
}
