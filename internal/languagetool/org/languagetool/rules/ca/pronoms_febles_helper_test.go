package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransformDavant(t *testing.T) {
	// em + vowel → m'
	require.Equal(t, "m'", TransformDavant("em", "ajuda"))
	// em + consonant → em
	require.Equal(t, "em ", TransformDavant("em", "porta"))
	// es + senten → se
	require.Equal(t, "se ", TransformDavant("es", "senten"))
}

func TestTransformDarrere(t *testing.T) {
	// after vowel-ending: apostrophe form
	got := TransformDarrere("em", "mira")
	require.NotEmpty(t, got)
	// after consonant
	got2 := TransformDarrere("em", "parlar")
	require.NotEmpty(t, got2)
}

func TestIncorrectOrders(t *testing.T) {
	require.Equal(t, "se'm ", Transform("me se", PronounDavant))
	require.Equal(t, "m'hi ", Transform("mi", PronounDavant))
}

func TestGetReflexiveAndDative(t *testing.T) {
	require.Equal(t, "em", GetReflexivePronoun("1S"))
	require.Equal(t, "li", GetDativePronoun("3S"))
	require.Equal(t, "els", GetDativePronoun("3P"))
}

func TestFixApostrophes(t *testing.T) {
	require.Equal(t, "de casa", FixApostrophes("d'casa"))
	require.Equal(t, "m'ajuda", FixApostrophes("em ajuda"))
}

func TestConvertPronounsForIntransitiveVerb(t *testing.T) {
	require.Equal(t, "se li diu", ConvertPronounsForIntransitiveVerb("se'l diu"))
}
