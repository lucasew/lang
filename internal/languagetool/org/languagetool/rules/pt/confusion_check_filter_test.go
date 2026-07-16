package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfusionCheckFilter_Embedded(t *testing.T) {
	f := NewConfusionCheckFilter()
	require.NotEmpty(t, f.Pairs)
}
