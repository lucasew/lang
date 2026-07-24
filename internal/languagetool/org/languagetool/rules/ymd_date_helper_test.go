package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYMDDateHelper_ParseDate(t *testing.T) {
	h := NewYMDDateHelper()
	out, err := h.ParseDate(map[string]string{"date": "2014-08-23", "weekDay": "Samstag"})
	require.NoError(t, err)
	require.Equal(t, "2014", out["year"])
	require.Equal(t, "08", out["month"])
	require.Equal(t, "23", out["day"])
	require.Equal(t, "Samstag", out["weekDay"])
	_, err = h.ParseDate(map[string]string{})
	require.Error(t, err)
}

func TestYMDDateHelper_CorrectDate(t *testing.T) {
	h := NewYMDDateHelper()
	m := NewRuleMatch(NewFakeRule("Y"), nil, 0, 4, "use {realDate} instead")
	out := h.CorrectDate(m, map[string]string{"year": "2013", "month": "03", "day": "15"})
	require.NotNil(t, out)
	require.Equal(t, "use 2014-03-15 instead", out.GetMessage())
}
