package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

// Twin: GermanTagger verbInfo branch — separable finite gets :NEB when
// indexOf==0 or first-char-lower (Java substring UTF-16 gate).
func TestVerbInfo_SeparableFiniteNEB(t *testing.T) {
	wt := tagging.MapWordTagger{
		"gebe": {tagging.NewTaggedWord("geben", "VER:1:SIN:PRÄ:NON")},
	}
	tagger := NewGermanTagger(wt)
	vex, err := LoadSpellingVerbExpansion(strings.NewReader("herum_geben\n"))
	require.NoError(t, err)
	tagger.SetSpellingVerbExpansion(vex)

	// conjugated surface registered via synth-less map: only prefix+base and zu in VerbInfos.
	// Register conjugated form like Java synthesizer would:
	vex.VerbInfos["herumgebe"] = PrefixInfixVerb{Prefix: "herum", Infix: "", VerbBaseform: "geben"}

	got := tagger.Tag([]string{"herumgebe", "es"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "VER:1:SIN:PRÄ:NON:NEB")
}

// Title case mid-sentence: indexOf!=0 and not first-char-lower → no :NEB, plain VER:1.
func TestVerbInfo_TitleMidSentenceNoNEB(t *testing.T) {
	wt := tagging.MapWordTagger{
		"gebe": {tagging.NewTaggedWord("geben", "VER:1:SIN:PRÄ:NON")},
	}
	tagger := NewGermanTagger(wt)
	vex, err := LoadSpellingVerbExpansion(strings.NewReader("herum_geben\n"))
	require.NoError(t, err)
	tagger.SetSpellingVerbExpansion(vex)
	vex.VerbInfos["Herumgebe"] = PrefixInfixVerb{Prefix: "herum", Infix: "", VerbBaseform: "geben"}

	// not at start; Title case fails first-char-lower
	got := tagger.Tag([]string{"Dann", "Herumgebe", "ich"})
	// Herumgebe is unknown to dict; expansion strips prefix using UTF-16 length of "herum"
	// noPrefixForm from "Herumgebe" with cut=len("herum")=5 → "gebe"?
	// Java substring(prefix.length()) on "Herumgebe" with prefix "herum" (len 5):
	// H-e-r-u-m-g-e-b-e — wait surface is Herumgebe, prefix length is of verbInfo.prefix "herum"=5,
	// substring(5) of "Herumgebe" = "gebe" (H=0,e=1,r=2,u=3,m=4,g=5 → gebe). Good.
	tags := posTagsOf(got[1])
	require.NotContains(t, tags, "VER:1:SIN:PRÄ:NON:NEB")
	require.Contains(t, tags, "VER:1:SIN:PRÄ:NON")
}

// zu-form: VER:EIZ:SFT when base tags contain :SFT (Java isSFT).
func TestVerbInfo_ZuEIZ_SFT(t *testing.T) {
	wt := tagging.MapWordTagger{
		"geben": {tagging.NewTaggedWord("geben", "VER:INF:SFT")},
	}
	tagger := NewGermanTagger(wt)
	vex, err := LoadSpellingVerbExpansion(strings.NewReader("herum_geben\n"))
	require.NoError(t, err)
	tagger.SetSpellingVerbExpansion(vex)

	rd := tagger.Lookup("herumzugeben")
	require.NotNil(t, rd)
	require.Equal(t, "VER:EIZ:SFT", *rd.GetReadings()[0].GetPOSTag())
	require.Equal(t, "herumgeben", *rd.GetReadings()[0].GetLemma())
}

// non-separable verbInfo: Java "VER:IMP:SIN"+flektion bug → VER:IMP:SINSFT
func TestVerbInfo_NonSepIMP_MissingColonTwin(t *testing.T) {
	wt := tagging.MapWordTagger{
		"zeih": {tagging.NewTaggedWord("zeihen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	vex, err := LoadSpellingVerbExpansion(strings.NewReader("ver_zeihen\n"))
	require.NoError(t, err)
	tagger.SetSpellingVerbExpansion(vex)
	// surface verzeih = prefix ver + form zeih
	vex.VerbInfos["verzeih"] = PrefixInfixVerb{Prefix: "ver", Infix: "", VerbBaseform: "zeihen"}

	got := tagger.Tag([]string{"verzeih"})
	tags := posTagsOf(got[0])
	// bug-for-bug: missing colon between SIN and SFT
	require.Contains(t, tags, "VER:IMP:SINSFT")
	require.Contains(t, tags, "VER:1:SIN:PRÄ:SFT")
}

// nounTagExpansionExceptions: Wegstrecken must not get SUB from expansion.
func TestNounTagExpansionException_Wegstrecken(t *testing.T) {
	wt := tagging.MapWordTagger{}
	tagger := NewGermanTagger(wt)
	vex, err := LoadSpellingVerbExpansion(strings.NewReader("weg_strecken\n"))
	require.NoError(t, err)
	tagger.SetSpellingVerbExpansion(vex)

	rd := tagger.Lookup("Wegstrecken")
	// exception → no SUB readings from nom; untagged / null POS
	if rd != nil {
		for _, r := range rd.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				require.NotContains(t, *r.GetPOSTag(), "SUB:", "Wegstrecken must not get noun expansion tags")
			}
		}
	}
}

// Dash/prefix path: Title mid-sentence must not force :NEB (Java lower|index==0 only).
func TestPrefixPath_TitleMidSentenceNoNEB(t *testing.T) {
	wt := tagging.MapWordTagger{
		"lädst": {tagging.NewTaggedWord("laden", "VER:2:SIN:PRÄ:NON")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Dann", "Einlädst", "du"})
	tags := posTagsOf(got[1])
	require.NotContains(t, tags, "VER:2:SIN:PRÄ:NON:NEB")
}

func TestPrefixPath_LowerGetsNEB(t *testing.T) {
	wt := tagging.MapWordTagger{
		"lädst": {tagging.NewTaggedWord("laden", "VER:2:SIN:PRÄ:NON")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Dann", "einlädst", "du"})
	tags := posTagsOf(got[1])
	require.Contains(t, tags, "VER:2:SIN:PRÄ:NON:NEB")
}
