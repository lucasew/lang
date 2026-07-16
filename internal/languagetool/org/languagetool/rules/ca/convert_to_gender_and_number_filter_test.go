package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitGenderAndNumber(t *testing.T) {
	s := SplitGenderAndNumber("NCMS000")
	require.NotNil(t, s)
	require.Equal(t, "NC", s.Prefix)
	require.Equal(t, "M", s.Gender)
	require.Equal(t, "S", s.Number)

	// verb: VMP00SM0 style — prefix V.P.., then number then gender
	s = SplitGenderAndNumber("VMP00SM0")
	require.NotNil(t, s)
	require.True(t, s.Prefix[0] == 'V')
	// group2=0 group3=0 for VMP00SM0? Pattern (V.P..)(.)(.).*
	// VMP00 SM0 → prefix VMP00, g2=S, g3=M, suffix=0 — wait
	// V.P.. = V + any + P + any + any = VMP00 (5 chars)
	// Actually V.P.. means V, ., P, ., . → 5 chars: V M P 0 0
	// then (.)(.) = S M, suffix 0
	// for V: gender=g3=M, number=g2=S
	require.Equal(t, "M", s.Gender)
	require.Equal(t, "S", s.Number)
}

func TestDesiredPostag(t *testing.T) {
	f := NewConvertToGenderAndNumberFilter()
	s := SplitGenderAndNumber("NCMS000")
	got := f.DesiredPostag(s, "F", "P")
	require.Contains(t, got, "F")
	require.Contains(t, got, "P")
}

func TestBoToBonAndIgnore(t *testing.T) {
	require.Equal(t, "bon", BoToBon("bo"))
	require.True(t, ShouldIgnoreForm("mes"))
	require.True(t, IsPostagException("NP00000"))
}
