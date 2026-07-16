package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDMYDateCheckFilter_PrepareArgs(t *testing.T) {
	f := NewDMYDateCheckFilter()
	_, err := f.PrepareArgs(map[string]string{"year": "2020"})
	require.Error(t, err)
	out, err := f.PrepareArgs(map[string]string{"date": "15-01-2020", "weekDay": "mercredi"})
	require.NoError(t, err)
	require.Equal(t, "15", out["day"])
	require.Equal(t, "01", out["month"])
	require.Equal(t, "2020", out["year"])
}
