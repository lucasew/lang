package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiacriticsCheckFilter_Embedded(t *testing.T) {
	f := NewDiacriticsCheckFilter()
	require.NotEmpty(t, f.Pairs)
}
