package ga

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Full twin of org.languagetool.tagging.disambiguation.rules.ga.UtilsTest
// plus extra coverage for public Utils methods used by Irish rules.

// Twin: UtilsTest.testToLowerCaseIrish
func TestToLowerCaseIrish(t *testing.T) {
	require.Equal(t, "test", ToLowerCaseIrish("TEST"))
	require.Equal(t, "test", ToLowerCaseIrish("Test"))
	require.Equal(t, "t-aon", ToLowerCaseIrish("tAON"))
	require.Equal(t, "n-aon", ToLowerCaseIrish("nAON"))
}

// Twin: UtilsTest.testUnLenited
func TestUnLenited(t *testing.T) {
	require.Equal(t, "Kate", UnLenite("Khate"))
	require.Equal(t, "can", UnLenite("chan"))
	require.Equal(t, "ba", UnLenite("bha"))
	require.Equal(t, "b", UnLenite("bh"))
	require.Equal(t, "", UnLenite("can")) // Java null
	require.Equal(t, "", UnLenite("a"))
}

// Twin: UtilsTest.testUnEclipseChar
func TestUnEclipseChar(t *testing.T) {
	// properly eclipsed
	require.Equal(t, "carr", UnEclipseChar("gcarr", 'g', 'c'))
	require.Equal(t, "Carr", UnEclipseChar("gCarr", 'g', 'c'))
	// improperly eclipsed
	require.Equal(t, "Carr", UnEclipseChar("G-carr", 'g', 'c'))
	require.Equal(t, "Carr", UnEclipseChar("Gcarr", 'g', 'c'))
	require.Equal(t, "CARR", UnEclipseChar("GCARR", 'g', 'c'))
	// not eclipsed
	require.Equal(t, "", UnEclipseChar("carr", 'g', 'c'))
}

// Twin: UtilsTest.testUnEclipse
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

// Twin: UtilsTest.testUnLeniteDefiniteS
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

// Twin: UtilsTest.testDemutate
func TestDemutate(t *testing.T) {
	tmp := Demutate("gcharr")
	require.Equal(t, "carr", tmp.GetWord())
	require.Equal(t, ":EclLen", tmp.GetAppendTag())
	tmp = Demutate("t-seomra")
	require.Equal(t, "seomra", tmp.GetWord())
	require.Equal(t, "(?:C[UMC]:)?Noun:.*:DefArt", tmp.GetRestrictToPos())
}

// Twin: UtilsTest.testFixSuffix
func TestFixSuffix(t *testing.T) {
	tmp := FixSuffix("caimiléaracht")
	require.Equal(t, "caimiléireacht", tmp.GetWord())
}

// Twin: UtilsTest.testUnPonc
func TestUnPonc(t *testing.T) {
	require.Equal(t, "chuir", UnPonc("ċuir"))
	require.Equal(t, "CHUIR", UnPonc("ĊUIR"))
	require.Equal(t, "Chuir", UnPonc("Ċuir"))
	require.Equal(t, "FÉACH", UnPonc("FÉAĊ"))
}

// Twin: UtilsTest.testSimplify
func TestSimplifyMathematical(t *testing.T) {
	// perl offsets 18 4 0 13 0 19 7 0 8 17 from bold capital A (U+1D400)
	boldUpper := "\U0001D412\U0001D404\U0001D400\U0001D40D\U0001D400\U0001D413\U0001D407\U0001D400\U0001D408\U0001D411"
	require.Equal(t, "SEANATHAIR", SimplifyMathematical(boldUpper))
	boldLower := "\U0001D42C\U0001D41E\U0001D41A\U0001D427\U0001D41A\U0001D42D\U0001D421\U0001D41A\U0001D422\U0001D42B"
	require.Equal(t, "seanathair", SimplifyMathematical(boldLower))

	// Monospace digit 9 ×3 with normaliseDigits=true (Java \uD835\uDFFF)
	require.Equal(t, "999", SimplifyMathematicalFlags("\U0001D7FF\U0001D7FF\U0001D7FF", false, true))
	// without digit flag, trail stays unmapped → keep original maths digits
	require.Equal(t, "\U0001D7FF\U0001D7FF\U0001D7FF", SimplifyMathematicalFlags("\U0001D7FF\U0001D7FF\U0001D7FF", false, false))
}

// Twin: UtilsTest.testGreekToLatin
func TestGreekToLatin(t *testing.T) {
	require.Equal(t, "BOTTOM", GreekToLatin("ΒΟΤΤΟΜ"))
}

// Twin: UtilsTest.testHasMixedGreekAndLatin
func TestHasMixedGreekAndLatin(t *testing.T) {
	require.Equal(t, true, HasMixedGreekAndLatin("Nοt"))
}

// Twin: UtilsTest.testIsAllMathsChars
func TestIsAllMathsChars(t *testing.T) {
	boldUpper := "\U0001D412\U0001D404\U0001D400\U0001D40D\U0001D400\U0001D413\U0001D407\U0001D400\U0001D408\U0001D411"
	require.Equal(t, false, IsAllMathsChars("foo"))
	require.Equal(t, false, IsAllMathsChars("f\U0001D412"))
	require.Equal(t, true, IsAllMathsChars(boldUpper))
}

// Twin: UtilsTest.testIsAllHalfWidthChars / testHalfwidthLatinToLatin
func TestIsAllHalfWidthChars(t *testing.T) {
	torrach := "ｔｏｒｒａｃｈ"
	require.Equal(t, true, IsAllHalfWidthChars(torrach))
	require.Equal(t, false, IsAllHalfWidthChars(torrach+"a"))
}

func TestHalfwidthLatinToLatin(t *testing.T) {
	torrach := "ｔｏｒｒａｃｈ"
	require.Equal(t, "torrach", HalfwidthLatinToLatin(torrach))
}

// Extra: Java-visible lenite/eclipse outcomes used by DativePluralsData.
func TestLeniteEclipse(t *testing.T) {
	require.Equal(t, "bhean", Lenite("bean"))
	// Java does not guard already-lenited forms — double-h
	require.Equal(t, "bhhean", Lenite("bhean"))
	// s-lenition only when second letter isSLenitable
	require.Equal(t, "shúil", Lenite("súil"))
	require.Equal(t, "spúnóg", Lenite("spúnóg"))
	require.Equal(t, "mbean", Eclipse("bean"))
	require.Equal(t, "nÉireann", Eclipse("Éireann"))
	require.Equal(t, "n-éan", Eclipse("éan"))
	require.True(t, IsVowel('á'))
	require.False(t, IsVowel('b'))
}

func TestStartsWithUppercaseIrish(t *testing.T) {
	require.True(t, StartsWithUppercase("mBéal"))
	require.True(t, StartsWithUppercase("dTúr"))
	require.True(t, StartsWithUppercase("tAigéan"))
	require.True(t, StartsWithUppercase("hAigéan"))
	require.True(t, StartsWithUppercase("nÁise"))
	require.True(t, StartsWithUppercase("tSeomra"))
	require.True(t, StartsWithUppercase("Béal"))
	require.False(t, StartsWithUppercase("mbéal"))
	require.False(t, StartsWithUppercase("béal"))
}

func TestIsAllUppercaseIrish(t *testing.T) {
	require.True(t, IsAllUppercase("BÉAL"))
	require.True(t, IsAllUppercase("MBÉAL"))
	require.False(t, IsAllUppercase("mBéal"))
	require.False(t, IsAllUppercase("Béal"))
}

func TestRetaggableSetAppendTag(t *testing.T) {
	r := NewRetaggable("x", "", ":MorphError")
	r.SetAppendTag(":Len:MorphError")
	require.Equal(t, ":MorphError:Len:MorphError", r.GetAppendTag())
	r2 := NewRetaggable("y", "", "")
	r2.SetAppendTag(":Ecl")
	require.Equal(t, ":Ecl", r2.GetAppendTag())
}

func TestMorphWordDefArt(t *testing.T) {
	// eclipsis+lenition yields mut + DefArt:MorphError extras
	out := MorphWord("gcharr")
	require.GreaterOrEqual(t, len(out), 2)
	require.Equal(t, "carr", out[0].GetWord())
	require.Equal(t, ":EclLen", out[0].GetAppendTag())
	require.Equal(t, ":DefArt:MorphError", out[1].GetAppendTag())
}

func TestIsPonc(t *testing.T) {
	require.True(t, IsPonc('ċ'))
	require.True(t, IsPonc('Ċ'))
	require.False(t, IsPonc('c'))
	require.True(t, ContainsPonc("aċ"))
	require.False(t, ContainsPonc("ach"))
}

// Italic / fraktur / monospace letter styles via getMathsChar chain.
func TestSimplifyMathematicalStyles(t *testing.T) {
	// Mathematical italic capital A U+1D434 → 'A'
	require.Equal(t, "A", SimplifyMathematical("\U0001D434"))
	// Mathematical monospace small z U+1D6A3 → 'z'
	require.Equal(t, "z", SimplifyMathematical("\U0001D6A3"))
	// Bold digit 5 U+1D7D3 only with digits flag
	require.Equal(t, "5", SimplifyMathematicalFlags("\U0001D7D3", false, true))
	// Greek bold capital Alpha U+1D6A8 with greek flag
	require.Equal(t, "Α", SimplifyMathematicalFlags("\U0001D6A8", true, false))
}
