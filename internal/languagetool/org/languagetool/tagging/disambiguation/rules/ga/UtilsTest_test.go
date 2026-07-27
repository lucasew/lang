package ga

// Twin of org.languagetool.tagging.disambiguation.rules.ga.UtilsTest.
// Implementations live in tagging/ga; this package mirrors the Java test path.
import (
	"testing"

	tagga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/stretchr/testify/require"
)

func TestUtils_ToLowerCaseIrish(t *testing.T) {
	require.Equal(t, "test", tagga.ToLowerCaseIrish("TEST"))
	require.Equal(t, "test", tagga.ToLowerCaseIrish("Test"))
	require.Equal(t, "t-aon", tagga.ToLowerCaseIrish("tAON"))
	require.Equal(t, "n-aon", tagga.ToLowerCaseIrish("nAON"))
}

func TestUtils_UnLenited(t *testing.T) {
	require.Equal(t, "Kate", tagga.UnLenite("Khate"))
	require.Equal(t, "can", tagga.UnLenite("chan"))
	require.Equal(t, "ba", tagga.UnLenite("bha"))
	require.Equal(t, "b", tagga.UnLenite("bh"))
	require.Equal(t, "", tagga.UnLenite("can"))
	require.Equal(t, "", tagga.UnLenite("a"))
}

func TestUtils_UnEclipseChar(t *testing.T) {
	require.Equal(t, "carr", tagga.UnEclipseChar("gcarr", 'g', 'c'))
	require.Equal(t, "Carr", tagga.UnEclipseChar("gCarr", 'g', 'c'))
	require.Equal(t, "Carr", tagga.UnEclipseChar("G-carr", 'g', 'c'))
	require.Equal(t, "Carr", tagga.UnEclipseChar("Gcarr", 'g', 'c'))
	require.Equal(t, "CARR", tagga.UnEclipseChar("GCARR", 'g', 'c'))
	require.Equal(t, "", tagga.UnEclipseChar("carr", 'g', 'c'))
}

func TestUtils_UnEclipse(t *testing.T) {
	require.Equal(t, "carr", tagga.UnEclipse("g-carr"))
	require.Equal(t, "doras", tagga.UnEclipse("n-doras"))
	require.Equal(t, "Geata", tagga.UnEclipse("N-geata"))
	require.Equal(t, "peann", tagga.UnEclipse("bpeann"))
	require.Equal(t, "bean", tagga.UnEclipse("mbean"))
	require.Equal(t, "Éin", tagga.UnEclipse("N-éin"))
	require.Equal(t, "focal", tagga.UnEclipse("bhfocal"))
	require.Equal(t, "Focal", tagga.UnEclipse("Bhfocal"))
	require.Equal(t, "Focal", tagga.UnEclipse("Bfocal"))
	require.Equal(t, "", tagga.UnEclipse("carr"))
}

func TestUtils_UnLeniteDefiniteS(t *testing.T) {
	require.Equal(t, "seomra1", tagga.UnLeniteDefiniteS("t-seomra1"))
	require.Equal(t, "seomra2", tagga.UnLeniteDefiniteS("tseomra2"))
	require.Equal(t, "Seomra3", tagga.UnLeniteDefiniteS("tSeomra3"))
	require.Equal(t, "Seomra4", tagga.UnLeniteDefiniteS("TSeomra4"))
	require.Equal(t, "Seomra5", tagga.UnLeniteDefiniteS("Tseomra5"))
	require.Equal(t, "Seomra6", tagga.UnLeniteDefiniteS("t-Seomra6"))
	require.Equal(t, "Seomra7", tagga.UnLeniteDefiniteS("T-Seomra7"))
	require.Equal(t, "Seomra8", tagga.UnLeniteDefiniteS("T-seomra8"))
	require.Equal(t, "", tagga.UnLeniteDefiniteS("seomra9"))
}

func TestUtils_Demutate(t *testing.T) {
	tmp := tagga.Demutate("gcharr")
	require.Equal(t, "carr", tmp.GetWord())
	require.Equal(t, ":EclLen", tmp.GetAppendTag())
	tmp = tagga.Demutate("t-seomra")
	require.Equal(t, "seomra", tmp.GetWord())
	require.Equal(t, "(?:C[UMC]:)?Noun:.*:DefArt", tmp.GetRestrictToPos())
}

func TestUtils_FixSuffix(t *testing.T) {
	tmp := tagga.FixSuffix("caimiléaracht")
	require.Equal(t, "caimiléireacht", tmp.GetWord())
}

func TestUtils_UnPonc(t *testing.T) {
	require.Equal(t, "chuir", tagga.UnPonc("ċuir"))
	require.Equal(t, "CHUIR", tagga.UnPonc("ĊUIR"))
	require.Equal(t, "Chuir", tagga.UnPonc("Ċuir"))
	require.Equal(t, "FÉACH", tagga.UnPonc("FÉAĊ"))
}

func TestUtils_Simplify(t *testing.T) {
	boldUpper := "\U0001D412\U0001D404\U0001D400\U0001D40D\U0001D400\U0001D413\U0001D407\U0001D400\U0001D408\U0001D411"
	require.Equal(t, "SEANATHAIR", tagga.SimplifyMathematical(boldUpper))
	boldLower := "\U0001D42C\U0001D41E\U0001D41A\U0001D427\U0001D41A\U0001D42D\U0001D421\U0001D41A\U0001D422\U0001D42B"
	require.Equal(t, "seanathair", tagga.SimplifyMathematical(boldLower))
	require.Equal(t, "999", tagga.SimplifyMathematicalFlags("\U0001D7FF\U0001D7FF\U0001D7FF", false, true))
}

func TestUtils_GreekToLatin(t *testing.T) {
	require.Equal(t, "BOTTOM", tagga.GreekToLatin("ΒΟΤΤΟΜ"))
}

func TestUtils_HasMixedGreekAndLatin(t *testing.T) {
	require.Equal(t, true, tagga.HasMixedGreekAndLatin("Nοt"))
}

func TestUtils_IsAllMathsChars(t *testing.T) {
	boldUpper := "\U0001D412\U0001D404\U0001D400\U0001D40D\U0001D400\U0001D413\U0001D407\U0001D400\U0001D408\U0001D411"
	require.Equal(t, false, tagga.IsAllMathsChars("foo"))
	require.Equal(t, false, tagga.IsAllMathsChars("f\U0001D412"))
	require.Equal(t, true, tagga.IsAllMathsChars(boldUpper))
}

func TestUtils_IsAllHalfWidthChars(t *testing.T) {
	torrach := "ｔｏｒｒａｃｈ"
	require.Equal(t, true, tagga.IsAllHalfWidthChars(torrach))
	require.Equal(t, false, tagga.IsAllHalfWidthChars(torrach+"a"))
}

func TestUtils_HalfwidthLatinToLatin(t *testing.T) {
	require.Equal(t, "torrach", tagga.HalfwidthLatinToLatin("ｔｏｒｒａｃｈ"))
}
