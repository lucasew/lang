package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestImperativeForm_SentenceStart(t *testing.T) {
	// "gehe" is IMP; short "Geh" at sentence start should get IMP via +e
	wt := tagging.MapWordTagger{
		"gehe": {tagging.NewTaggedWord("gehen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Geh", "bitte", "!"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.True(t, stringsHasPrefix(*got[0].GetReadings()[0].GetPOSTag(), "VER:IMP:SIN"))
}

func TestImperativeForm_NotInMiddle(t *testing.T) {
	wt := tagging.MapWordTagger{
		"gehe": {tagging.NewTaggedWord("gehen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	// single-token sentence: atStart requires size > 1
	got := tagger.Tag([]string{"Geh"})
	// no imperative expand (only one token)
	require.Nil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestSubstantivatedForms(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Verletzte": {tagging.NewTaggedWord("Verletzte", "SUB:NOM:SIN:FEM:ADJ")},
	}
	tagger := NewGermanTagger(wt)
	// "Verletzter" followed by lowercase "kam"
	got := tagger.Tag([]string{"Ein", "Verletzter", "kam"})
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
	require.Equal(t, "SUB:NOM:SIN:MAS:ADJ", *got[1].GetReadings()[0].GetPOSTag())
}

func TestSubstantivated_BlockedByNextUpper(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Verletzte": {tagging.NewTaggedWord("Verletzte", "SUB:NOM:SIN:FEM:ADJ")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Verletzter", "Arzt"})
	// next is uppercase → not substantivated
	require.Nil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestDDDEr_2019er(t *testing.T) {
	tagger := NewGermanTagger(tagging.MapWordTagger{})
	got := tagger.Tag([]string{"2019er", "Wert"})
	require.NotEmpty(t, got[0].GetReadings())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.True(t, stringsHasPrefix(*got[0].GetReadings()[0].GetPOSTag(), "ADJ:"))
}

func TestMitarbeitenden(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Mitarbeitende": {tagging.NewTaggedWord("mitarbeitende", "SUB:NOM:SIN:MAS")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Neumitarbeitende"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestGenderGap_JedeR(t *testing.T) {
	wt := tagging.MapWordTagger{
		"jede":  {tagging.NewTaggedWord("jede", "PRO:IND:NOM:SIN:FEM")},
		"jeder": {tagging.NewTaggedWord("jeder", "PRO:IND:NOM:SIN:MAS")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"jede", "*", "r"})
	// first token should merge tags of "jede" and "jeder"
	require.GreaterOrEqual(t, len(got[0].GetReadings()), 1)
	// at least one reading present
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

// Java: idx-1 == "„" && idx-3 == ":" → lowercase dict readings for speech start.
// Layout: [filler, ":", filler, "„", "Das", ...]
func TestDirectSpeech_LowercaseAfterColonQuote(t *testing.T) {
	wt := tagging.MapWordTagger{
		"das": {tagging.NewTaggedWord("das", "ART:DEF:NOM:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Er", ":", "sagte", "„", "Das", "ist", "gut"})
	require.NotNil(t, got[4].GetReadings()[0].GetPOSTag(), "Das after : … „ must get lowercase 'das' tags")
	require.Equal(t, "ART:DEF:NOM:SIN:NEU", *got[4].GetReadings()[0].GetPOSTag())
}

// Mid-sentence capitalized unknown without :/„ pattern must not invent lowercase tags.
func TestDirectSpeech_NoFalseLowercaseMidSentence(t *testing.T) {
	wt := tagging.MapWordTagger{
		"das": {tagging.NewTaggedWord("das", "ART:DEF:NOM:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Und", "Das", "fehlt"})
	// "Das" mid-sentence: Java leaves untagged when only lowercase "das" is in dict
	require.Nil(t, got[1].GetReadings()[0].GetPOSTag())
}

// Java getImperativeForm: skip if removalTagger.tag(w).contains(tagged)
func TestImperativeForm_RespectsRemovalTagger(t *testing.T) {
	// "gehe" yields IMP; short "Geh" would expand unless removed for "geh"
	imp := tagging.NewTaggedWord("gehen", "VER:IMP:SIN:SFT")
	wt := tagging.MapWordTagger{
		"gehe": {imp},
	}
	// removal list: form "geh" → same lemma/POS as would be derived from gehe
	// Java checks removalTagger.tag(w) with w=lowercased "geh" after sentence start
	removal := tagging.MapWordTagger{
		"geh": {imp},
	}
	comb := tagging.NewCombiningTaggerWithRemoval(wt, tagging.MapWordTagger{}, removal, false)
	tagger := NewGermanTagger(comb)
	require.NotNil(t, tagger.RemovalTagger)
	got := tagger.Tag([]string{"Geh", "bitte", "!"})
	// Without removal, IMP is assigned; with removal, null POS (unknown after failed arms)
	require.Nil(t, got[0].GetReadings()[0].GetPOSTag(), "removed imperative short form must not be re-tagged")
}

// Control: same dict without removal still tags "Geh"
func TestImperativeForm_WithoutRemovalStillTags(t *testing.T) {
	wt := tagging.MapWordTagger{
		"gehe": {tagging.NewTaggedWord("gehen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Geh", "bitte", "!"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.True(t, stringsHasPrefix(*got[0].GetReadings()[0].GetPOSTag(), "VER:IMP:SIN"))
}

// Java Lookup: tag(singleton, ignoreCase=false); null POS → null return.
func TestLookup_IgnoreCaseFalseAndNullOnUnknown(t *testing.T) {
	wt := tagging.MapWordTagger{
		"das":  {tagging.NewTaggedWord("das", "ART:DEF:NOM:SIN:NEU")},
		"Haus": {tagging.NewTaggedWord("Haus", "SUB:NOM:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	// Exact known surface
	require.NotNil(t, tagger.Lookup("Haus"))
	// Capitalized form only in lower dict: Lookup must not invent case fold (ignoreCase=false)
	require.Nil(t, tagger.Lookup("Das"))
	// Fully unknown
	require.Nil(t, tagger.Lookup("xyzzy"))
	// Sentence Tag (ignoreCase=true) at start still lowercases first word
	sent := tagger.Tag([]string{"Das", "Haus"})
	require.NotNil(t, sent[0].GetReadings()[0].GetPOSTag())
}

func stringsHasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

// Twin: Werkstudent : innen-Zielgruppe → tags of Zielgruppe
// Java: innenPattern1 = "in(nen)-[A-ZÖÄÜ][a-zöäüß-]+" (nen required)
func TestGenderGap_InnenPattern1(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Werkstudent": {tagging.NewTaggedWord("Werkstudent", "SUB:NOM:SIN:MAS")},
		"Zielgruppe":  {tagging.NewTaggedWord("Zielgruppe", "SUB:NOM:SIN:FEM")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Werkstudent", ":", "innen-Zielgruppe"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "SUB:NOM:SIN:FEM")
	// pattern replaces taggerTokens entirely with lastPart tags
	require.NotContains(t, tags, "SUB:NOM:SIN:MAS")
}

// Twin: Java (nen) is required — "in-Zielgruppe" must not take Pattern1 path
func TestGenderGap_InnenPattern1_RequiresNen(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Werkstudent": {tagging.NewTaggedWord("Werkstudent", "SUB:NOM:SIN:MAS")},
		"Zielgruppe":  {tagging.NewTaggedWord("Zielgruppe", "SUB:NOM:SIN:FEM")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Werkstudent", ":", "in-Zielgruppe"})
	tags := posTagsOf(got[0])
	// falls back to word alone
	require.Contains(t, tags, "SUB:NOM:SIN:MAS")
	require.NotContains(t, tags, "SUB:NOM:SIN:FEM")
}

// Twin: Werkstudent : innenzielgruppe → UppercaseFirst after last "innen"
func TestGenderGap_InnenPattern2(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Zielgruppe": {tagging.NewTaggedWord("Zielgruppe", "SUB:NOM:SIN:FEM")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Werkstudent", ":", "innenzielgruppe"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "SUB:NOM:SIN:FEM")
}

// Twin: after colon empty firstWord branch lowercases; firstWord only flips there.
func TestFirstWord_AfterColonLowercase(t *testing.T) {
	wt := tagging.MapWordTagger{
		"das": {tagging.NewTaggedWord("das", "ART:DEF:NOM:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	// after ":", "Das" empty → lowercase "das"
	got := tagger.Tag([]string{"Wort", ":", "Das"})
	tags := posTagsOf(got[2])
	require.Contains(t, tags, "ART:DEF:NOM:SIN:NEU")
}
