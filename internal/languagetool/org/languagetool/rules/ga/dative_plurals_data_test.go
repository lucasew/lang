package ga

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/stretchr/testify/require"
)

func TestDativePluralsEntry(t *testing.T) {
	e := NewDativePluralsEntry("fearaibh", "fear", "m", "fir")
	require.Equal(t, "fir", e.GetStandard())
	e.SetEquivalent("fir")
	require.True(t, e.HasEquivalent())
	require.Equal(t, "Noun:Masc:Dat:Pl", e.GetBaseTag())
	e2 := NewDativePluralsEntry("mnáibh", "bean", "f", "mná")
	require.Equal(t, "Noun:Fem:Dat:Pl", e2.GetBaseTag())
	// getStandard: equivalent preferred over replacement
	e3 := NewDativePluralsEntry("áiteachaibh", "áit", "f", "áiteacha")
	e3.SetEquivalent("áiteanna")
	require.Equal(t, "áiteanna", e3.GetStandard())
}

func TestParseDativePluralsData(t *testing.T) {
	in := strings.NewReader("fearaibh;fear;m;fir\nbliantaibh:blianta;bliain;f;blianta\n")
	d, err := ParseDativePluralsData(in)
	require.NoError(t, err)
	require.Len(t, d.Entries, 2)
	require.Equal(t, "fir", d.GetSimpleReplacements()["fearaibh"])
	// mutations on base form
	require.Equal(t, taggingga.Lenite("fir"), d.GetSimpleReplacements()[taggingga.Lenite("fearaibh")])
	require.Equal(t, taggingga.Eclipse("fir"), d.GetSimpleReplacements()[taggingga.Eclipse("fearaibh")])
	// modernisations map + mutation values use getReplacement() for lenite/eclipse
	require.Equal(t, "blianta", d.GetModernisations()["bliantaibh"])
	require.Equal(t, taggingga.Lenite("blianta"), d.GetModernisations()[taggingga.Lenite("bliantaibh")])
}

func TestParseDativePluralsData_WrongFieldCount(t *testing.T) {
	// Java: throw RuntimeException("Incorrect number of fields: " + line)
	_, err := ParseDativePluralsData(strings.NewReader("a;b;c\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "Incorrect number of fields:")
}

func TestParseDativePluralsData_TooManyColons(t *testing.T) {
	// Java: throw on form with more than 1 modern form
	_, err := ParseDativePluralsData(strings.NewReader("a:b:c;lemma;m;repl\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "Form has more than 1 modern form:")
}

func TestLoadDativePluralsData_OfficialMutations(t *testing.T) {
	d := LoadDativePluralsData()
	require.NotNil(t, d)
	simple := d.GetSimpleReplacements()
	require.NotEmpty(t, simple)
	require.Equal(t, DativePluralsFilePath, "/ga/dative-plurals.txt")

	// Official base keys from dative-plurals.txt
	require.Equal(t, "mná", simple["mnáibh"])
	require.Equal(t, "áiteanna", simple["áiteachaibh"]) // equivalent preferred
	require.Equal(t, "ba", simple["búaibh"])
	require.Equal(t, "ba", simple["buaibh"]) // modernised form → standard

	// Lenition / eclipse / h-prothesis keys (assert real keys, not smoke).
	// Java Utils.eclipse: no case for 'm' → eclipse("mnáibh") == "mnáibh" (no "mmnáibh" key).
	require.Equal(t, taggingga.Lenite("mná"), simple[taggingga.Lenite("mnáibh")])
	require.Equal(t, "mhná", simple["mhnáibh"])
	require.Equal(t, taggingga.Lenite("mnáibh"), "mhnáibh")
	// b-initial: eclipse prefixes "m" (Java case 'b')
	require.Equal(t, "ba", simple["búaibh"])
	require.Equal(t, taggingga.Eclipse("ba"), simple[taggingga.Eclipse("búaibh")])
	require.Equal(t, "mba", simple["mbúaibh"])
	require.Equal(t, "mba", simple["mbuaibh"]) // modernised form also mutated

	// Vowel-initial form: h-prothesis + eclipse (n- for lowercase vowel)
	require.Equal(t, "háiteanna", simple["háiteachaibh"])
	require.Equal(t, "háiteanna", simple["h-áiteachaibh"])
	require.Equal(t, taggingga.Eclipse("áiteanna"), simple[taggingga.Eclipse("áiteachaibh")])
	require.Equal(t, "n-áiteanna", simple["n-áiteachaibh"])

	// Modernisations map for official modernised entry (búaibh:buaibh)
	mod := d.GetModernisations()
	require.Equal(t, "buaibh", mod["búaibh"])
	// lenite/eclipse values use getReplacement() ("ba"), not getModern
	require.Equal(t, taggingga.Lenite("ba"), mod[taggingga.Lenite("búaibh")])
	require.Equal(t, "bha", mod["bhúaibh"])
	require.Equal(t, taggingga.Eclipse("ba"), mod[taggingga.Eclipse("búaibh")])
	require.Equal(t, "mba", mod["mbúaibh"])
}

func TestDativePlurals_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := dativePluralsFS.ReadFile("data/dative-plurals.txt")
	require.NoError(t, err)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// this file: .../internal/languagetool/org/languagetool/rules/ga/*.go
	// seven Dir() steps reach repo root (ga→rules→languagetool→org→languagetool→internal→repo)
	repoRoot := thisFile
	for i := 0; i < 7; i++ {
		repoRoot = filepath.Dir(repoRoot)
	}
	official := filepath.Join(repoRoot,
		"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/dative-plurals.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/dative-plurals.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official ga/dative-plurals.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded ga/data/dative-plurals.txt must be byte-identical to official Java resource")
}

func TestDhaNoBeirtData(t *testing.T) {
	var d DhaNoBeirtData
	nums := d.GetNumberReplacements()
	require.Equal(t, "beirt", nums["dhá"])
	require.NotNil(t, d.GetDaoine())
}

func TestIrishPartialPosTagFilter(t *testing.T) {
	f := NewIrishPartialPosTagFilter(func(s string) []string { return []string{"Noun"} })
	require.NotNil(t, f)
	ClearDefaultIrishPartialPosTagger()
	f2 := NewNoDisambiguationIrishPartialPosTagFilter(nil)
	require.NotNil(t, f2)
	// fail-closed without process-wide tagger
	ok, err := f2.Accept("x", "^(x)$", "Noun", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestIrishPartialPosTagFilter_WithTagAndDisambig(t *testing.T) {
	SetDefaultIrishPartialPosTagger(func(p string) []string {
		if p == "cat" {
			return []string{"Noun"}
		}
		return nil
	})
	WireIrishFilterDisambiguator(stubGADisambig{})
	t.Cleanup(func() {
		ClearDefaultIrishPartialPosTagger()
		ClearIrishFilterDisambiguator()
	})
	f := NewIrishPartialPosTagFilter(nil)
	ok, err := f.Accept("cats", "^(cat)s$", "Noun", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

type stubGADisambig struct{}

func (stubGADisambig) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return s
}
