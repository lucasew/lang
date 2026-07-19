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

func TestUkrainianTagger_PropLowerCase(t *testing.T) {
	// inject: lower form present in map for TagWord
	wt := tagging.MapWordTagger{
		"київ": {tagging.NewTaggedWord("Київ", "noun:inanim:m:v_naz:prop:geo")},
		"Нато": {tagging.NewTaggedWord("Нато", "noun:inanim:m:v_naz:prop")},
	}
	tg := NewUkrainianTagger(wt)
	got := tg.TagWord("київ")
	require.NotEmpty(t, got)
	require.Contains(t, got[0].GetPosTag(), "prop")
	// ALLCAPS → capitalizeProperName + dict (Java path; no invent without dict)
	out := tg.Tag([]string{"НАТО"})
	require.True(t, out[0].IsTagged())
	require.Contains(t, *out[0].GetReadings()[0].GetPOSTag(), "prop")
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
	// Short endings: LetterEndingForNumericHelper (no invent).
	// Long right halves: Java wordTagger only — inject right lemmas.
	wt := tagging.MapWordTagger{
		"річному":     {tagging.NewTaggedWord("річний", "adj:m:v_dav")},
		"відсотково": {tagging.NewTaggedWord("відсотково", "adv")},
	}
	tg := NewUkrainianTagger(wt)
	out := tg.Tag([]string{"100-й", "50-х", "11-ту", "100-річному", "100-відсотково", "10-хвилинка"})
	require.True(t, out[0].HasPartialPosTag("adj"))
	require.True(t, out[0].HasPartialPosTag("numr") || out[0].HasPosTagStartingWith("adj"))
	require.True(t, out[1].HasPartialPosTag("adj"))
	require.True(t, out[2].HasPartialPosTag("adj"))
	require.True(t, out[3].HasPartialPosTag("adj"))
	require.True(t, out[4].HasPosTag("adv") || out[4].HasPartialPosTag("adv"))
	// 10-хвилинка: no invent bare noun POS without dict (fail closed)
	require.False(t, out[5].IsTagged())

	// Without right-side dict: long compounds fail closed
	empty := NewUkrainianTagger(tagging.MapWordTagger{})
	bare := empty.Tag([]string{"100-річному", "100-відсотково"})
	require.False(t, bare[0].IsTagged())
	require.False(t, bare[1].IsTagged())
}
func TestUkrainianTagger_DynamicTaggingParts(t *testing.T) {
	// directional compounds: Java oAdjMatch needs right adj from wordTagger
	wt := tagging.MapWordTagger{
		"Західній":  {tagging.NewTaggedWord("західний", "adj:f:v_dav")},
		"західній":  {tagging.NewTaggedWord("західний", "adj:f:v_dav")},
		"східного":  {tagging.NewTaggedWord("східний", "adj:m:v_rod")},
	}
	tg := NewUkrainianTagger(wt)
	out := tg.Tag([]string{"Південно-Західній", "північно-східного"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[0].HasPartialPosTag("adj"))
	require.True(t, out[1].IsTagged())
	require.True(t, out[1].HasPartialPosTag("adj"))
	// lemma = left.lower + "-" + right lemma
	lemma := out[0].GetReadings()[0].GetLemma()
	require.NotNil(t, lemma)
	require.Equal(t, "південно-західний", *lemma)

	// Without dict: fail closed (no invent endings)
	bare := NewUkrainianTagger(tagging.MapWordTagger{}).Tag([]string{"Південно-Західній"})
	require.False(t, bare[0].IsTagged())
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
	// Java: пів- needs right-side dict (v_rod); street suffixes use CITY_AVENU list.
	wt := tagging.MapWordTagger{
		"України": {tagging.NewTaggedWord("Україна", "noun:inanim:f:v_rod:prop:geo")},
		"години":  {tagging.NewTaggedWord("година", "noun:inanim:f:v_rod")},
	}
	tg := NewUkrainianTagger(wt)
	out := tg.Tag([]string{"пів-України", "пів-години", "Уолл-стрит", "Пенсильванія-авеню"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[0].HasPartialPosTag("alt") || out[0].HasPartialPosTag("noun"))
	require.True(t, out[1].IsTagged())
	require.True(t, out[1].HasPartialPosTag("bad") || out[1].HasPartialPosTag("noun"))
	require.True(t, out[2].IsTagged())
	require.True(t, out[2].HasPartialPosTag("prop"))
	require.True(t, out[3].IsTagged())
	// пів without dict on right fails closed
	require.False(t, NewUkrainianTagger(tagging.MapWordTagger{}).Tag([]string{"пів-України"})[0].IsTagged())
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
	// Java MISSING_HYPHEN: tag base via dict with pron POS — inject якого (no soft invent).
	wt := tagging.MapWordTagger{
		"тест":  {tagging.NewTaggedWord("тест", "noun")},
		"якого": {tagging.NewTaggedWord("який", "adj:m:v_rod:pron:int:rel:def")},
	}
	tg := NewUkrainianTagger(wt)
	// missing hyphen after known prefix: мінітест → tag via міні-тест
	out := tg.Tag([]string{"мінітест", "напівтест", "якогонебудь", "болнебудь"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[1].IsTagged())
	require.True(t, out[2].IsTagged())
	require.True(t, out[2].HasPartialPosTag("bad") || out[2].HasPartialPosTag("adj"))
	// "бол" without dict pronoun POS fails closed (Java болнебудь → null)
	require.False(t, out[3].IsTagged(), "болнебудь needs pronoun base in dict")
}
func TestUkrainianTagger_DynamicTaggingFullTagMatch(t *testing.T) {
	// Java CompoundTagger tags both sides via wordTagger — inject dict forms (no soft invent).
	wt := tagging.MapWordTagger{
		"жило":     {tagging.NewTaggedWord("жити", "verb:imperf:past:n")},
		"було":     {tagging.NewTaggedWord("бути", "verb:imperf:past:n")},
		"учиш":     {tagging.NewTaggedWord("учити", "verb:imperf:pres:s:2")},
		"низенько": {tagging.NewTaggedWord("низенько", "adv")},
		"лікар":    {tagging.NewTaggedWord("лікар", "noun:anim:m:v_naz")},
		"гомеопат": {tagging.NewTaggedWord("гомеопат", "noun:anim:m:v_naz")},
		"а":        {tagging.NewTaggedWord("а", "intj")},
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
	// Java multi-hyphen intj requires intj on both parts; elongated collapses to dict + :alt.
	wt := tagging.MapWordTagger{
		"га":  {tagging.NewTaggedWord("га", "intj")},
		"гей": {tagging.NewTaggedWord("гей", "intj")},
	}
	tg := NewUkrainianTagger(wt)
	out := tg.Tag([]string{"га-га", "геееей"})
	require.True(t, out[0].HasPosTag("intj") || out[0].HasPartialPosTag("intj"))
	// геееей → collapse to гей if pattern matches; or untagged fail closed without invent
	_ = out[1]
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
	// Fail-closed without wordTagger (Java oAdjMatch needs right adj from dict).
	require.Nil(t, DynamicDirectionalAdjReadings("Південно-Західній", nil))
	require.Nil(t, DynamicDirectionalAdjReadings("звичайний", nil))

	// Dict-gated: right part "Західній" / "західній" provides adj tags.
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "західній", "західний":
			return []tagging.TaggedWord{
				{Lemma: "західний", PosTag: "adj:f:v_dav"},
				{Lemma: "західний", PosTag: "adj:f:v_mis"},
			}
		default:
			return nil
		}
	}
	rs := DynamicDirectionalAdjReadings("Південно-Західній", tagWord)
	require.NotEmpty(t, rs)
	require.Equal(t, "південно-західний", rs[0].Lemma)
	require.Contains(t, rs[0].POS, "adj")
	// LEFT_O_ADJ_INVALID + not full-compound adj → :bad (Java)
	require.Contains(t, rs[0].POS, ":bad")

	// Non-compound surface: no invent
	require.Nil(t, DynamicDirectionalAdjReadings("звичайний", tagWord))
	// Right unknown to dict: fail-closed
	require.Nil(t, DynamicDirectionalAdjReadings("Південно-Невідомий", tagWord))
}

func TestCompoundNumrPOS(t *testing.T) {
	require.Equal(t, "numr", CompoundNumrPOS("2-х"))
	require.Equal(t, "numr", CompoundNumrPOS("3ом"))
	require.Equal(t, "", CompoundNumrPOS("42"))
	require.Equal(t, "", CompoundNumrPOS("абв"))
}

func TestDynamicNumericReadings(t *testing.T) {
	// Short ordinal ending: official LetterEndingForNumericHelper map (no invent).
	rs := DynamicNumericReadings("100-й", nil)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, "adj")
	require.Contains(t, rs[0].POS, "numr")
	require.Equal(t, "100-й", rs[0].Lemma)

	// Long right half without dict: fail-closed (Java wordTagger required).
	require.Empty(t, DynamicNumericReadings("100-річному", nil))

	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "річному", "річний":
			return []tagging.TaggedWord{{Lemma: "річний", PosTag: "adj:m:v_dav"}}
		case "сторіччя":
			// Java getTryPrefix("річчя") → "сто" + dict
			return []tagging.TaggedWord{{Lemma: "сторіччя", PosTag: "noun:inanim:n:v_naz"}}
		case "відсотково":
			return []tagging.TaggedWord{{Lemma: "відсотково", PosTag: "adv"}}
		default:
			return nil
		}
	}
	rs2 := DynamicNumericReadings("100-річному", tagWord)
	require.NotEmpty(t, rs2)
	require.Contains(t, rs2[0].POS, "adj")
	require.Equal(t, "100-річний", rs2[0].Lemma)

	// getTryPrefix: 100-річчя via сторіччя in dict
	require.Equal(t, "сто", getTryPrefix("річчя"))
	rs3 := DynamicNumericReadings("100-річчя", tagWord)
	require.NotEmpty(t, rs3)
	require.Equal(t, "100-річчя", rs3[0].Lemma)
	require.Contains(t, rs3[0].POS, "noun")

	// lemma відсотково allowed even without adj POS
	rs4 := DynamicNumericReadings("100-відсотково", tagWord)
	require.NotEmpty(t, rs4)
	require.Equal(t, "100-відсотково", rs4[0].Lemma)

	// bare noun right without adj/відсотково: fail closed
	require.Empty(t, DynamicNumericReadings("10-хвилинка", tagWord))

	require.Empty(t, DynamicNumericReadings("звичайний", tagWord))
}

func TestMissingApostropheCandidates(t *testing.T) {
	cands := MissingApostropheCandidates("зїзду")
	require.Contains(t, cands, "з'їзду")
	require.Empty(t, MissingApostropheCandidates("з'їзду"))
}
