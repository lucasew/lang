package ga

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLeniteEclipse(t *testing.T) {
	require.Equal(t, "bhean", Lenite("bean"))
	require.Equal(t, "bhean", Lenite("bhean"))
	require.True(t, len(Eclipse("bean")) > 0 && Eclipse("bean")[0] == 'm')
	require.True(t, IsVowel('á'))
	require.False(t, IsVowel('b'))
}

func TestFixSuffix(t *testing.T) {
	r := FixSuffix("caimiléaracht")
	require.Equal(t, "caimiléireacht", r.GetWord())
}

func TestToLowerCaseIrish(t *testing.T) {
	require.Equal(t, "test", ToLowerCaseIrish("Test"))
	require.Equal(t, "test", ToLowerCaseIrish("TEST"))
	require.Equal(t, "t-aon", ToLowerCaseIrish("tAON"))
	require.Equal(t, "n-aon", ToLowerCaseIrish("nAON"))
}

func TestUnLenited(t *testing.T) {
	require.Equal(t, "Kate", UnLenite("Khate"))
	require.Equal(t, "can", UnLenite("chan"))
	require.Equal(t, "ba", UnLenite("bha"))
	require.Equal(t, "b", UnLenite("bh"))
	require.Equal(t, "", UnLenite("can"))
	require.Equal(t, "", UnLenite("a"))
}

func TestUnEclipseChar(t *testing.T) {
	require.Equal(t, "carr", UnEclipseChar("gcarr", 'g', 'c'))
	require.Equal(t, "Carr", UnEclipseChar("gCarr", 'g', 'c'))
	require.Equal(t, "Carr", UnEclipseChar("G-carr", 'g', 'c'))
	require.Equal(t, "Carr", UnEclipseChar("Gcarr", 'g', 'c'))
	require.Equal(t, "CARR", UnEclipseChar("GCARR", 'g', 'c'))
	require.Equal(t, "", UnEclipseChar("carr", 'g', 'c'))
}

func TestUnEclipse(t *testing.T) {
	require.Equal(t, "carr", UnEclipse("g-carr"))
	require.Equal(t, "doras", UnEclipse("n-doras"))
	require.Equal(t, "Geata", UnEclipse("N-geata"))
	require.Equal(t, "peann", UnEclipse("bpeann"))
	require.Equal(t, "bean", UnEclipse("mbean"))
	require.Equal(t, "Éin", UnEclipse("N-éin"))
	require.Equal(t, "focal", UnEclipse("bhfocal"))
	require.Equal(t, "Focal", UnEclipse("Bhfocal"))
	require.Equal(t, "Focal", UnEclipse("Bfocal"))
	require.Equal(t, "", UnEclipse("carr"))
}

func TestUnLeniteDefiniteS(t *testing.T) {
	require.Equal(t, "seomra1", UnLeniteDefiniteS("t-seomra1"))
	require.Equal(t, "seomra2", UnLeniteDefiniteS("tseomra2"))
	require.Equal(t, "Seomra3", UnLeniteDefiniteS("tSeomra3"))
	require.Equal(t, "Seomra4", UnLeniteDefiniteS("TSeomra4"))
	require.Equal(t, "Seomra5", UnLeniteDefiniteS("Tseomra5"))
	require.Equal(t, "Seomra6", UnLeniteDefiniteS("t-Seomra6"))
	require.Equal(t, "Seomra7", UnLeniteDefiniteS("T-Seomra7"))
	require.Equal(t, "Seomra8", UnLeniteDefiniteS("T-seomra8"))
	require.Equal(t, "", UnLeniteDefiniteS("seomra9"))
}

func TestDemutate(t *testing.T) {
	tmp := Demutate("gcharr")
	require.Equal(t, "carr", tmp.GetWord())
	require.Equal(t, ":EclLen", tmp.GetAppendTag())
	tmp = Demutate("t-seomra")
	require.Equal(t, "seomra", tmp.GetWord())
	require.Equal(t, "(?:C[UMC]:)?Noun:.*:DefArt", tmp.GetRestrictToPos())
}

func TestUnPonc(t *testing.T) {
	require.Equal(t, "chuir", UnPonc("ċuir"))
	require.Equal(t, "CHUIR", UnPonc("ĊUIR"))
	require.Equal(t, "Chuir", UnPonc("Ċuir"))
	require.Equal(t, "FÉACH", UnPonc("FÉAĊ"))
}

func TestSimplifyMathematical(t *testing.T) {
	// Bold capital S E A N A T H A I R (offsets 18,4,0,13,0,19,7,0,8,17)
	boldUpper := "\U0001D412\U0001D404\U0001D400\U0001D40D\U0001D400\U0001D413\U0001D407\U0001D400\U0001D408\U0001D411"
	require.Equal(t, "SEANATHAIR", SimplifyMathematical(boldUpper))
	boldLower := "\U0001D42C\U0001D41E\U0001D41A\U0001D427\U0001D41A\U0001D42D\U0001D421\U0001D41A\U0001D422\U0001D42B"
	require.Equal(t, "seanathair", SimplifyMathematical(boldLower))
}
