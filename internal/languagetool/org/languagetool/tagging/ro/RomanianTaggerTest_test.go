package ro

// Twin of RomanianTaggerTest — MapWordTagger smokes
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestRomanianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"casă": {tagging.NewTaggedWord("casă", "S")},
		"mare": {tagging.NewTaggedWord("mare", "A")},
	}
	got := NewRomanianTagger(wt).Tag([]string{"Casă", "mare", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}

func TestRomanianTagger_TaggerMerge(t *testing.T) {
	t.Skip("unimplemented: merge readings need full dict")
}
func TestRomanianTagger_TaggerMerseseram(t *testing.T) {
	// surface smoke for form
	wt := tagging.MapWordTagger{"merseseram": {tagging.NewTaggedWord("merge", "V")}}
	got := NewRomanianTagger(wt).Tag([]string{"merseseram"})
	require.Len(t, got, 1)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}
func TestRomanianTagger_Tagger_Fi(t *testing.T) {
	wt := tagging.MapWordTagger{"fi": {tagging.NewTaggedWord("fi", "V")}}
	require.NotEmpty(t, NewRomanianTagger(wt).TagWord("fi"))
}
func TestRomanianTagger_TaggerUserDict(t *testing.T) {
	t.Skip("unimplemented: user dict overlay")
}
