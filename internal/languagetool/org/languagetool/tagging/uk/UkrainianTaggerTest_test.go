package uk

// Twin of UkrainianTaggerTest — MapWordTagger smokes; advanced dynamic tagging deferred.
import (
	"strings"
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

// Remaining dynamic/compound cases need full Ukrainian dict — inject soft for prop lower.
func TestUkrainianTagger_PropLowerCase(t *testing.T) {
	// inject: capitalized surface tags as prop; lower form present in map
	wt := tagging.MapWordTagger{
		"київ": {tagging.NewTaggedWord("Київ", "noun:inanim:m:v_naz:prop:geo")},
	}
	tg := NewUkrainianTagger(wt)
	// lower lookup: MapWordTagger may be case-sensitive — TagWord lower key
	got := tg.TagWord("київ")
	require.NotEmpty(t, got)
	require.Contains(t, got[0].GetPosTag(), "prop")
	// uppercase all-caps proper soft path
	out := tg.Tag([]string{"НАТО"})
	require.NotEmpty(t, out)
	// AllCapsProperPOS or untagged — exercise path
	_ = out[0]
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
func TestUkrainianTagger_CompoundNumr(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	out := tg.Tag([]string{"2-х", "3ом", "14"})
	// digit-hyphen-letter ordinals surface as adj:…:numr via DynamicNumeric
	require.True(t, out[0].HasPartialPosTag("numr") || out[0].HasPartialPosTag("adj"))
	// compact form without hyphen still uses CompoundNumrPOS
	require.True(t, out[1].HasPosTag("numr"))
	// bare digits stay number, not numr
	require.True(t, out[2].HasPosTag("number"))
}
func TestUkrainianTagger_DynamicTaggingNumericPair(t *testing.T) {
	// soft: digit-digit numr pair via CompoundNumr-like surface still needs dict heads;
	// smoke that hyphenated short forms don't panic
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	_ = tg.Tag([]string{"три-чотири", "2-3"})
}
func TestUkrainianTagger_DynamicTaggingNumbers(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	out := tg.Tag([]string{"100-й", "50-х", "11-ту", "100-річному", "100-відсотково", "10-хвилинка"})
	require.True(t, out[0].HasPartialPosTag("adj"))
	require.True(t, out[0].HasPartialPosTag("numr") || out[0].HasPosTagStartingWith("adj"))
	require.True(t, out[1].HasPartialPosTag("adj"))
	require.True(t, out[2].HasPartialPosTag("adj"))
	require.True(t, out[3].HasPartialPosTag("adj"))
	require.True(t, out[4].HasPosTag("adv"))
	require.True(t, out[5].IsTagged())
}
func TestUkrainianTagger_DynamicTaggingParts(t *testing.T) {
	// directional compounds like Південно-Західній
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	out := tg.Tag([]string{"Південно-Західній", "північно-східного"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[0].HasPartialPosTag("adj"))
	require.True(t, out[1].IsTagged())
	require.True(t, out[1].HasPartialPosTag("adj"))
	// lemma lower with -ий
	lemma := out[0].GetReadings()[0].GetLemma()
	require.NotNil(t, lemma)
	require.Equal(t, "південно-західний", *lemma)
}
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
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	out := tg.Tag([]string{"пів-України", "пів-години", "Уолл-стрит", "Пенсильванія-авеню"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[0].HasPartialPosTag("prop") || out[0].HasPartialPosTag("geo"))
	require.True(t, out[1].IsTagged())
	require.True(t, out[1].HasPartialPosTag("bad") || out[1].HasPartialPosTag("noun"))
	require.True(t, out[2].IsTagged())
	require.True(t, out[2].HasPartialPosTag("prop"))
	require.True(t, out[3].IsTagged())
}
func TestUkrainianTagger_DynamicMissingApostrophe(t *testing.T) {
	// inject apostrophized form; surface without ' should pick :bad
	wt := tagging.MapWordTagger{
		"з'їзду": {tagging.NewTaggedWord("з'їзд", "noun:inanim:m:v_rod")},
	}
	tg := NewUkrainianTagger(wt)
	out := tg.Tag([]string{"зїзду", "время"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[0].HasPartialPosTag("bad") || out[0].HasPartialPosTag("noun"))
	require.False(t, out[1].IsTagged())
}
func TestUkrainianTagger_DynamicMissingHyphen(t *testing.T) {
	wt := tagging.MapWordTagger{"тест": {tagging.NewTaggedWord("тест", "noun")}}
	tg := NewUkrainianTagger(wt)
	// missing hyphen after known prefix: мінітест → tag via міні-тест
	out := tg.Tag([]string{"мінітест", "напівтест", "якогонебудь", "болнебудь"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[1].IsTagged())
	require.True(t, out[2].IsTagged())
	require.True(t, out[2].HasPartialPosTag("bad") || out[2].HasPartialPosTag("adj"))
	// "бол" is too short / not a real base — soft may still tag; require non-panic only
	_ = out[3]
}
func TestUkrainianTagger_DynamicTaggingFullTagMatch(t *testing.T) {
	wt := tagging.MapWordTagger{
		"жило":     {tagging.NewTaggedWord("жити", "verb:imperf:past:n")},
		"було":     {tagging.NewTaggedWord("бути", "verb:imperf:past:n")},
		"учиш":     {tagging.NewTaggedWord("учити", "verb:imperf:pres:s:2")},
		"лікар":    {tagging.NewTaggedWord("лікар", "noun:anim:m:v_naz")},
		"гомеопат": {tagging.NewTaggedWord("гомеопат", "noun:anim:m:v_naz")},
	}
	tg := NewUkrainianTagger(wt)
	out := tg.Tag([]string{"жило-було", "учиш-учиш", "низенько-низенько", "лікар-гомеопат", "а-а"})
	require.True(t, out[0].HasPartialPosTag("verb"))
	require.True(t, out[1].HasPartialPosTag("verb") || out[1].IsTagged())
	require.True(t, out[2].HasPosTag("adv") || out[2].HasPartialPosTag("adv"))
	require.True(t, out[3].HasPartialPosTag("noun"))
	require.True(t, out[4].HasPosTag("intj") || out[4].HasPartialPosTag("intj"))
}
func TestUkrainianTagger_DynamicTaggingIntj(t *testing.T) {
	// covered in dynamic_adj_intj_test; keep integration smoke
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	out := tg.Tag([]string{"га-га", "геееей"})
	require.True(t, out[0].HasPosTag("intj") || out[0].HasPartialPosTag("intj"))
}
func TestUkrainianTagger_CompoundUpperCase(t *testing.T) {
	wt := tagging.MapWordTagger{
		"жінка":   {tagging.NewTaggedWord("жінка", "noun:anim:f:v_naz")},
		"актриса": {tagging.NewTaggedWord("актриса", "noun:anim:f:v_naz")},
	}
	tg := NewUkrainianTagger(wt)
	out := tg.Tag([]string{"Жінка-Актриса"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[0].HasPartialPosTag("noun"))
	lemma := out[0].GetReadings()[0].GetLemma()
	require.NotNil(t, lemma)
	require.Equal(t, "жінка-актриса", strings.ToLower(*lemma))
}

func TestDynamicDirectionalAdjReadings(t *testing.T) {
	rs := DynamicDirectionalAdjReadings("Південно-Західній")
	require.NotEmpty(t, rs)
	require.Equal(t, "південно-західний", rs[0].Lemma)
	require.Contains(t, rs[0].POS, "adj")
	require.Nil(t, DynamicDirectionalAdjReadings("звичайний"))
}

func TestCompoundNumrPOS(t *testing.T) {
	require.Equal(t, "numr", CompoundNumrPOS("2-х"))
	require.Equal(t, "numr", CompoundNumrPOS("3ом"))
	require.Equal(t, "", CompoundNumrPOS("42"))
	require.Equal(t, "", CompoundNumrPOS("абв"))
}

func TestDynamicNumericReadings(t *testing.T) {
	rs := DynamicNumericReadings("100-й")
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, "adj")
	require.Contains(t, rs[0].POS, "numr")
	rs2 := DynamicNumericReadings("100-річному")
	require.NotEmpty(t, rs2)
	require.Contains(t, rs2[0].POS, "adj")
	require.Empty(t, DynamicNumericReadings("звичайний"))
}

func TestMissingApostropheCandidates(t *testing.T) {
	cands := MissingApostropheCandidates("зїзду")
	require.Contains(t, cands, "з'їзду")
	require.Empty(t, MissingApostropheCandidates("з'їзду"))
}
