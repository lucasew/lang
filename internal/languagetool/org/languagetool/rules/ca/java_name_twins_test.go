package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJavaNameTwins(t *testing.T) {
	require.Equal(t, "CA_UNPAIRED_EXCLAMATION", NewCatalanUnpairedExclamationMarksRule(nil).GetID())
	v, ok := (NounToVerbHelper{}).VerbForNoun("ajuda")
	require.True(t, ok)
	require.Equal(t, "ajudar", v)
	require.NotEmpty(t, (PronomsFeblesHelper{}).GetReflexivePronoun("1S"))
	require.NotNil(t, NewCatalanRemoteRewriteHelper())
}
