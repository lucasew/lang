package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/ToolsTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of ToolsTest.testBasicConcatenation
func TestTools_BasicConcatenation(t *testing.T) {
	require.Equal(t, "huisdeur", GlueParts([]string{"huis", "deur"}))
	require.Equal(t, "tv-programma", GlueParts([]string{"tv", "programma"}))
	require.Equal(t, "auto2-deurs", GlueParts([]string{"auto2", "deurs"}))
	require.Equal(t, "zee-eend", GlueParts([]string{"zee", "eend"}))
	require.Equal(t, "mms-eend", GlueParts([]string{"mms", "eend"}))
	require.Equal(t, "EersteKlasservice", GlueParts([]string{"EersteKlas", "service"}))
	require.Equal(t, "3Dprinter", GlueParts([]string{"3D", "printer"}))
	require.Equal(t, "grootmoederhuis", GlueParts([]string{"groot", "moeder", "huis"}))
	require.Equal(t, "sport-tv-uitzending", GlueParts([]string{"sport", "tv", "uitzending"}))
	require.Equal(t, "auto-pilot", GlueParts([]string{"auto-", "pilot"}))
	require.Equal(t, "foto-5dcamera", GlueParts([]string{"foto", "5d", "camera"}))
	require.Equal(t, "xyZ-xyz", GlueParts([]string{"xyZ", "xyz"}))
	require.Equal(t, "xyZ-Xyz", GlueParts([]string{"xyZ", "Xyz"}))
	require.Equal(t, "xyz-Xyz", GlueParts([]string{"xyz", "Xyz"}))
	require.Equal(t, "xxx-z-yyy", GlueParts([]string{"xxx-z", "yyy"}))
	// Tools twin method
	require.Equal(t, "huisdeur", (Tools{}).GlueParts([]string{"huis", "deur"}))
}
