package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/CompoundFilterTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of CompoundFilterTest.testFilter
func TestCompoundFilter_Filter_Twin(t *testing.T) {
	f := NewCompoundFilter()
	require.Equal(t, "tv-meubel", f.Suggest([]string{"tv", "meubel"}))
	require.Equal(t, "rijinstructeur", f.Suggest([]string{"rij", "instructeur"}))
	require.Equal(t, "ANWB-tv-wagen", f.Suggest([]string{"ANWB", "tv", "wagen"}))
}
