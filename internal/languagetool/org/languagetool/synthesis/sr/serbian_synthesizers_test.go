package sr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerbianSynthesizers(t *testing.T) {
	require.Equal(t, EkavianSynthDict, NewSerbianSynthesizer(nil).ResourceFileName)
	require.Equal(t, EkavianSynthDict, NewEkavianSynthesizer(nil).ResourceFileName)
	require.Equal(t, JekavianSynthDict, NewJekavianSynthesizer(nil).ResourceFileName)
}
