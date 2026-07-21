package de

// Twin of GermanTaggerTest — MapWordTagger / expansion twins of Java control flow.
// Full german.dict corpus cases need the official binary (not vendored here).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestGermanTagger_AdjectivesFromSpellingTxt(t *testing.T) {
	// Java: spelling.txt /A /P expansions via adjInfos; skip fünftjüngste.
	adj := &SpellingAdjExpansion{byForm: map[string][]tagging.TaggedWord{
		"meistgewünscht":  {tagging.NewTaggedWord("meistgewünscht", "ADJ:PRD:GRU")},
		"meistgewünschtes": {tagging.NewTaggedWord("meistgewünscht", "ADJ:NOM:SIN:NEU:GRU:IND")},
		"abgemindert":     {tagging.NewTaggedWord("abgemindert", "PA2:PRD:GRU:VER")},
	}}
	tagger := NewGermanTagger(tagging.MapWordTagger{})
	tagger.SetSpellingAdjExpansion(adj)
	require.NotNil(t, tagger.Lookup("meistgewünscht"))
	require.NotNil(t, tagger.Lookup("abgemindert"))
	// unknown comparative-style form: no invent
	require.Nil(t, tagger.Lookup("fünftjüngste"))
}

func TestGermanTagger_LemmaOfForDashCompounds(t *testing.T) {
	// Java: sanitizeWord last dash segment; lemma rebuild with stem.
	wt := tagging.MapWordTagger{
		"Verband": {tagging.NewTaggedWord("Verband", "SUB:NOM:SIN:MAS")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Lookup("Zahn-Arzt-Verband")
	require.NotNil(t, got)
	lemmas := make([]string, 0)
	for _, r := range got.GetReadings() {
		if r.GetLemma() != nil {
			lemmas = append(lemmas, *r.GetLemma())
		}
	}
	// stem "Zahn-Arzt-" + lemma "Verband" via addStem path
	require.Contains(t, lemmas, "Zahn-Arzt-Verband")
}

func TestGermanTagger_GenderGap(t *testing.T) {
	// Java testGenderGap: Freund * innen → :PLU:FEM; jede * r → PRO FEM+MAS
	wt := tagging.MapWordTagger{
		"Freund":     {tagging.NewTaggedWord("Freund", "SUB:NOM:SIN:MAS")},
		"Freundin":   {tagging.NewTaggedWord("Freundin", "SUB:NOM:SIN:FEM")},
		"Freundinnen": {
			tagging.NewTaggedWord("Freundin", "SUB:NOM:PLU:FEM"),
			tagging.NewTaggedWord("Freundin", "SUB:AKK:PLU:FEM"),
		},
		"jede":  {tagging.NewTaggedWord("jede", "PRO:IND:NOM:SIN:FEM")},
		"jeder": {tagging.NewTaggedWord("jeder", "PRO:IND:NOM:SIN:MAS")},
		"Mitarbeiter":   {tagging.NewTaggedWord("Mitarbeiter", "SUB:NOM:SIN:MAS")},
		"Mitarbeiterin": {tagging.NewTaggedWord("Mitarbeiterin", "SUB:NOM:SIN:FEM")},
	}
	tagger := NewGermanTagger(wt)
	for _, gap := range []string{"*", "_", ":", "/"} {
		got := tagger.Tag([]string{"viele", "Freund", gap, "innen"})
		tags := posTagsOf(got[1])
		joined := strings.Join(tags, " ")
		require.Contains(t, joined, ":PLU:FEM", "gap %q", gap)
	}
	got := tagger.Tag([]string{"jede", "*", "r", "Mitarbeiter", "*", "in"})
	j0 := strings.Join(posTagsOf(got[0]), " ")
	require.Contains(t, j0, "PRO:IND:NOM:SIN:FEM")
	require.Contains(t, j0, "PRO:IND:NOM:SIN:MAS")
	j3 := strings.Join(posTagsOf(got[3]), " ")
	require.Contains(t, j3, "SUB:NOM:SIN:FEM")
	require.Contains(t, j3, "SUB:NOM:SIN:MAS")
}

func TestGermanTagger_IgnoreDomain(t *testing.T) {
	// Java: bundestag . de multi compound skip domain → untagged
	wt := tagging.MapWordTagger{
		"de": {tagging.NewTaggedWord("de", "UNKNOWN:DE")},
	}
	tagger := NewGermanTagger(wt)
	tagger.SplitCompound = func(word string) []string {
		if word == "bundestag" {
			return []string{"bundes", "tag"}
		}
		return []string{word}
	}
	got := tagger.Tag([]string{"bundestag", ".", "de"})
	require.Nil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_IgnoreImperative(t *testing.T) {
	// Java: zehnfach must not get false IMP via +e
	wt := tagging.MapWordTagger{
		"zehnfache": {tagging.NewTaggedWord("zehnfach", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"zehnfach"})
	// single-token → atStart requires size>1; no allowed prev → untagged
	require.Nil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Haus": {
			tagging.NewTaggedWord("Haus", "SUB:AKK:SIN:NEU"),
			tagging.NewTaggedWord("Haus", "SUB:DAT:SIN:NEU"),
			tagging.NewTaggedWord("Haus", "SUB:NOM:SIN:NEU"),
		},
		"Hauses": {tagging.NewTaggedWord("Haus", "SUB:GEN:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	require.NotNil(t, tagger.Lookup("Haus"))
	require.NotNil(t, tagger.Lookup("Hauses"))
	// Lookup uses ignoreCase=false
	require.Nil(t, tagger.Lookup("hauses"))
	require.Nil(t, tagger.Lookup("Groß"))
}

func TestGermanTagger_ExtendedTagger(t *testing.T) {
	// both pre- and post-spelling-reform forms present in dict twin
	wt := tagging.MapWordTagger{
		"Kuß":  {tagging.NewTaggedWord("Kuß", "SUB:NOM:SIN:MAS")},
		"Kuss": {tagging.NewTaggedWord("Kuss", "SUB:NOM:SIN:MAS")},
		"Haß":  {tagging.NewTaggedWord("Haß", "SUB:NOM:SIN:MAS")},
		"Hass": {tagging.NewTaggedWord("Hass", "SUB:NOM:SIN:MAS")},
	}
	tagger := NewGermanTagger(wt)
	require.NotNil(t, tagger.Lookup("Kuß"))
	require.NotNil(t, tagger.Lookup("Kuss"))
	require.NotNil(t, tagger.Lookup("Haß"))
	require.NotNil(t, tagger.Lookup("Hass"))
}

func TestGermanTagger_AfterColon(t *testing.T) {
	// Java: after ":" empty firstWord branch lowercases "Als"
	wt := tagging.MapWordTagger{
		"als": {tagging.NewTaggedWord("als", "KOUS"), tagging.NewTaggedWord("als", "KOKOM"),
			tagging.NewTaggedWord("als", "ADV"), tagging.NewTaggedWord("als", "APPR")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Er", "sagte", ":", "Als", "Erstes", "würde", "ich"})
	require.Len(t, got, 7)
	require.Equal(t, "Als", got[3].GetToken())
	require.GreaterOrEqual(t, len(got[3].GetReadings()), 4)
	require.NotNil(t, got[3].GetReadings()[0].GetPOSTag())
}

func TestGermanTagger_TaggerBaseforms(t *testing.T) {
	wt := tagging.MapWordTagger{
		"übrigbleibst": {tagging.NewTaggedWord("übrigbleiben", "VER:2:SIN:PRÄ:NON:NEB")},
		"Haus": {
			tagging.NewTaggedWord("Haus", "SUB:AKK:SIN:NEU"),
			tagging.NewTaggedWord("Haus", "SUB:DAT:SIN:NEU"),
			tagging.NewTaggedWord("Haus", "SUB:NOM:SIN:NEU"),
		},
		"Häuser": {
			tagging.NewTaggedWord("Haus", "SUB:AKK:PLU:NEU"),
			tagging.NewTaggedWord("Haus", "SUB:GEN:PLU:NEU"),
			tagging.NewTaggedWord("Haus", "SUB:NOM:PLU:NEU"),
		},
	}
	tagger := NewGermanTagger(wt)
	r1 := tagger.Lookup("übrigbleibst")
	require.NotNil(t, r1)
	require.Equal(t, "übrigbleiben", *r1.GetReadings()[0].GetLemma())
	r2 := tagger.Lookup("Haus")
	require.Len(t, r2.GetReadings(), 3)
	r3 := tagger.Lookup("Häuser")
	require.Len(t, r3.GetReadings(), 3)
	for _, r := range r3.GetReadings() {
		require.Equal(t, "Haus", *r.GetLemma())
	}
}

func TestGermanTagger_Tag(t *testing.T) {
	wt := tagging.MapWordTagger{
		"das": {tagging.NewTaggedWord("der", "ART:DEF:NOM:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	// ignoreCase=false: "Das" not in dict → null POS
	got := tagger.TagIgnoreCase([]string{"Das"}, false)
	require.Nil(t, got[0].GetReadings()[0].GetPOSTag())
	// ignoreCase=true: sentence-start lowercase merge
	got2 := tagger.TagIgnoreCase([]string{"Das"}, true)
	require.NotNil(t, got2[0].GetReadings()[0].GetPOSTag())
	require.Contains(t, *got2[0].GetReadings()[0].GetPOSTag(), "ART:")
}

func TestGermanTagger_TagWithManualDictExtension(t *testing.T) {
	// words from added.txt / german.dict extension
	wt := tagging.MapWordTagger{
		"Wichtigtuerinnen": {
			tagging.NewTaggedWord("Wichtigtuerin", "SUB:AKK:PLU:FEM"),
			tagging.NewTaggedWord("Wichtigtuerin", "SUB:DAT:PLU:FEM"),
			tagging.NewTaggedWord("Wichtigtuerin", "SUB:GEN:PLU:FEM"),
			tagging.NewTaggedWord("Wichtigtuerin", "SUB:NOM:PLU:FEM"),
		},
	}
	got := NewGermanTagger(wt).Tag([]string{"Wichtigtuerinnen"})
	require.Len(t, got[0].GetReadings(), 4)
	require.Equal(t, "Wichtigtuerin", *got[0].GetReadings()[0].GetLemma())
}

func TestGermanTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{"Tisch": {tagging.NewTaggedWord("Tisch", "SUB:NOM:SIN:MAS")}}
	tagger := NewGermanTagger(wt)
	require.Equal(t, GermanDictPath, tagger.GetDictionaryPath())
	require.Len(t, tagger.TagWord("Tisch"), 1)
}

func TestGermanTagger_IsWeiseException(t *testing.T) {
	// Java testIsWeiseException — stem must be ADJ in word tagger only
	wt := tagging.MapWordTagger{
		"lustig":  {tagging.NewTaggedWord("lustig", "ADJ:PRD:GRU")},
		"ideal":   {tagging.NewTaggedWord("ideal", "ADJ:PRD:GRU")},
		"überw":   {tagging.NewTaggedWord("überweisen", "VER:IMP:SIN:SFT")},
		"verweis": {tagging.NewTaggedWord("verweisen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
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

func TestGermanTagger_PrefixVerbsFromSpellingTxt(t *testing.T) {
	// Minimal verbInfos twin: herausfallen → VER with :NEB / INF
	wt := tagging.MapWordTagger{
		"fallen": {
			tagging.NewTaggedWord("fallen", "VER:1:PLU:PRÄ:NON"),
			tagging.NewTaggedWord("fallen", "VER:3:PLU:PRÄ:NON"),
			tagging.NewTaggedWord("fallen", "VER:INF:NON"),
		},
	}
	tagger := NewGermanTagger(wt)
	ex := &SpellingVerbExpansion{VerbInfos: map[string]PrefixInfixVerb{
		"herausfallen": {Prefix: "heraus", Infix: "", VerbBaseform: "fallen"},
	}}
	tagger.SetSpellingVerbExpansion(ex)
	got := tagger.Tag([]string{"herausfallen"})
	joined := readingsString(got[0])
	require.Contains(t, joined, "VER:")
	require.Contains(t, joined, "NEB")
	require.NotContains(t, joined, "ADJ:")
}

func TestGermanTagger_PrefixVerbsSeparable(t *testing.T) {
	wt := tagging.MapWordTagger{
		"guckst": {tagging.NewTaggedWord("gucken", "VER:2:SIN:PRÄ:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"nachguckst"})
	joined := readingsString(got[0])
	require.Contains(t, joined, "VER:2:SIN:PRÄ:SFT:NEB")
	require.NotContains(t, joined, "ADJ:")
}

func TestGermanTagger_PrefixVerbsNotMod(t *testing.T) {
	// MOD readings must not be copied from base (lassen)
	wt := tagging.MapWordTagger{
		"lassen": {
			tagging.NewTaggedWord("lassen", "VER:INF:NON"),
			tagging.NewTaggedWord("lassen", "VER:MOD:INF"),
			tagging.NewTaggedWord("lassen", "VER:1:PLU:PRÄ:NON"),
		},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"rauslassen"})
	joined := readingsString(got[0])
	require.Contains(t, joined, "VER:INF:NON")
	require.NotContains(t, joined, "VER:MOD")
}

func TestGermanTagger_PrefixVerbsNonSeparable(t *testing.T) {
	wt := tagging.MapWordTagger{
		"gären": {
			tagging.NewTaggedWord("gären", "VER:INF:NON"),
			tagging.NewTaggedWord("gären", "VER:INF:SFT"),
			tagging.NewTaggedWord("gären", "VER:1:PLU:PRÄ:NON"),
		},
	}
	tagger := NewGermanTagger(wt)
	// non-sep prefix "ver" — via dash/prefix unknown path or expansion
	ex := &SpellingVerbExpansion{VerbInfos: map[string]PrefixInfixVerb{
		"vergären": {Prefix: "ver", Infix: "", VerbBaseform: "gären"},
	}}
	tagger.SetSpellingVerbExpansion(ex)
	got := tagger.Tag([]string{"vergären"})
	joined := readingsString(got[0])
	require.Contains(t, joined, "VER:")
	require.NotContains(t, joined, ":NEB")
}

func TestGermanTagger_NoVerb(t *testing.T) {
	// notAVerb substrings block false verb prefix tagging
	wt := tagging.MapWordTagger{
		"schichte": {tagging.NewTaggedWord("schichten", "VER:1:SIN:PRÄ:SFT")},
		"eich":     {tagging.NewTaggedWord("eichen", "VER:IMP:SIN:SFT")},
		"spiel":    {tagging.NewTaggedWord("spielen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	// "geschichte" contains "schichte"? notAVerb has "geschichte" etc.
	// Java: notAVerb blocks; surface unknown → no VER invent from prefix path
	for _, w := range []string{"geschichte", "bereich", "beispiel"} {
		got := tagger.Tag([]string{w})
		joined := readingsString(got[0])
		require.NotContains(t, joined, "VER", w)
	}
}

func TestGermanTagger_VerbAndPa2(t *testing.T) {
	// PA2 dual reading for separable prefix past participle path
	wt := tagging.MapWordTagger{
		"geschickt": {tagging.NewTaggedWord("schicken", "VER:PA2:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"zurückgeschickt"})
	joined := readingsString(got[0])
	// may get VER:PA2 and/or PA2:PRD depending on branch
	require.True(t, strings.Contains(joined, "PA2") || strings.Contains(joined, "VER:PA"), joined)
}

func readingsString(rd *languagetool.AnalyzedTokenReadings) string {
	if rd == nil {
		return ""
	}
	var b strings.Builder
	for _, r := range rd.GetReadings() {
		if r == nil {
			continue
		}
		b.WriteString(r.GetToken())
		b.WriteByte('[')
		if r.GetLemma() != nil {
			b.WriteString(*r.GetLemma())
		}
		b.WriteByte('/')
		if r.GetPOSTag() != nil {
			b.WriteString(*r.GetPOSTag())
		}
		b.WriteString("] ")
	}
	return b.String()
}
