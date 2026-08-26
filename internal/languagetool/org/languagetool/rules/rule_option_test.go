package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleOption_Codec(t *testing.T) {
	require.Equal(t, "i42", ObjectToString(42))
	require.Equal(t, "btrue", ObjectToString(true))
	require.Equal(t, "shello", ObjectToString("hello"))
	o, err := StringToObject("i7")
	require.NoError(t, err)
	require.Equal(t, 7, o)
	// legacy plain int
	o, err = StringToObject("99")
	require.NoError(t, err)
	require.Equal(t, 99, o)
	arr, err := StringToObjects("i1;bfalse;sxy")
	require.NoError(t, err)
	require.Equal(t, []any{1, false, "xy"}, arr)
	require.Equal(t, "i1;i2", ObjectsToString([]any{1, 2}))
}

func TestRuleOption_Defaults(t *testing.T) {
	ro := NewRuleOption(true, "Enable", nil, nil)
	require.Equal(t, 0, ro.MinConfigurableValue)
	require.Equal(t, 100, ro.MaxConfigurableValue)
	ro2 := NewRuleOption(5, "n", 1, 10)
	require.Equal(t, 1, ro2.MinConfigurableValue)
	require.Equal(t, 10, ro2.MaxConfigurableValue)
}
