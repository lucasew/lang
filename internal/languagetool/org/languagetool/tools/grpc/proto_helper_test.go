package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtoHelper(t *testing.T) {
	require.Equal(t, "", NullAsEmpty(nil))
	s := "x"
	require.Equal(t, "x", NullAsEmpty(&s))
	require.Nil(t, EmptyAsNull(""))
	require.Equal(t, "a", *EmptyAsNull("a"))
	require.Equal(t, "m", CoalesceURL("m", "r"))
	require.Equal(t, "r", CoalesceURL("", "r"))

	rd := NewRuleData("R", "1", "d")
	require.Equal(t, "R", rd.GetID())
	require.Equal(t, "1", rd.GetSubID())
}
