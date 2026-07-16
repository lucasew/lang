package gui

// Twin of ToolsTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTools_ShortenComment(t *testing.T) {
	testString := "Lorem ipsum dolor sit amet, consectetur (adipisici elit), sed eiusmod tempor incidunt."
	require.Equal(t, testString, ShortenComment(testString))

	testLongString := "Lorem ipsum dolor sit amet, consectetur (adipisici elit), sed eiusmod tempor incidunt ut labore (et dolore magna aliqua)."
	testLongStringShortened := "Lorem ipsum dolor sit amet, consectetur (adipisici elit), sed eiusmod tempor incidunt ut labore."
	require.Equal(t, testLongStringShortened, ShortenComment(testLongString))

	testVeryLongString := "Lorem ipsum dolor sit amet, consectetur (adipisici elit), sed eiusmod (tempor incidunt [ut labore et dolore magna aliqua])."
	testVeryLongStringShortened := "Lorem ipsum dolor sit amet, consectetur (adipisici elit), sed eiusmod (tempor incidunt)."
	require.Equal(t, testVeryLongStringShortened, ShortenComment(testVeryLongString))
}

func TestTools_GetLabel(t *testing.T) {
	require.Equal(t, "This is a Label", GetLabel("This is a &Label"))
	require.Equal(t, "Bits & Pieces", GetLabel("Bits && Pieces"))
}

func TestTools_GetOOoLabel(t *testing.T) {
	require.Equal(t, "Bits & Pieces", GetOOoLabel("Bits && Pieces"))
}

func TestTools_GetMnemonic(t *testing.T) {
	require.Equal(t, 'F', GetMnemonic("&File"))
	require.Equal(t, 'O', GetMnemonic("&OK"))
	require.Equal(t, rune(0), GetMnemonic("File && String operations"))
	require.Equal(t, 'O', GetMnemonic("File && String &Operations"))
}
