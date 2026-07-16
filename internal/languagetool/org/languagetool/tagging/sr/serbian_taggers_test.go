package sr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerbianTaggers(t *testing.T) {
	require.Equal(t, EkavianDictionaryPath, NewSerbianTagger(nil).GetDictionaryPath())
	require.Equal(t, EkavianDictionaryPath, NewEkavianTagger(nil).GetDictionaryPath())
	require.Equal(t, JekavianDictionaryPath, NewJekavianTagger(nil).GetDictionaryPath())
}
