package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.Range (+ FragmentWithLanguage colocated smoke).

func TestRange(t *testing.T) {
	r := NewRange(0, 5, "en")
	require.Equal(t, 0, r.GetFromPos())
	require.Equal(t, 5, r.GetToPos())
	require.Equal(t, "en", r.GetLang())
	require.True(t, r.Equal(NewRange(0, 5, "en")))
	require.False(t, r.Equal(NewRange(0, 5, "de")))
	require.False(t, r.Equal(NewRange(1, 5, "en")))
	require.False(t, r.Equal(NewRange(0, 6, "en")))
	// Java Objects.requireNonNull(lang) allows empty string
	emptyLang := NewRange(0, 1, "")
	require.Equal(t, "", emptyLang.GetLang())
}

func TestFragmentWithLanguage(t *testing.T) {
	f := NewFragmentWithLanguage("fr", "bonjour")
	require.Equal(t, "fr", f.GetLangCode())
	require.Equal(t, "bonjour", f.GetFragment())
	require.Equal(t, "| fr: bonjour |", f.String())
	// Java allows empty strings (only null is rejected)
	empty := NewFragmentWithLanguage("", "")
	require.Equal(t, "| :  |", empty.String())
}
