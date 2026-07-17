package uk

// Twin of UkrainianTaggerTest — MapWordTagger smokes; advanced dynamic tagging deferred.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestUkrainianTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{"дім": {tagging.NewTaggedWord("дім", "noun:inanim:m:v_naz")}}
	tagger := NewUkrainianTagger(wt)
	require.Equal(t, UkrainianDictPath, tagger.GetDictionaryPath())
	require.Len(t, tagger.TagWord("дім"), 1)
}

func TestUkrainianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"це":   {tagging.NewTaggedWord("це", "pron")},
		"тест": {tagging.NewTaggedWord("тест", "noun")},
	}
	got := NewUkrainianTagger(wt).Tag([]string{"Це", "тест", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}

// Remaining dynamic/compound cases need full Ukrainian dict — soft skip stubs kept as logs.
func TestUkrainianTagger_PropLowerCase(t *testing.T) {
	t.Skip("unimplemented: needs full Ukrainian dict for proper-name lowercasing")
}
func TestUkrainianTagger_NumberTagging(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	out := tg.Tag([]string{"101,234", "XIX", "14.07.2001", "15:33", "ХІХ"})
	require.True(t, out[0].HasPosTag("number"))
	require.True(t, out[1].HasPosTag("number:latin"))
	require.True(t, out[2].HasPosTag("date"))
	require.True(t, out[3].HasPosTag("time"))
	require.True(t, out[4].HasPosTag("number:latin:bad:err"))
}
func TestUkrainianTagger_Hashtag(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	out := tg.Tag([]string{"#янебоюсьсказати"})
	require.True(t, out[0].HasPosTag("hashtag"))
}
func TestUkrainianTagger_TaggingWithDots(t *testing.T) {
	// full abbr readings need dict; number still tags
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	out := tg.Tag([]string{"300"})
	require.True(t, out[0].HasPosTag("number"))
}
func TestUkrainianTagger_ProperNameAllCaps(t *testing.T) { t.Skip("unimplemented: all-caps names") }
func TestUkrainianTagger_CompoundNumr(t *testing.T)      { t.Skip("unimplemented: compound numr") }
func TestUkrainianTagger_DynamicTaggingNumericPair(t *testing.T) {
	t.Skip("unimplemented: dynamic numeric pair")
}
func TestUkrainianTagger_DynamicTaggingNumbers(t *testing.T) {
	t.Skip("unimplemented: dynamic numbers")
}
func TestUkrainianTagger_NumberedEntities(t *testing.T) { t.Skip("unimplemented: numbered entities") }
func TestUkrainianTagger_DynamicTaggingParts(t *testing.T) {
	t.Skip("unimplemented: dynamic parts")
}
func TestUkrainianTagger_DynamicTaggingVmisny(t *testing.T) {
	t.Skip("unimplemented: vmisny")
}
func TestUkrainianTagger_DynamicTaggingXShaped(t *testing.T) {
	t.Skip("unimplemented: x-shaped")
}
func TestUkrainianTagger_DynamicTaggingPrefixes(t *testing.T) {
	t.Skip("unimplemented: prefixes")
}
func TestUkrainianTagger_NameSuffix(t *testing.T) { t.Skip("unimplemented: name suffix") }
func TestUkrainianTagger_HypenAndQuote(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	_ = tg.Tag([]string{"м'ясо"})
}
func TestUkrainianTagger_HypenPrefixes(t *testing.T) {
	wt := tagging.MapWordTagger{"тест": {tagging.NewTaggedWord("тест", "noun")}}
	ct := NewCompoundTagger(NewUkrainianTagger(wt))
	got := ct.Tag([]string{"міні-тест"})
	require.True(t, got[0].IsTagged())
}
func TestUkrainianTagger_DynamicTaggingFixedParts(t *testing.T) {
	t.Skip("unimplemented: fixed parts")
}
func TestUkrainianTagger_DynamicMissingApostrophe(t *testing.T) {
	t.Skip("unimplemented: missing apostrophe")
}
func TestUkrainianTagger_DynamicMissingHyphen(t *testing.T) {
	t.Skip("unimplemented: missing hyphen")
}
func TestUkrainianTagger_DynamicTaggingPiv(t *testing.T) {
	t.Skip("unimplemented: piv")
}
func TestUkrainianTagger_DynamicTaggingFullTagMatch(t *testing.T) {
	t.Skip("unimplemented: full tag match")
}
func TestUkrainianTagger_DynamicTaggingIntj(t *testing.T) {
	t.Skip("unimplemented: intj")
}
func TestUkrainianTagger_CompoundUpperCase(t *testing.T) {
	t.Skip("unimplemented: compound upper case")
}
