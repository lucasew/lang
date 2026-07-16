package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUppercaseNounReadingFilter(t *testing.T) {
	f := NewUppercaseNounReadingFilter()
	require.True(t, f.Accept("stand"))
	f.HasNounReading = func(u string) bool { return u == "Stand" }
	require.True(t, f.Accept("stand"))
	require.False(t, f.Accept("laufen"))
}
