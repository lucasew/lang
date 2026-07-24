package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullAsEmptyEmptyAsNull(t *testing.T) {
	require.Equal(t, "", NullAsEmpty(nil))
	s := "x"
	require.Equal(t, "x", NullAsEmpty(&s))
	require.Nil(t, EmptyAsNull(""))
	p := EmptyAsNull("a")
	require.NotNil(t, p)
	require.Equal(t, "a", *p)
	require.Equal(t, "m", CoalesceURL("m", "r"))
	require.Equal(t, "r", CoalesceURL("", "r"))
}
