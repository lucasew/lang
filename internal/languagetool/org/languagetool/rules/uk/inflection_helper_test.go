package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInflection_Equals(t *testing.T) {
	a := Inflection{Gender: "m", Case: "v_naz"}
	b := Inflection{Gender: "s", Case: "v_naz"}
	require.True(t, a.Equals(b))
	require.False(t, a.Equals(Inflection{Gender: "m", Case: "v_rod"}))
}

func TestInflection_CompareTo(t *testing.T) {
	m := Inflection{Gender: "m", Case: "v_naz"}
	f := Inflection{Gender: "f", Case: "v_naz"}
	require.True(t, m.CompareTo(f) < 0)
}
