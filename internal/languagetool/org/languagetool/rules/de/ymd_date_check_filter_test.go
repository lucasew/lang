package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYMDDateCheckFilter_PrepareArgs(t *testing.T) {
	f := NewYMDDateCheckFilter()
	out, err := f.PrepareArgs(map[string]string{"date": "2014-08-23", "weekDay": "Samstag"})
	require.NoError(t, err)
	require.Equal(t, "2014", out["year"])
	_, err = f.PrepareArgs(map[string]string{"date": "2014-08-23", "year": "2014"})
	require.Error(t, err)
}
