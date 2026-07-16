package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfusionCheckFilter_Embedded(t *testing.T) {
	f := NewConfusionCheckFilter()
	require.NotEmpty(t, f.Pairs)
	res := f.Suggest("acaro", "NCMS000", "", "se escribe con tilde", "{suggestion}")
	require.True(t, res.OK)
	require.Equal(t, "ácaro", res.Replacement)
}
