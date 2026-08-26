package broker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapClassBroker(t *testing.T) {
	b := NewMapClassBroker()
	b.Register("org.example.Foo", func() any { return 42 })
	v, err := b.ForName("org.example.Foo")
	require.NoError(t, err)
	require.Equal(t, 42, v)
	_, err = b.ForName("missing")
	require.Error(t, err)
}
