package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestNumericLongForm_Riccha(t *testing.T) {
	// п'ятсотдвадцятип'ятиріччя — groups num + tag("сторіччя")
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "п'ятсот", "пятсот":
			return []tagging.TaggedWord{{Lemma: "п'ятсот", PosTag: "numr:p:v_naz"}}
		case "двадцяти":
			return []tagging.TaggedWord{{Lemma: "двадцять", PosTag: "numr:p:v_rod"}}
		case "п'яти", "пяти":
			return []tagging.TaggedWord{{Lemma: "п'ять", PosTag: "numr:p:v_rod"}}
		case "сторіччя":
			return []tagging.TaggedWord{
				{Lemma: "сторіччя", PosTag: "noun:inanim:n:v_naz"},
				{Lemma: "сторіччя", PosTag: "noun:inanim:n:v_rod"},
				{Lemma: "сторіччя", PosTag: "noun:inanim:n:v_kly"}, // skipped
				{Lemma: "сторіччя", PosTag: "noun:inanim:p:v_naz"}, // :p: skipped
			}
		}
		return nil
	}
	word := "п'ятсотдвадцятип'ятиріччя"
	rs := NumericLongFormReadings(word, tag)
	require.NotEmpty(t, rs)
	for _, r := range rs {
		require.NotContains(t, *r.GetPOSTag(), "v_kly")
		require.NotContains(t, *r.GetPOSTag(), ":p:")
		// lemma = groups + strip "сто" from сторіччя → річчя
		require.Equal(t, "п'ятсотдвадцятип'ятиріччя", *r.GetLemma())
	}
	// fail closed: no dict for сто+end
	require.Empty(t, NumericLongFormReadings(word, func(string) []tagging.TaggedWord { return nil }))
	// fail closed: non-num middle group
	require.Empty(t, NumericLongFormReadings(word, func(w string) []tagging.TaggedWord {
		if w == "сторіччя" {
			return []tagging.TaggedWord{{Lemma: "сторіччя", PosTag: "noun:inanim:n:v_naz"}}
		}
		if w == "двадцяти" {
			return []tagging.TaggedWord{{Lemma: "двадцять", PosTag: "noun:inanim:n:v_rod"}} // not num
		}
		return []tagging.TaggedWord{{Lemma: w, PosTag: "numr:p:v_naz"}}
	}))
}

func TestNumericLongForm_Otyi(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "сто":
			return []tagging.TaggedWord{{Lemma: "сто", PosTag: "numr:p:v_naz"}}
		case "двадцяти":
			return []tagging.TaggedWord{{Lemma: "двадцять", PosTag: "numr:p:v_rod"}}
		case "п'яти", "пяти":
			return []tagging.TaggedWord{{Lemma: "п'ять", PosTag: "numr:p:v_rod"}}
		case "мільйонний":
			return []tagging.TaggedWord{
				{Lemma: "мільйонний", PosTag: "adj:m:v_naz:numr"},
				{Lemma: "мільйонний", PosTag: "noun:inanim:m:v_naz"}, // skipped (not adj)
			}
		case "річний":
			return []tagging.TaggedWord{{Lemma: "річний", PosTag: "adj:m:v_naz"}}
		}
		return nil
	}
	// стодвадцятип'ятирічний — group4 is річний via OTYI (річн…)
	word := "стодвадцятип'ятирічний"
	rs := NumericLongFormReadings(word, tag)
	require.NotEmpty(t, rs)
	require.True(t, stringsHasPrefix(*rs[0].GetPOSTag(), "adj"))
	require.Equal(t, "стодвадцятип'ятирічний", *rs[0].GetLemma())

	// short word skipped
	require.Empty(t, NumericLongFormReadings("річчя", tag))
}

func stringsHasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

func TestAltDashReadings(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "ось-ось" {
			return []tagging.TaggedWord{{Lemma: "ось-ось", PosTag: "adv"}}
		}
		return nil
	}
	rs := AltDashReadings("ось\u2013ось", tag)
	require.GreaterOrEqual(t, len(rs), 2)
	require.Equal(t, "adv", *rs[0].GetPOSTag())
	// last is null for original surface
	require.Nil(t, rs[len(rs)-1].GetPOSTag())
	require.Empty(t, AltDashReadings("ось-ось", tag))
}

func TestBracketAltReadings(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "слово" {
			return []tagging.TaggedWord{{Lemma: "слово", PosTag: "noun:inanim:n:v_naz"}}
		}
		return nil
	}
	rs := BracketAltReadings("сл[о]во", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	require.Empty(t, BracketAltReadings("слово", tag))
}

func TestSolidLeftOAdjInvalid(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "західний" || w == "Західний" {
			return []tagging.TaggedWord{{Lemma: "західний", PosTag: "adj:m:v_naz"}}
		}
		return nil
	}
	// південнозахідний solid len>=9
	rs := SolidLeftOAdjInvalidReadings("південнозахідний", tag)
	require.NotEmpty(t, rs)
	require.Equal(t, "південнозахідний", *rs[0].GetLemma())
	// hyphenated skipped (oAdj path)
	require.Empty(t, SolidLeftOAdjInvalidReadings("південно-західний", tag))
}
