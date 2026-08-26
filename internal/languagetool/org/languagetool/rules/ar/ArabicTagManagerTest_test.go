package ar

// Twin of ArabicTagManagerTest — uses tagging/ar.ArabicTagManager
import (
	"testing"

	tagar "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/stretchr/testify/require"
)

// Port of ArabicTagManagerTest.testTagger
func TestArabicTagManager_Tagger(t *testing.T) {
	tagManager := tagar.NewArabicTagManager()
	require.Equal(t, "NJ-;M1I-;-K-", tagManager.SetJar("NJ-;M1I-;---", "K"))
	require.Equal(t, "NJ-;M1I-;---", tagManager.SetJar("NJ-;M1I-;---", "-"))
	require.Equal(t, "NJ-;M1I-;--L", tagManager.SetDefinite("NJ-;M1I-;---", "L"))
	require.Equal(t, "NJ-;M1I-;--H", tagManager.SetDefinite("NJ-;M1I-;--H", "L"))
	require.Equal(t, "NJ-;M1I-;--H", tagManager.SetPronoun("NJ-;M1I-;---", "H"))
	require.Equal(t, "NJ-;M1I-;W--", tagManager.SetConjunction("NJ-;M1I-;---", "W"))
	require.Equal(t, "V-1;M1I----;W--", tagManager.SetConjunction("V-1;M1I----;---", "W"))
	require.Equal(t, "و", tagManager.GetConjunctionPrefix("V-1;M1I----;W--"))
}
