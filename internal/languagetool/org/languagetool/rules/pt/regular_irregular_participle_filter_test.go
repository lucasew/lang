package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRegularParticiple(t *testing.T) {
	require.True(t, IsRegularParticiple("assado"))
	require.True(t, IsRegularParticiple("assadas"))
	require.False(t, IsRegularParticiple("aceite"))
}

func TestRegularIrregularParticipleFilter_Suggest(t *testing.T) {
	f := NewRegularIrregularParticipleFilter()
	got := f.Suggest("RegularToIrregular", "aceitado", []string{"aceite", "aceitado"}, "{suggestion}")
	require.Equal(t, "aceite", got)
	got = f.Suggest("IrregularToRegular", "aceite", []string{"aceite", "aceitado"}, "{suggestion}")
	require.Equal(t, "aceitado", got)
}
