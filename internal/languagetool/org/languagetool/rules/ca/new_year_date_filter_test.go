package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewYearDateFilter_ShouldFlag(t *testing.T) {
	jan := true
	y := 2024
	f := &NewYearDateFilter{ForceJanuary: &jan, ForceYear: &y}
	require.True(t, f.ShouldFlag(2023, 1))
	require.False(t, f.ShouldFlag(2023, 12))
	require.False(t, f.ShouldFlag(2024, 1))
}
