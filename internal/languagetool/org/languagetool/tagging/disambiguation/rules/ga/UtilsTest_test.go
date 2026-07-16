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
}
