package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYMDDateCheckFilter_PrepareArgs(t *testing.T) {
	f := NewYMDDateCheckFilter()
	_, err := f.PrepareArgs(map[string]string{"year": "2020", "weekDay": "1"})
	require.Error(t, err)
	out, err := f.PrepareArgs(map[string]string{"date": "2020-01-15", "weekDay": "4"})
	require.NoError(t, err)
	require.Equal(t, "2020", out["year"])
	require.Equal(t, "01", out["month"])
	require.Equal(t, "15", out["day"])
}
