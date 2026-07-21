package ga

// Twin of UtilsTest — implementations live in tagging/ga (same logic).
// This package path is the Java twin location; tests re-validate via local re-exports.
import (
	"testing"

	tagga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/stretchr/testify/require"
)

func TestUtils_ToLowerCaseIrish(t *testing.T) {
	require.Equal(t, "test", tagga.ToLowerCaseIrish("Test"))
	require.Equal(t, "t-aon", tagga.ToLowerCaseIrish("tAON"))
	require.Equal(t, "n-aon", tagga.ToLowerCaseIrish("nAON"))
}

func TestUtils_UnLenited(t *testing.T) {
	require.Equal(t, "Kate", tagga.UnLenite("Khate"))
	require.Equal(t, "", tagga.UnLenite("can"))
}

func TestUtils_UnEclipseChar(t *testing.T) {
	require.Equal(t, "carr", tagga.UnEclipseChar("gcarr", 'g', 'c'))
	require.Equal(t, "", tagga.UnEclipseChar("carr", 'g', 'c'))
}

func TestUtils_UnEclipse(t *testing.T) {
	require.Equal(t, "carr", tagga.UnEclipse("g-carr"))
	require.Equal(t, "focal", tagga.UnEclipse("bhfocal"))
	require.Equal(t, "", tagga.UnEclipse("carr"))
}

func TestUtils_UnLeniteDefiniteS(t *testing.T) {
	require.Equal(t, "seomra1", tagga.UnLeniteDefiniteS("t-seomra1"))
	require.Equal(t, "", tagga.UnLeniteDefiniteS("seomra9"))
}

func TestUtils_Demutate(t *testing.T) {
	tmp := tagga.Demutate("gcharr")
	require.Equal(t, "carr", tmp.GetWord())
	require.Equal(t, ":EclLen", tmp.GetAppendTag())
}

func TestUtils_FixSuffix(t *testing.T) {
	require.Equal(t, "caimiléireacht", tagga.FixSuffix("caimiléaracht").GetWord())
}

func TestUtils_UnPonc(t *testing.T) {
	require.Equal(t, "chuir", tagga.UnPonc("ċuir"))
	require.Equal(t, "CHUIR", tagga.UnPonc("ĊUIR"))
}

func TestUtils_Simplify(t *testing.T) {
	boldUpper := "\U0001D412\U0001D404\U0001D400\U0001D40D\U0001D400\U0001D413\U0001D407\U0001D400\U0001D408\U0001D411"
	require.Equal(t, "SEANATHAIR", tagga.SimplifyMathematical(boldUpper))
	boldLower := "\U0001D42C\U0001D41E\U0001D41A\U0001D427\U0001D41A\U0001D42D\U0001D421\U0001D41A\U0001D422\U0001D42B"
	require.Equal(t, "seanathair", tagga.SimplifyMathematical(boldLower))
}

// Twin: Utils.greekToLatin("ΒΟΤΤΟΜ") → "BOTTOM"
func TestUtils_GreekToLatin(t *testing.T) {
	require.Equal(t, "BOTTOM", tagga.GreekToLatin("ΒΟΤΤΟΜ"))
}

// Twin: Utils.hasMixedGreekAndLatin("Nοt") — Latin N + Greek ο + Latin t
func TestUtils_HasMixedGreekAndLatin(t *testing.T) {
	require.True(t, tagga.HasMixedGreekAndLatin("Nοt"))
	require.False(t, tagga.HasMixedGreekAndLatin("Not"))
	require.False(t, tagga.HasMixedGreekAndLatin("Νοτ")) // all Greek lookalikes
}

func TestUtils_IsAllMathsChars(t *testing.T) {
	boldUpper := "\U0001D412\U0001D404\U0001D400\U0001D40D\U0001D400\U0001D413\U0001D407\U0001D400\U0001D408\U0001D411"
	require.False(t, tagga.IsAllMathsChars("foo"))
	require.False(t, tagga.IsAllMathsChars("f\U0001D412"))
	require.True(t, tagga.IsAllMathsChars(boldUpper))
}

func TestUtils_IsAllHalfWidthChars(t *testing.T) {
	torrach := "ｔｏｒｒａｃｈ"
	require.True(t, tagga.IsAllHalfWidthChars(torrach))
	require.False(t, tagga.IsAllHalfWidthChars(torrach+"a"))
}

func TestUtils_HalfwidthLatinToLatin(t *testing.T) {
	require.Equal(t, "torrach", tagga.HalfwidthLatinToLatin("ｔｏｒｒａｃｈ"))
}
