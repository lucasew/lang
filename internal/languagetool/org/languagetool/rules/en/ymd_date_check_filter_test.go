package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYMDDateCheckFilter_PrepareArgs(t *testing.T) {
	f := NewYMDDateCheckFilter()
	_, err := f.PrepareArgs(map[string]string{"month": "1"})
	require.Error(t, err)
	out, err := f.PrepareArgs(map[string]string{"date": "1999-12-31", "weekDay": "6"})
	require.NoError(t, err)
	require.Equal(t, "1999", out["year"])
	require.Equal(t, "12", out["month"])
	require.Equal(t, "31", out["day"])
}
