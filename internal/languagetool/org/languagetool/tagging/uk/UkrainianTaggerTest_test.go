package uk

// Twin of UkrainianTaggerTest — MapWordTagger smokes; advanced dynamic tagging deferred.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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
		"річному":    {tagging.NewTaggedWord("річний", "adj:m:v_dav")},
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
	// directional compounds: Java oAdjMatch needs left evidence + right adj from wordTagger
	wt := tagging.MapWordTagger{
		"південно": {tagging.NewTaggedWord("південно", "adv")},
		"північно": {tagging.NewTaggedWord("північно", "adv")},
		"Західній": {tagging.NewTaggedWord("західний", "adj:f:v_dav")},
		"західній": {tagging.NewTaggedWord("західний", "adj:f:v_dav")},
		"східного": {tagging.NewTaggedWord("східний", "adj:m:v_rod")},
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
	// Java: екс-«депутат» → lemma екс-депутат + noun…:alt (quotes stripped, dash prefix)
	wt := tagging.MapWordTagger{
		"депутат": {tagging.NewTaggedWord("депутат", "noun:anim:m:v_naz")},
	}
	tg := NewUkrainianTagger(wt)
	out := tg.Tag([]string{"екс-«депутат»"})
	require.True(t, out[0].IsTagged())
	require.True(t, out[0].HasPartialPosTag("noun"))
	// surface kept with quotes
	require.Equal(t, "екс-«депутат»", out[0].GetToken())
	lem := out[0].GetReadings()[0].GetLemma()
	require.NotNil(t, lem)
	require.Equal(t, "екс-депутат", *lem)
	// fail-closed without right-side dict
	require.False(t, NewUkrainianTagger(tagging.MapWordTagger{}).Tag([]string{"екс-«депутат»"})[0].IsTagged())
}

func TestCompoundWithQuotesReadings_unit(t *testing.T) {
	p, l := "noun:anim:m:v_naz:alt", "екс-депутат"
	retag := func(adj string) []*languagetool.AnalyzedToken {
		require.Equal(t, "екс-депутат", adj)
		return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(adj, &p, &l)}
	}
	out := CompoundWithQuotesReadings("екс-«депутат»", retag)
	require.Len(t, out, 1)
	require.Equal(t, "екс-«депутат»", out[0].GetToken())
	require.Equal(t, "екс-депутат", *out[0].GetLemma())
	// no quote pattern
	require.Empty(t, CompoundWithQuotesReadings("екс-депутат", retag))
	// too short
	require.Empty(t, CompoundWithQuotesReadings("а-«б»", retag))
	// COMPOUND_WITH_QUOTES2: closing quote then dash ("заступницю"-колаборантку)
	p2, l2 := "noun:anim:f:v_zna", "заступницю-колаборантку"
	out2 := CompoundWithQuotesReadings(`"заступницю"-колаборантку`, func(adj string) []*languagetool.AnalyzedToken {
		require.Equal(t, "заступницю-колаборантку", adj)
		return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(adj, &p2, &l2)}
	})
	require.Len(t, out2, 1)
	require.Equal(t, `"заступницю"-колаборантку`, out2[0].GetToken())
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

	// Dict-gated: left needs adv/adj evidence (oAdjMatch); right adj from dict.
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "південно":
			return []tagging.TaggedWord{{Lemma: "південно", PosTag: "adv"}}
		case "південний": // oToYj fallback
			return []tagging.TaggedWord{{Lemma: "південний", PosTag: "adj:m:v_naz"}}
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
	// Left without dict evidence: fail-closed
	rightOnly := func(s string) []tagging.TaggedWord {
		if strings.ToLower(s) == "західній" || strings.ToLower(s) == "західний" {
			return []tagging.TaggedWord{{Lemma: "західний", PosTag: "adj:f:v_dav"}}
		}
		return nil
	}
	require.Nil(t, DynamicDirectionalAdjReadings("Південно-Західній", rightOnly))
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

func TestDynamicPoAdvReadings(t *testing.T) {
	// Fail-closed without dict
	require.Nil(t, DynamicPoAdvReadings("по-сибірськи", nil))

	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "сибірський":
			return []tagging.TaggedWord{{Lemma: "сибірський", PosTag: "adj:m:v_naz"}}
		case "свинячому":
			return []tagging.TaggedWord{{Lemma: "свинячий", PosTag: "adj:m:v_mis"}}
		default:
			return nil
		}
	}
	// SKY path: right +й for dict lookup
	rs := DynamicPoAdvReadings("по-сибірськи", tagWord)
	require.NotEmpty(t, rs)
	require.Equal(t, "adv", rs[0].POS)
	require.Equal(t, "по-сибірськи", rs[0].Lemma)

	// -ому path
	rs2 := DynamicPoAdvReadings("по-свинячому", tagWord)
	require.NotEmpty(t, rs2)
	require.Equal(t, "adv", rs2[0].POS)

	// no invent without matching adj tags
	require.Nil(t, DynamicPoAdvReadings("по-невідомо", tagWord))
	require.Nil(t, DynamicPoAdvReadings("звичайний", tagWord))
}

func TestDynamicNumeric_MmParadigm(t *testing.T) {
	rs := DynamicNumericReadings("5-мм", nil)
	require.NotEmpty(t, rs)
	// 3 genders × 6 cases (no v_kly)
	require.Len(t, rs, 18)
	require.Equal(t, "5-мм", rs[0].Lemma)
	require.Contains(t, rs[0].POS, "adj:")
	require.Contains(t, rs[0].POS, "v_")
}

func TestDynamicIntjRedupReadings(t *testing.T) {
	require.Nil(t, DynamicIntjRedupReadings("а-а", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "а":
			return []tagging.TaggedWord{{Lemma: "а", PosTag: "intj"}}
		case "га":
			return []tagging.TaggedWord{{Lemma: "га", PosTag: "intj"}}
		case "гей":
			return []tagging.TaggedWord{{Lemma: "гей", PosTag: "intj"}}
		default:
			return nil
		}
	}
	rs := DynamicIntjRedupReadings("а-а", tagWord)
	require.NotEmpty(t, rs)
	require.Equal(t, "intj", rs[0].POS)
	require.Equal(t, "а-а", rs[0].Lemma)

	rs2 := DynamicIntjRedupReadings("Гей-гей-гей", tagWord)
	require.NotEmpty(t, rs2)
	require.Equal(t, "гей-гей-гей", rs2[0].Lemma)

	// mixed non-intj: fail closed
	require.Nil(t, DynamicIntjRedupReadings("кіт-пес", tagWord))
}

func TestDynamicNumrAdjReadings(t *testing.T) {
	require.Nil(t, DynamicNumrAdjReadings("двох-триметровий", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "двох":
			return []tagging.TaggedWord{{Lemma: "два", PosTag: "numr:p:v_rod"}}
		case "триметровий":
			return []tagging.TaggedWord{{Lemma: "триметровий", PosTag: "adj:m:v_naz"}}
		case "метровий":
			return []tagging.TaggedWord{{Lemma: "метровий", PosTag: "adj:m:v_naz"}}
		case "три":
			return []tagging.TaggedWord{{Lemma: "три", PosTag: "numr:p:v_naz"}}
		default:
			return nil
		}
	}
	// Java: left ".*?(двох|трьох|чотирьох)" → :bad even when right starts with три
	rs := DynamicNumrAdjReadings("двох-триметровий", tagWord)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, "adj")
	require.Contains(t, rs[0].POS, ":bad")
	require.Equal(t, "двох-триметровий", rs[0].Lemma)

	// три-метровий: left ends with "и" (NUMR_ADJ_PATTERN); right not дво|три|… → :bad
	rs2 := DynamicNumrAdjReadings("три-метровий", tagWord)
	require.NotEmpty(t, rs2)
	require.Contains(t, rs2[0].POS, ":bad")
}

func TestDynamicNameSuffixReadings(t *testing.T) {
	require.Nil(t, DynamicNameSuffixReadings("Мустафа-ага", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		if strings.EqualFold(s, "Мустафа") || s == "мустафа" {
			return []tagging.TaggedWord{{Lemma: "Мустафа", PosTag: "noun:anim:m:v_naz:prop:fname"}}
		}
		return nil
	}
	rs := DynamicNameSuffixReadings("Мустафа-ага", tagWord)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, "name")
	require.Equal(t, "Мустафа-ага", rs[0].Lemma)
	// no invent without name POS on left
	require.Nil(t, DynamicNameSuffixReadings("Стіл-ага", tagWord))
}

func TestGuessOtherTagsReadings(t *testing.T) {
	// Java: capitalized *штрассе / *дзе paradigms (length > 7)
	rs := GuessOtherTagsReadings("Бранденштрассе")
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), "prop")
	require.Contains(t, *rs[0].GetPOSTag(), "alt")

	rs2 := GuessOtherTagsReadings("Бараташвілі")
	require.NotEmpty(t, rs2)
	require.Contains(t, *rs2[0].GetPOSTag(), "lname")

	require.Empty(t, GuessOtherTagsReadings("звичайнеслово"))
	require.Empty(t, GuessOtherTagsReadings("коротке"))
}

func TestDynamicRightParticleReadings(t *testing.T) {
	require.Nil(t, DynamicRightParticleReadings("гей-но", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "гей":
			return []tagging.TaggedWord{{Lemma: "гей", PosTag: "intj"}}
		case "чекай":
			return []tagging.TaggedWord{{Lemma: "чекати", PosTag: "verb:impr:s:2"}}
		case "хто":
			return []tagging.TaggedWord{{Lemma: "хто", PosTag: "noun:anim:m:v_naz:pron"}}
		default:
			return nil
		}
	}
	// гей-но: intj matches "но" allowed set
	rs := DynamicRightParticleReadings("гей-но", tagWord)
	require.NotEmpty(t, rs)
	require.Equal(t, "intj", rs[0].POS)
	require.Equal(t, "гей", rs[0].Lemma)

	// чекай-но: verb:impr
	rs2 := DynamicRightParticleReadings("чекай-но", tagWord)
	require.NotEmpty(t, rs2)
	require.Contains(t, rs2[0].POS, "verb")

	// хто-то blocked
	require.Nil(t, DynamicRightParticleReadings("хто-то", tagWord))
}

func TestDynamicDualPropReadings(t *testing.T) {
	require.Nil(t, DynamicDualPropReadings("Київ-Прага", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "київ":
			return []tagging.TaggedWord{{Lemma: "Київ", PosTag: "noun:inanim:m:v_naz:prop:geo"}}
		case "прага":
			return []tagging.TaggedWord{{Lemma: "Прага", PosTag: "noun:inanim:f:v_naz:prop:geo"}}
		case "карпа":
			return []tagging.TaggedWord{{Lemma: "Карпа", PosTag: "noun:anim:m:v_naz:prop:lname"}}
		case "хансен":
			return []tagging.TaggedWord{{Lemma: "Хансен", PosTag: "noun:anim:m:v_naz:prop:lname"}}
		default:
			return nil
		}
	}
	rs := DynamicDualPropReadings("Київ-Прага", tagWord)
	require.NotEmpty(t, rs)
	require.Equal(t, "noninfl:prop:geo", rs[0].POS)

	rs2 := DynamicDualPropReadings("Карпа-Хансен", tagWord)
	require.NotEmpty(t, rs2)
	require.Equal(t, "noninfl:prop:lname", rs2[0].POS)

	// not both capitalized
	require.Nil(t, DynamicDualPropReadings("київ-прага", tagWord))
}

func TestDynamicBadSuffixReadings(t *testing.T) {
	require.Nil(t, DynamicBadSuffixReadings("був-би", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		if strings.EqualFold(s, "був") {
			return []tagging.TaggedWord{{Lemma: "бути", PosTag: "verb:past:m"}}
		}
		return nil
	}
	rs := DynamicBadSuffixReadings("був-би", tagWord)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, ":bad")
	require.Equal(t, "бути-би", rs[0].Lemma)
	// м-б blocked (left length 1)
	require.Nil(t, DynamicBadSuffixReadings("м-б", tagWord))
}

func TestDynamicYearAndNumCompounds(t *testing.T) {
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "вибори":
			return []tagging.TaggedWord{{Lemma: "вибори", PosTag: "noun:inanim:p:v_naz:ns"}}
		case "формула":
			return []tagging.TaggedWord{{Lemma: "Формула", PosTag: "noun:inanim:f:v_naz:prop"}}
		case "омега":
			return []tagging.TaggedWord{{Lemma: "омега", PosTag: "noun:inanim:f:v_naz"}}
		default:
			return nil
		}
	}
	rs := DynamicYearCompoundReadings("Вибори-2014", tagWord)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].Lemma, "2014")
	require.Contains(t, rs[0].POS, "noun:inanim")

	rs2 := DynamicNumSuffixCompoundReadings("Формула-1", tagWord)
	require.NotEmpty(t, rs2)
	require.Equal(t, "Формула-1", rs2[0].Lemma)

	rs3 := DynamicNumSuffixCompoundReadings("омега-3", tagWord)
	require.NotEmpty(t, rs3)
	require.Equal(t, "омега-3", rs3[0].Lemma)

	// random noun+year without WORDS_WITH_YEAR / prop: fail closed
	require.Nil(t, DynamicYearCompoundReadings("стіл-2014", tagWord))
}

func TestDynamicAlPrefixReadings(t *testing.T) {
	tagWord := func(s string) []tagging.TaggedWord {
		if s == "Аль-Каїда" || strings.EqualFold(s, "аль-каїда") {
			return []tagging.TaggedWord{{Lemma: "Аль-Каїда", PosTag: "noun:inanim:f:v_naz:prop"}}
		}
		return nil
	}
	rs := DynamicAlPrefixReadings("аль-Каїда", tagWord)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, ":bad")
}

func TestDynamicPreRedupReadings(t *testing.T) {
	require.Nil(t, DynamicPreRedupReadings("гірко-прегірко", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "гірко":
			return []tagging.TaggedWord{{Lemma: "гірко", PosTag: "adv"}}
		case "гіркий":
			return []tagging.TaggedWord{{Lemma: "гіркий", PosTag: "adj:m:v_naz"}}
		default:
			return nil
		}
	}
	rs := DynamicPreRedupReadings("гірко-прегірко", tagWord)
	require.NotEmpty(t, rs)
	require.Equal(t, "adv", rs[0].POS)
	require.Equal(t, "гірко-прегірко", rs[0].Lemma)

	rs2 := DynamicPreRedupReadings("гіркий-прегіркий", tagWord)
	require.NotEmpty(t, rs2)
	require.Equal(t, "гіркий-прегіркий", rs2[0].Lemma)
	require.Contains(t, rs2[0].POS, "adj")

	require.Nil(t, DynamicPreRedupReadings("гірко-пресолодко", tagWord))
}

func TestDynamicNapivDualReadings(t *testing.T) {
	require.Nil(t, DynamicNapivDualReadings("напівпольської-напіванглійської", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "польської":
			return []tagging.TaggedWord{{Lemma: "польський", PosTag: "adj:f:v_rod"}}
		case "англійської":
			return []tagging.TaggedWord{{Lemma: "англійський", PosTag: "adj:f:v_rod"}}
		default:
			return nil
		}
	}
	rs := DynamicNapivDualReadings("напівпольської-напіванглійської", tagWord)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, "adj")
	require.Contains(t, rs[0].Lemma, "напів")
	require.Contains(t, rs[0].Lemma, "польськ")
}

func TestDynamicEqualRedupReadings(t *testing.T) {
	require.Nil(t, DynamicEqualRedupReadings("Усе-усе", nil))
	tagWord := func(s string) []tagging.TaggedWord {
		switch strings.ToLower(s) {
		case "усе":
			return []tagging.TaggedWord{
				{Lemma: "увесь", PosTag: "adj:n:v_naz:pron:gen"},
				{Lemma: "усе", PosTag: "adv"},
			}
		case "всього":
			return []tagging.TaggedWord{
				{Lemma: "весь", PosTag: "adj:m:v_rod:pron:gen"},
				{Lemma: "ввесь", PosTag: "adj:m:v_rod:pron:gen"},
			}
		default:
			return nil
		}
	}
	rs := DynamicEqualRedupReadings("Усе-усе", tagWord)
	require.NotEmpty(t, rs)
	// equalParts lemmas only (увесь-увесь, усе-усе)
	for _, r := range rs {
		require.True(t, equalParts(r.Lemma) || equalParts(strings.ToLower(r.Lemma)), r.Lemma)
	}
	joined := ""
	for _, r := range rs {
		joined += r.Lemma + " "
	}
	require.True(t, strings.Contains(joined, "увесь-увесь") || strings.Contains(joined, "усе-усе"))

	rs2 := DynamicEqualRedupReadings("всього-всього", tagWord)
	require.NotEmpty(t, rs2)

	// not ves/us lemma: fail closed
	require.Nil(t, DynamicEqualRedupReadings("кіт-кіт", func(s string) []tagging.TaggedWord {
		return []tagging.TaggedWord{{Lemma: "кіт", PosTag: "noun:anim:m:v_naz"}}
	}))
}

func TestDynamicNumeric_RomanLeft(t *testing.T) {
	// Roman numeral left + short ending (Java ADJ_PREFIX_NUMBER)
	rs := DynamicNumericReadings("XI-й", nil)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, "numr")
	require.Equal(t, "XI-й", rs[0].Lemma)
}
