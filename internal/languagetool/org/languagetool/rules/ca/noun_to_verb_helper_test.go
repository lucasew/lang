package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNounToVerb(t *testing.T) {
	require.Equal(t, "acceptar", NounToVerb("acceptació"))
	require.Equal(t, "crear", NounToVerb("creació"))
	require.Equal(t, "", NounToVerb("xyz"))
}
