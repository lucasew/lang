package tools

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrappedVoid(t *testing.T) {
	called := false
	w := WrappedVoid(func() error {
		called = true
		return nil
	})
	require.NoError(t, w.Call())
	require.True(t, called)
	require.Error(t, WrappedVoid(func() error { return errors.New("x") }).Call())
}

func TestWrappedValue(t *testing.T) {
	w := WrappedValue[int](func() (int, error) { return 42, nil })
	v, err := w.Call()
	require.NoError(t, err)
	require.Equal(t, 42, v)
}
