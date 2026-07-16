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
