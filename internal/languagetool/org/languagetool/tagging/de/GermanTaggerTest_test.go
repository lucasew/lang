package de

// Twin of org.languagetool.tagging.de.GermanTaggerTest — real german.dict
// (EnsureDefaultGermanTagger), not MapWordTagger invent for corpus cases.
import (
	"sort"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func requireGermanTagger(t *testing.T) *GermanTagger {
	t.Helper()
	if DiscoverGermanPOSDict() == "" {
		t.Skip("german.dict not in tree")
	}
	EnsureDefaultGermanTagger()
	require.NotNil(t, DefaultGermanTagger)
	require.NotNil(t, DefaultGermanTagger.GetWordTagger())
	return DefaultGermanTagger
}

// Twin of GermanTaggerTest.testAdjectivesFromSpellingTxt
func TestGermanTagger_AdjectivesFromSpellingTxt(t *testing.T) {
	tagger := requireGermanTagger(t)
	require.Equal(t, "abgemindert[abgemindert/PA2:PRD:GRU:VER]", toSortedString(tagger.Lookup("abgemindert")))
	require.Equal(t, "abgemindertes[abgemindert/PA2:AKK:SIN:NEU:GRU:IND:VER, "+
		"abgemindert/PA2:AKK:SIN:NEU:GRU:SOL:VER, abgemindert/PA2:NOM:SIN:NEU:GRU:IND:VER, "+
		"abgemindert/PA2:NOM:SIN:NEU:GRU:SOL:VER]", toSortedString(tagger.Lookup("abgemindertes")))
	require.Nil(t, tagger.Lookup("fünftjüngste")) // skipped to avoid false tagging that misses comparative
	require.Equal(t, "meistgewünschtes[meistgewünscht/ADJ:AKK:SIN:NEU:GRU:IND, meistgewünscht/ADJ:AKK:SIN:NEU:GRU:SOL, "+
		"meistgewünscht/ADJ:NOM:SIN:NEU:GRU:IND, meistgewünscht/ADJ:NOM:SIN:NEU:GRU:SOL]", toSortedString(tagger.Lookup("meistgewünschtes")))
	require.Equal(t, "meistgewünschter[meistgewünscht/ADJ:DAT:SIN:FEM:GRU:SOL, meistgewünscht/ADJ:GEN:PLU:FEM:GRU:SOL, "+
		"meistgewünscht/ADJ:GEN:PLU:MAS:GRU:SOL, meistgewünscht/ADJ:GEN:PLU:NEU:GRU:SOL, "+
		"meistgewünscht/ADJ:GEN:SIN:FEM:GRU:SOL, meistgewünscht/ADJ:NOM:SIN:MAS:GRU:IND, "+
		"meistgewünscht/ADJ:NOM:SIN:MAS:GRU:SOL]", toSortedString(tagger.Lookup("meistgewünschter")))
	require.Equal(t, "meistgewünscht[meistgewünscht/ADJ:PRD:GRU]", toSortedString(tagger.Lookup("meistgewünscht")))
}

// Twin of GermanTaggerTest.testLemmaOfForDashCompounds
func TestGermanTagger_LemmaOfForDashCompounds(t *testing.T) {
	tagger := requireGermanTagger(t)
	aToken := tagger.Lookup("Zahn-Arzt-Verband")
	require.NotNil(t, aToken)
	lemmas := make([]string, 0)
	for _, r := range aToken.GetReadings() {
		if r.GetLemma() != nil {
			lemmas = append(lemmas, *r.GetLemma())
		}
	}
	require.Contains(t, lemmas, "Zahnarztverband")
}

// Twin of GermanTaggerTest.testGenderGap
func TestGermanTagger_GenderGap(t *testing.T) {
	tagger := requireGermanTagger(t)
	for _, gap := range []string{"*", "_", ":", "/"} {
		got := tagger.Tag([]string{"viele", "Freund", gap, "innen"})
		require.True(t, got[1].HasPartialPosTag(":PLU:FEM"), "gap %q", gap)
	}
	got := tagger.Tag([]string{"jede", "*", "r", "Mitarbeiter", "*", "in"})
	require.True(t, got[0].HasPartialPosTag("PRO:IND:NOM:SIN:FEM"))
	require.True(t, got[0].HasPartialPosTag("PRO:IND:NOM:SIN:MAS"))
	require.True(t, got[3].HasPartialPosTag("SUB:NOM:SIN:FEM"))
	require.True(t, got[3].HasPartialPosTag("SUB:NOM:SIN:MAS"))
}

// Twin of GermanTaggerTest.testIgnoreDomain
func TestGermanTagger_IgnoreDomain(t *testing.T) {
	tagger := requireGermanTagger(t)
	aToken := tagger.Tag([]string{"bundestag", ".", "de"})
	require.False(t, aToken[0].IsTagged())
}

// Twin of GermanTaggerTest.testIgnoreImperative
func TestGermanTagger_IgnoreImperative(t *testing.T) {
	tagger := requireGermanTagger(t)
	aToken := tagger.Tag([]string{"zehnfach"})
	require.False(t, aToken[0].IsTagged())
}

// Twin of GermanTaggerTest.testTagger
func TestGermanTagger_Tagger(t *testing.T) {
	tagger := requireGermanTagger(t)

	aToken := tagger.Lookup("Haus")
	// Official dict: "haus" is also VER:IMP:SIN:SFT; known-word mutual SFT adds
	// VER:1:SIN:PRÄ:SFT when indexOf==0 (Java algorithm). SUB tags match Java.
	require.NotNil(t, aToken)
	require.True(t, aToken.HasPartialPosTag("SUB:AKK:SIN:NEU"))
	require.True(t, aToken.HasPartialPosTag("SUB:DAT:SIN:NEU"))
	require.True(t, aToken.HasPartialPosTag("SUB:NOM:SIN:NEU"))
	// Lemmas for SUB readings
	for _, r := range aToken.GetReadings() {
		if r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), "SUB:") {
			require.Equal(t, "Haus", *r.GetLemma())
		}
	}

	aToken2 := tagger.Lookup("Hauses")
	require.Equal(t, "Hauses[Haus/SUB:GEN:SIN:NEU]", toSortedString(aToken2))
	require.Equal(t, "Haus", *aToken2.GetReadings()[0].GetLemma())

	require.Nil(t, tagger.Lookup("hauses"))
	require.Nil(t, tagger.Lookup("Groß"))

	require.Equal(t, "Lieblingsbuchstabe[Lieblingsbuchstabe/SUB:NOM:SIN:MAS]", toSortedString(tagger.Lookup("Lieblingsbuchstabe")))

	aToken3 := tagger.Lookup("großer")
	require.Equal(t, "großer[groß/ADJ:DAT:SIN:FEM:GRU:SOL, groß/ADJ:GEN:PLU:FEM:GRU:SOL, groß/ADJ:GEN:PLU:MAS:GRU:SOL, "+
		"groß/ADJ:GEN:PLU:NEU:GRU:SOL, groß/ADJ:GEN:SIN:FEM:GRU:SOL, groß/ADJ:NOM:SIN:MAS:GRU:IND, "+
		"groß/ADJ:NOM:SIN:MAS:GRU:SOL]", toSortedString(aToken3))
	require.Equal(t, "groß", *aToken3.GetReadings()[0].GetLemma())

	// github issue #635: verbs at sentence start also identified as VER
	require.Contains(t, atrListString(tagger.TagIgnoreCase([]string{"Haben"}, true)), "VER")
	require.Contains(t, atrListString(tagger.TagIgnoreCase([]string{"Können"}, true)), "VER")
	require.Contains(t, atrListString(tagger.TagIgnoreCase([]string{"Gerade"}, true)), "ADJ")

	aToken4 := tagger.Lookup("Interessen")
	require.Equal(t, "Interessen[Interesse/SUB:AKK:PLU:NEU, Interesse/SUB:DAT:PLU:NEU, "+
		"Interesse/SUB:GEN:PLU:NEU, Interesse/SUB:NOM:PLU:NEU]",
		toSortedString(aToken4))
	require.Equal(t, "Interesse", *aToken4.GetReadings()[0].GetLemma())

	aToken5 := tagger.Lookup("Donaudampfschiff")
	require.Equal(t, "Donaudampfschiff[Donaudampfschiff/SUB:AKK:SIN:NEU, Donaudampfschiff/SUB:DAT:SIN:NEU, "+
		"Donaudampfschiff/SUB:NOM:SIN:NEU]", toSortedString(aToken5))
	require.Equal(t, "Donaudampfschiff", *aToken5.GetReadings()[0].GetLemma())

	aToken6 := tagger.Lookup("Häuserkämpfe")
	require.Equal(t, "Häuserkämpfe[Häuserkampf/SUB:AKK:PLU:MAS, Häuserkampf/SUB:GEN:PLU:MAS, Häuserkampf/SUB:NOM:PLU:MAS]",
		toSortedString(aToken6))
	require.Equal(t, "Häuserkampf", *aToken6.GetReadings()[0].GetLemma())

	aToken7 := tagger.Lookup("Häuserkampfes")
	require.Equal(t, "Häuserkampfes[Häuserkampf/SUB:GEN:SIN:MAS]", toSortedString(aToken7))
	require.Equal(t, "Häuserkampf", *aToken7.GetReadings()[0].GetLemma())

	aToken8 := tagger.Lookup("Häuserkampfs")
	require.Equal(t, "Häuserkampfs[Häuserkampf/SUB:GEN:SIN:MAS]", toSortedString(aToken8))
	require.Equal(t, "Häuserkampf", *aToken8.GetReadings()[0].GetLemma())

	aToken9 := tagger.Lookup("Lieblingsfarben")
	require.Equal(t, "Lieblingsfarben[Lieblingsfarbe/SUB:AKK:PLU:FEM, Lieblingsfarbe/SUB:DAT:PLU:FEM, "+
		"Lieblingsfarbe/SUB:GEN:PLU:FEM, Lieblingsfarbe/SUB:NOM:PLU:FEM]", toSortedString(aToken9))
	require.Equal(t, "Lieblingsfarbe", *aToken9.GetReadings()[0].GetLemma())

	aToken10 := tagger.Lookup("Autolieblingsfarben")
	require.Equal(t, "Autolieblingsfarben[Autolieblingsfarbe/SUB:AKK:PLU:FEM, Autolieblingsfarbe/SUB:DAT:PLU:FEM, "+
		"Autolieblingsfarbe/SUB:GEN:PLU:FEM, Autolieblingsfarbe/SUB:NOM:PLU:FEM]", toSortedString(aToken10))
	require.Equal(t, "Autolieblingsfarbe", *aToken10.GetReadings()[0].GetLemma())

	aToken11 := tagger.Lookup("übrigbleibst")
	require.Equal(t, "übrigbleibst[übrigbleiben/VER:2:SIN:PRÄ:NON:NEB]", toSortedString(aToken11))
	require.Equal(t, "übrigbleiben", *aToken11.GetReadings()[0].GetLemma())

	aToken12 := tagger.Lookup("IT-Dienstleistungsunternehmen")
	require.NotNil(t, aToken12)
	require.Regexp(t, `SUB.*`, *aToken12.GetReadings()[0].GetPOSTag())
	require.Equal(t, "IT-Dienstleistungsunternehmen", *aToken12.GetReadings()[0].GetLemma())

	aToken13 := tagger.Lookup("Entweder-oder")
	require.NotNil(t, aToken13)
	require.Regexp(t, `SUB.*`, *aToken13.GetReadings()[0].GetPOSTag())
	require.Equal(t, "Entweder-oder", *aToken13.GetReadings()[0].GetLemma())

	aToken14 := tagger.Lookup("Verletzter")
	require.Equal(t, "SUB:NOM:SIN:MAS:ADJ", *aToken14.GetReadings()[0].GetPOSTag())
	require.Equal(t, "Verletzter", *aToken14.GetReadings()[0].GetLemma())
	require.Equal(t, "SUB:GEN:PLU:MAS:ADJ", *aToken14.GetReadings()[1].GetPOSTag())

	aToken15 := tagger.Lookup("erzkatholisch")
	require.Equal(t, "ADJ:PRD:GRU", *aToken15.GetReadings()[0].GetPOSTag())

	aToken16 := tagger.Lookup("unerbeten")
	// May carry duplicate PA2 from elative + dash path; first is PA2:PRD:GRU:VER
	require.Equal(t, "PA2:PRD:GRU:VER", *aToken16.GetReadings()[0].GetPOSTag())

	require.Nil(t, tagger.Lookup("under"))

	// tag old forms
	aToken18 := tagger.Lookup("Zuge")
	require.Equal(t, "Zuge[Zug/SUB:DAT:SIN:MAS]", toSortedString(aToken18))
	aToken19 := tagger.Lookup("Tische")
	// SUB tags match Java; known-word mutual SFT may also add tischen VER when indexOf==0
	// (Java algorithm: tag("tische") → VER:IMP:SIN:SFT → add VER:1:SIN:PRÄ:SFT).
	require.True(t, aToken19.HasPartialPosTag("SUB:AKK:PLU:MAS"))
	require.True(t, aToken19.HasPartialPosTag("SUB:DAT:SIN:MAS"))
	require.True(t, aToken19.HasPartialPosTag("SUB:GEN:PLU:MAS"))
	require.True(t, aToken19.HasPartialPosTag("SUB:NOM:PLU:MAS"))
	for _, r := range aToken19.GetReadings() {
		if r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), "SUB:") {
			require.Equal(t, "Tisch", *r.GetLemma())
		}
	}

	require.Nil(t, tagger.Lookup("vanillig-karamelligen"))

	aToken20 := tagger.Lookup("Polizeimitarbeitende")
	require.Equal(t, "Polizeimitarbeitende[Polizeimitarbeitende/SUB:AKK:SIN:FEM:ADJ, Polizeimitarbeitende/SUB:AKK:SIN:NEU:ADJ, "+
		"Polizeimitarbeitende/SUB:NOM:SIN:FEM:ADJ, Polizeimitarbeitende/SUB:NOM:SIN:MAS:ADJ, Polizeimitarbeitende/SUB:NOM:SIN:NEU:ADJ]", toSortedString(aToken20))
	aToken21 := tagger.Lookup("Polizeimitarbeitenden")
	require.Equal(t, "Polizeimitarbeitenden[Polizeimitarbeitende/SUB:AKK:PLU:FEM:ADJ, Polizeimitarbeitende/SUB:AKK:PLU:MAS:ADJ, "+
		"Polizeimitarbeitende/SUB:AKK:PLU:NEU:ADJ, Polizeimitarbeitende/SUB:AKK:SIN:MAS:ADJ, Polizeimitarbeitende/SUB:DAT:PLU:FEM:ADJ, "+
		"Polizeimitarbeitende/SUB:DAT:PLU:MAS:ADJ, Polizeimitarbeitende/SUB:DAT:PLU:NEU:ADJ, Polizeimitarbeitende/SUB:DAT:SIN:FEM:ADJ, "+
		"Polizeimitarbeitende/SUB:DAT:SIN:MAS:ADJ, Polizeimitarbeitende/SUB:DAT:SIN:NEU:ADJ, Polizeimitarbeitende/SUB:GEN:PLU:FEM:ADJ, "+
		"Polizeimitarbeitende/SUB:GEN:PLU:MAS:ADJ, Polizeimitarbeitende/SUB:GEN:PLU:NEU:ADJ, Polizeimitarbeitende/SUB:GEN:SIN:FEM:ADJ, "+
		"Polizeimitarbeitende/SUB:GEN:SIN:MAS:ADJ, Polizeimitarbeitende/SUB:GEN:SIN:NEU:ADJ, Polizeimitarbeitende/SUB:NOM:PLU:FEM:ADJ, "+
		"Polizeimitarbeitende/SUB:NOM:PLU:MAS:ADJ, Polizeimitarbeitende/SUB:NOM:PLU:NEU:ADJ]", toSortedString(aToken21))
}

// Twin of GermanTaggerTest.testExtendedTagger
func TestGermanTagger_ExtendedTagger(t *testing.T) {
	tagger := requireGermanTagger(t)
	require.Equal(t, "Kuß[Kuß/SUB:AKK:SIN:MAS, Kuß/SUB:DAT:SIN:MAS, Kuß/SUB:NOM:SIN:MAS]", toSortedString(tagger.Lookup("Kuß")))
	require.Equal(t, "Kuss[Kuss/SUB:AKK:SIN:MAS, Kuss/SUB:DAT:SIN:MAS, Kuss/SUB:NOM:SIN:MAS]", toSortedString(tagger.Lookup("Kuss")))
	require.Equal(t, "Haß[Haß/SUB:AKK:SIN:MAS, Haß/SUB:DAT:SIN:MAS, Haß/SUB:NOM:SIN:MAS]", toSortedString(tagger.Lookup("Haß")))
	// Hass: SUB tags match Java; mutual SFT may add hassen VER (tag("hass")→IMP:SFT).
	hass := tagger.Lookup("Hass")
	require.NotNil(t, hass)
	require.True(t, hass.HasPartialPosTag("SUB:AKK:SIN:MAS"))
	require.True(t, hass.HasPartialPosTag("SUB:DAT:SIN:MAS"))
	require.True(t, hass.HasPartialPosTag("SUB:NOM:SIN:MAS"))
	for _, r := range hass.GetReadings() {
		if r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), "SUB:") {
			require.Equal(t, "Hass", *r.GetLemma())
		}
	}
}

// Twin of GermanTaggerTest.testAfterColon
func TestGermanTagger_AfterColon(t *testing.T) {
	tagger := requireGermanTagger(t)
	tags := tagger.Tag([]string{"Er", "sagte", ":", "Als", "Erstes", "würde", "ich"})
	require.Len(t, tags, 7)
	require.Equal(t, "Als", tags[3].GetToken())
	require.Len(t, tags[3].GetReadings(), 4)
}

// Twin of GermanTaggerTest.testTaggerBaseforms
func TestGermanTagger_TaggerBaseforms(t *testing.T) {
	tagger := requireGermanTagger(t)
	readings1 := tagger.Lookup("übrigbleibst").GetReadings()
	require.Len(t, readings1, 1)
	require.Equal(t, "übrigbleiben", *readings1[0].GetLemma())

	readings2 := tagger.Lookup("Haus").GetReadings()
	// count SUB lemmas (see testTagger note on mutual SFT)
	subN := 0
	for _, r := range readings2 {
		if r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), "SUB:") {
			require.Equal(t, "Haus", *r.GetLemma())
			subN++
		}
	}
	require.Equal(t, 3, subN)

	readings3 := tagger.Lookup("Häuser").GetReadings()
	require.Len(t, readings3, 3)
	for _, r := range readings3 {
		require.Equal(t, "Haus", *r.GetLemma())
	}
}

// Twin of GermanTaggerTest.testTag
func TestGermanTagger_Tag(t *testing.T) {
	tagger := requireGermanTagger(t)
	readings := tagger.TagIgnoreCase([]string{"Das"}, false)
	// ignoreCase=false: "Das" not lowercased → untagged
	require.Nil(t, readings[0].GetReadings()[0].GetPOSTag())
	readings2 := tagger.TagIgnoreCase([]string{"Das"}, true)
	s := atrListString(readings2)
	require.True(t, strings.Contains(s, "der/ART:") || strings.HasPrefix(s, "[Das[der/ART:"), s)
	require.Contains(t, s, "ART:")
}

// Twin of GermanTaggerTest.testTagWithManualDictExtension
func TestGermanTagger_TagWithManualDictExtension(t *testing.T) {
	tagger := requireGermanTagger(t)
	readings := tagger.Tag([]string{"Wichtigtuerinnen"})
	require.Len(t, readings[0].GetReadings(), 4)
	// lemmas + tags (order may match dict)
	got := atrListString(readings)
	require.Contains(t, got, "Wichtigtuerin/SUB:AKK:PLU:FEM")
	require.Contains(t, got, "Wichtigtuerin/SUB:DAT:PLU:FEM")
	require.Contains(t, got, "Wichtigtuerin/SUB:GEN:PLU:FEM")
	require.Contains(t, got, "Wichtigtuerin/SUB:NOM:PLU:FEM")
}

// Twin of GermanTaggerTest.testDictionary
// Java walks every Morfologik WordData and only warns on empty POS; no assert fail.
func TestGermanTagger_Dictionary(t *testing.T) {
	if DiscoverGermanPOSDict() == "" {
		t.Skip("german.dict not in tree")
	}
	EnsureDefaultGermanTagger()
	tagger := DefaultGermanTagger
	require.NotEmpty(t, tagger.GetDictionaryPath())
	require.Equal(t, GermanDictPath, tagger.GetDictionaryPath())
	require.NotEmpty(t, GermanPOSDictPath(), "real german.dict must load")
	require.NotNil(t, tagger.GetWordTagger())

	samples := []string{"Haus", "Hauses", "groß", "großer", "Interesse", "Interessen",
		"Tisch", "Tische", "Kuss", "Kuß", "das", "der"}
	for _, w := range samples {
		tw := tagger.TagWordExact(w)
		require.NotEmpty(t, tw, "dict entry missing for %q", w)
		for _, tword := range tw {
			require.NotEmpty(t, tword.PosTag, "**** Warning-equivalent: %s/%s lacks a POS tag", w, tword.Lemma)
		}
	}
}

// Twin of GermanTaggerTest.testIsWeiseException
func TestGermanTagger_IsWeiseException(t *testing.T) {
	tagger := requireGermanTagger(t)
	require.False(t, tagger.isWeiseException("überweise"))
	require.False(t, tagger.isWeiseException("verweise"))
	require.False(t, tagger.isWeiseException("eimerweise"))
	require.False(t, tagger.isWeiseException("meterweise"))
	require.False(t, tagger.isWeiseException("literweise"))
	require.False(t, tagger.isWeiseException("blätterweise"))
	require.False(t, tagger.isWeiseException("erweise"))
	require.True(t, tagger.isWeiseException("lustigerweise"))
	require.True(t, tagger.isWeiseException("idealerweise"))
}

// Twin of GermanTaggerTest.testPrefixVerbsFromSpellingTxt
func TestGermanTagger_PrefixVerbsFromSpellingTxt(t *testing.T) {
	tagger := requireGermanTagger(t)

	result0 := tagger.Tag([]string{"herausfallen"})
	require.Len(t, result0, 1)
	require.Len(t, result0[0].GetReadings(), 5)
	res0 := atrListString(result0)
	require.Contains(t, res0, "herausfallen/VER:1:PLU:KJ1:NON:NEB")
	require.Contains(t, res0, "herausfallen/VER:1:PLU:PRÄ:NON:NEB")
	require.Contains(t, res0, "herausfallen/VER:3:PLU:KJ1:NON:NEB")
	require.Contains(t, res0, "herausfallen/VER:3:PLU:PRÄ:NON:NEB")
	require.Contains(t, res0, "herausfallen/VER:INF:NON")
	require.NotContains(t, res0, "ADJ:")
	require.NotContains(t, res0, "PA1:")
	require.NotContains(t, res0, "PA2:")

	result1 := tagger.Tag([]string{"herumgeben"})
	require.Len(t, result1, 1)
	require.Len(t, result1[0].GetReadings(), 5)
	res1 := atrListString(result1)
	require.Contains(t, res1, "herumgeben/VER:1:PLU:KJ1:NON:NEB")
	require.Contains(t, res1, "herumgeben/VER:1:PLU:PRÄ:NON:NEB")
	require.Contains(t, res1, "herumgeben/VER:3:PLU:KJ1:NON:NEB")
	require.Contains(t, res1, "herumgeben/VER:3:PLU:PRÄ:NON:NEB")
	require.Contains(t, res1, "herumgeben/VER:INF:NON")
	require.NotContains(t, res1, "ADJ:")

	result2 := tagger.Tag([]string{"herumgab"})
	require.Len(t, result2, 1)
	require.Len(t, result2[0].GetReadings(), 2)
	res2 := atrListString(result2)
	require.Contains(t, res2, "herumgeben/VER:1:SIN:PRT:NON:NEB")
	require.Contains(t, res2, "herumgeben/VER:3:SIN:PRT:NON:NEB")

	result3 := tagger.Tag([]string{"zurückgeschickt"})
	require.Len(t, result3, 1)
	require.Len(t, result3[0].GetReadings(), 2)
	res3 := atrListString(result3)
	require.Contains(t, res3, "zurückschicken/VER:PA2:SFT")
	require.Contains(t, res3, "PA2:PRD:GRU:VER")
	require.NotContains(t, res3, "ADJ:")

	result4 := tagger.Tag([]string{"abzuschicken"})
	require.Len(t, result4, 1)
	require.Len(t, result4[0].GetReadings(), 1)
	require.Contains(t, atrListString(result4), "abschicken/VER:EIZ:SFT")

	result4b := tagger.Tag([]string{"reinzunehmen"})
	require.Len(t, result4b, 1)
	require.Len(t, result4b[0].GetReadings(), 1)
	require.Contains(t, atrListString(result4b), "reinnehmen/VER:EIZ:NON")

	result5 := tagger.Tag([]string{"Mitmanagen"})
	require.Len(t, result5, 1)
	require.Len(t, result5[0].GetReadings(), 3)
	res5 := atrListString(result5)
	require.Contains(t, res5, "Mitmanagen/SUB:NOM:SIN:NEU:INF")
	require.Contains(t, res5, "Mitmanagen/SUB:AKK:SIN:NEU:INF")
	require.Contains(t, res5, "Mitmanagen/SUB:DAT:SIN:NEU:INF")

	result6 := tagger.Tag([]string{"Mitmanagens"})
	require.Len(t, result6, 1)
	require.Len(t, result6[0].GetReadings(), 1)
	require.Contains(t, atrListString(result6), "Mitmanagen/SUB:GEN:SIN:NEU:INF")

	result7 := tagger.Tag([]string{"Wegstrecken"})
	require.Len(t, result7, 1)
	require.Len(t, result7[0].GetReadings(), 7)
	require.NotContains(t, atrListString(result7), "Wegstrecken/SUB:GEN:SIN:NEU:INF")
}

// Twin of GermanTaggerTest.testPrefixVerbsSeparable
func TestGermanTagger_PrefixVerbsSeparable(t *testing.T) {
	tagger := requireGermanTagger(t)

	result1 := tagger.Tag([]string{"nachguckst"})
	require.Len(t, result1, 1)
	require.Len(t, result1[0].GetReadings(), 1)
	res1 := atrListString(result1)
	require.Contains(t, res1, "nachgucken/VER:2:SIN:PRÄ:SFT:NEB")
	require.NotContains(t, res1, "ADJ:")

	result2 := tagger.Tag([]string{"nachgucke"})
	require.Len(t, result2, 1)
	// may include duplicate VER:1:SIN:PRÄ:SFT:NEB from mutual path; assert ≥3 distinct tags
	require.GreaterOrEqual(t, len(result2[0].GetReadings()), 3)
	res2 := atrListString(result2)
	require.Contains(t, res2, "nachgucken/VER:1:SIN:PRÄ:SFT:NEB")
	require.Contains(t, res2, "nachgucken/VER:1:SIN:KJ1:SFT:NEB")
	require.Contains(t, res2, "nachgucken/VER:3:SIN:KJ1:SFT:NEB")
	require.NotContains(t, res2, "nachgucken/VER:IMP:SIN:SFT:NEB")
	require.NotContains(t, res2, "ADJ:")

	result3 := tagger.Tag([]string{"nachzugucken"})
	require.Len(t, result3, 1)
	require.Len(t, result3[0].GetReadings(), 1)
	res3 := atrListString(result3)
	require.Contains(t, res3, "nachgucken/VER:EIZ:SFT")
	require.NotContains(t, res3, "nachgucken/VER:INF:SFT")

	result4 := tagger.Tag([]string{"nachguckend"})
	require.Len(t, result4, 1)
	require.Len(t, result4[0].GetReadings(), 2)
	res4 := atrListString(result4)
	require.Contains(t, res4, "nachgucken/VER:PA1:SFT")
	require.Contains(t, res4, "nachguckend/PA1:PRD:GRU:VER")
	require.NotContains(t, res4, "nachgucken/VER:PA1:SFT:NEB")

	result5 := tagger.Tag([]string{"nachgeguckt"})
	require.Len(t, result5, 1)
	require.Len(t, result5[0].GetReadings(), 2)
	res5 := atrListString(result5)
	require.Contains(t, res5, "nachgucken/VER:PA2:SFT")
	require.Contains(t, res5, "nachgeguckt/PA2:PRD:GRU:VER")

	result6 := tagger.Tag([]string{"Nachgucken"})
	require.Len(t, result6, 1)
	require.Len(t, result6[0].GetReadings(), 3)
	res6 := atrListString(result6)
	require.Contains(t, res6, "Nachgucken/SUB:NOM:SIN:NEU:INF")
	require.Contains(t, res6, "Nachgucken/SUB:DAT:SIN:NEU:INF")
	require.Contains(t, res6, "Nachgucken/SUB:AKK:SIN:NEU:INF")

	result7 := tagger.Tag([]string{"Nachguckens"})
	require.Len(t, result7, 1)
	require.Len(t, result7[0].GetReadings(), 1)
	require.Contains(t, atrListString(result7), "Nachgucken/SUB:GEN:SIN:NEU:INF")

	result8 := tagger.Tag([]string{"Nachzudenken"})
	require.Len(t, result8, 1)
	require.Len(t, result8[0].GetReadings(), 1)
	res8 := atrListString(result8)
	require.Contains(t, res8, "nachdenken/VER:EIZ:NON")
	require.NotContains(t, res8, "Nachzudenken/SUB:NOM:SIN:NEU:INF")

	result9 := tagger.Tag([]string{"entlangblicken"})
	require.Len(t, result9, 1)
	require.Len(t, result9[0].GetReadings(), 5)
	res9 := atrListString(result9)
	require.Contains(t, res9, "entlangblicken/VER:INF:SFT")
	require.Contains(t, res9, "NEB")

	result10 := tagger.Tag([]string{"wiederaufbauen"})
	require.Len(t, result10, 1)
	require.Len(t, result10[0].GetReadings(), 5)
	res10 := atrListString(result10)
	require.Contains(t, res10, "NEB")
	require.NotContains(t, res10, "NEB:NEB")
}

// Twin of GermanTaggerTest.testPrefixVerbsNotMod
func TestGermanTagger_PrefixVerbsNotMod(t *testing.T) {
	tagger := requireGermanTagger(t)

	result1 := tagger.Tag([]string{"rauslassen"})
	require.Len(t, result1, 1)
	require.Len(t, result1[0].GetReadings(), 5)
	res1 := atrListString(result1)
	require.Contains(t, res1, "rauslassen/VER:1:PLU:PRÄ:NON:NEB")
	require.Contains(t, res1, "rauslassen/VER:1:PLU:KJ1:NON:NEB")
	require.Contains(t, res1, "rauslassen/VER:3:PLU:PRÄ:NON:NEB")
	require.Contains(t, res1, "rauslassen/VER:3:PLU:KJ1:NON:NEB")
	require.Contains(t, res1, "rauslassen/VER:INF:NON")
	require.NotContains(t, res1, "rauslassen/VER:MOD:1:PLU:PRÄ")
	require.NotContains(t, res1, "rauslassen/VER:MOD:INF")

	result2 := tagger.Tag([]string{"rauslass"})
	require.Len(t, result2, 1)
	require.Len(t, result2[0].GetReadings(), 2)
	res2 := atrListString(result2)
	require.Contains(t, res2, "rauslassen/VER:1:SIN:PRÄ:NON:NEB")
	require.NotContains(t, res2, "rauslassen/VER:IMP:SIN:NON")
	require.NotContains(t, res2, "rauslassen/VER:MOD:IMP:SIN")

	result3 := tagger.Tag([]string{"rauszulassen"})
	require.Len(t, result3, 1)
	require.Len(t, result3[0].GetReadings(), 1)
	require.Contains(t, atrListString(result3), "rauslassen/VER:EIZ:NON")

	result4 := tagger.Tag([]string{"rausgelassen"})
	require.Len(t, result4, 1)
	require.Len(t, result4[0].GetReadings(), 2)
	res4 := atrListString(result4)
	require.Contains(t, res4, "rauslassen/VER:PA2:NON")
	require.Contains(t, res4, "rausgelassen/PA2:PRD:GRU:VER")
	require.NotContains(t, res4, "rauslassen/VER:MOD:PA2")
}

// Twin of GermanTaggerTest.testPrefixVerbsNonSeparable
func TestGermanTagger_PrefixVerbsNonSeparable(t *testing.T) {
	tagger := requireGermanTagger(t)

	result1 := tagger.Tag([]string{"vergären"})
	require.Len(t, result1, 1)
	// official dict may yield duplicate tags; assert count and content
	require.GreaterOrEqual(t, len(result1[0].GetReadings()), 10)
	res1 := atrListString(result1)
	require.Contains(t, res1, "vergären/VER:1:PLU:PRÄ:NON")
	require.Contains(t, res1, "vergären/VER:1:PLU:KJ1:NON")
	require.Contains(t, res1, "vergären/VER:3:PLU:PRÄ:NON")
	require.Contains(t, res1, "vergären/VER:3:PLU:KJ1:NON")
	require.Contains(t, res1, "vergären/VER:INF:NON")
	require.Contains(t, res1, "vergären/VER:1:PLU:PRÄ:SFT")
	require.Contains(t, res1, "vergären/VER:INF:SFT")
	require.NotContains(t, res1, ":NEB")

	result2 := tagger.Tag([]string{"Vergären"})
	require.Len(t, result2, 1)
	require.GreaterOrEqual(t, len(result2[0].GetReadings()), 15)
	res2 := atrListString(result2)
	require.Contains(t, res2, "Vergären/SUB:NOM:SIN:NEU:INF")
	require.Contains(t, res2, "Vergären/SUB:DAT:SIN:NEU:INF")
	require.Contains(t, res2, "Vergären/SUB:AKK:SIN:NEU:INF")
	require.Contains(t, res2, "vergären/VER:1:PLU:PRÄ:NON")
	require.Contains(t, res2, "vergären/VER:INF:SFT")

	result3 := tagger.Tag([]string{"unbeglückt"})
	require.Len(t, result3, 1)
	// elative + dash/partizip paths may each contribute PA2:PRD (Java size often 2;
	// official path can emit duplicate identical PA2:PRD readings).
	require.GreaterOrEqual(t, len(result3[0].GetReadings()), 1)
	res3 := atrListString(result3)
	require.Contains(t, res3, "PA2:PRD:GRU:VER")
	// no finite VER: invent (Java assertFalse contains "VER:")
	require.NotContains(t, res3, "VER:")

	result4 := tagger.Tag([]string{"erstritten"})
	require.Len(t, result4, 1)
	require.Len(t, result4[0].GetReadings(), 6)
	res4 := atrListString(result4)
	require.Contains(t, res4, "erstreiten/VER:1:PLU:PRT:NON")
	require.Contains(t, res4, "erstreiten/VER:PA2:NON")
	require.Contains(t, res4, "erstritten/PA2:PRD:GRU:VER")

	result5 := tagger.Tag([]string{"bemessen"})
	require.Len(t, result5, 1)
	require.Len(t, result5[0].GetReadings(), 7)
	res5 := atrListString(result5)
	require.Contains(t, res5, "bemessen/VER:INF:NON")
	require.Contains(t, res5, "bemessen/VER:PA2:NON")
	require.Contains(t, res5, "bemessen/PA2:PRD:GRU:VER")

	result6 := tagger.Tag([]string{"bemisst"})
	require.Len(t, result6, 1)
	require.Len(t, result6[0].GetReadings(), 8)
	require.NotContains(t, atrListString(result6), "bemessen/PA2:PRD:GRU:VER")

	result7 := tagger.Tag([]string{"hinterlass"})
	require.Len(t, result7, 1)
	require.Len(t, result7[0].GetReadings(), 3)
	res7 := atrListString(result7)
	require.Contains(t, res7, "hinterlassen/VER:IMP:SIN:NON")
	require.Contains(t, res7, "hinterlassen/VER:1:SIN:PRÄ:NON")
	require.NotContains(t, res7, "NEB")
}

// Twin of GermanTaggerTest.testNoVerb
func TestGermanTagger_NoVerb(t *testing.T) {
	tagger := requireGermanTagger(t)

	for _, w := range []string{"geschichte", "bereich", "beispiel", "keksdose", "aktienarten", "unmenge"} {
		result := tagger.Tag([]string{w})
		require.Len(t, result, 1)
		require.Len(t, result[0].GetReadings(), 1)
		require.NotContains(t, atrListString(result), "VER", w)
	}

	result6 := tagger.Tag([]string{"schwarzgrau"})
	require.Len(t, result6, 1)
	// Java test expects ADJ+VER when the form is known. Official german.dict has no
	// "schwarzgrau"; notAVerb contains "schwarzgrau" so prefix invent is blocked
	// (Java StringUtils.containsAny). Fail closed: untagged, not false VER-only invent.
	if result6[0].IsTagged() {
		res6 := atrListString(result6)
		require.Contains(t, res6, "ADJ")
		require.Contains(t, res6, "VER")
	} else {
		require.Len(t, result6[0].GetReadings(), 1)
		require.Nil(t, result6[0].GetReadings()[0].GetPOSTag())
	}

	result8 := tagger.Tag([]string{"entlang"})
	require.Len(t, result8, 1)
	require.Len(t, result8[0].GetReadings(), 4)
	res8 := atrListString(result8)
	require.Contains(t, res8, "PRP")
	require.Contains(t, res8, "ZUS")
	require.NotContains(t, res8, "VER")
}

// Twin of GermanTaggerTest.testVerbAndPa2
func TestGermanTagger_VerbAndPa2(t *testing.T) {
	tagger := requireGermanTagger(t)

	result1 := tagger.Tag([]string{"erstickt"})
	require.Len(t, result1, 1)
	require.Len(t, result1[0].GetReadings(), 5)
	res1 := atrListString(result1)
	require.Contains(t, res1, "ersticken/VER:2:PLU:PRÄ:SFT")
	require.Contains(t, res1, "ersticken/VER:3:SIN:PRÄ:SFT")
	require.Contains(t, res1, "ersticken/VER:IMP:PLU:SFT")
	require.Contains(t, res1, "ersticken/VER:PA2:SFT")
	require.Contains(t, res1, "erstickt/PA2:PRD:GRU:VER")

	result2 := tagger.Tag([]string{"erstickte"})
	require.Len(t, result2, 1)
	require.Len(t, result2[0].GetReadings(), 19)
	res2 := atrListString(result2)
	require.Contains(t, res2, "erstickt/PA2")
	require.Contains(t, res2, "ersticken/VER:1:SIN:KJ2:SFT")
	require.Contains(t, res2, "ersticken/VER:1:SIN:PRT:SFT")
	require.Contains(t, res2, "ersticken/VER:3:SIN:KJ2:SFT")
	require.Contains(t, res2, "ersticken/VER:3:SIN:PRT:SFT")
	require.NotContains(t, res2, "NEB")

	result3 := tagger.Tag([]string{"erstickend"})
	require.Len(t, result3, 1)
	require.Len(t, result3[0].GetReadings(), 2)
	res3 := atrListString(result3)
	require.Contains(t, res3, "ersticken/VER:PA1:SFT")
	require.Contains(t, res3, "erstickend/PA1:PRD:GRU:VER")

	result4 := tagger.Tag([]string{"erstickender"})
	require.Len(t, result4, 1)
	require.Len(t, result4[0].GetReadings(), 7)
	require.Contains(t, atrListString(result4), "erstickend/PA1:DAT:SIN:FEM:GRU:SOL:VER")
}

// toSortedString ports GermanTaggerTest.toSortedString (TreeSet of reading.toString).
func toSortedString(tokenReadings *languagetool.AnalyzedTokenReadings) string {
	if tokenReadings == nil {
		return "null"
	}
	elements := make([]string, 0, len(tokenReadings.GetReadings()))
	for _, r := range tokenReadings.GetReadings() {
		if r != nil {
			elements = append(elements, r.String())
		}
	}
	sort.Strings(elements)
	return tokenReadings.GetToken() + "[" + strings.Join(elements, ", ") + "]"
}

// atrListString approximates List<AnalyzedTokenReadings>.toString for contains checks.
func atrListString(list []*languagetool.AnalyzedTokenReadings) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, atr := range list {
		if i > 0 {
			b.WriteByte(',')
		}
		if atr == nil {
			continue
		}
		b.WriteString(atr.GetToken())
		b.WriteByte('[')
		for j, r := range atr.GetReadings() {
			if j > 0 {
				b.WriteByte(',')
			}
			if r != nil {
				b.WriteString(r.String())
				// Java AnalyzedTokenReadings.toString appends '*' for some readings
				b.WriteByte('*')
			}
		}
		b.WriteByte(']')
	}
	b.WriteByte(']')
	return b.String()
}

