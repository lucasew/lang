package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDynamicLanguage(t *testing.T) {
	d := NewDynamicLanguage("English", "en-US", "/dicts/en.dict")
	require.Equal(t, "en", d.GetShortCode())
	require.Equal(t, "en-US", d.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "English", d.GetName())
}
