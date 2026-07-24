package broker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultClassBroker(t *testing.T) {
	b := NewDefaultClassBroker()
	require.NotNil(t, b)
	b.Register("org.example.Foo", func() any { return 42 })
	v, err := b.ForName("org.example.Foo")
	require.NoError(t, err)
	require.Equal(t, 42, v)
}
