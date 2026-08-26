package br

// Twin of languagetool-language-modules/br/src/test/java/org/languagetool/rules/br/TopoReplaceRuleTest.java
//
// Java:
//   TopoReplaceRule rule = new TopoReplaceRule(TestTools.getEnglishMessages());
//   JLanguageTool lt = new JLanguageTool(new Breton());
//   assertEquals(1, rule.match(lt.getAnalyzedSentence("France a zo ur vro.")).length);
//   assertEquals(0, rule.match(lt.getAnalyzedSentence("France 3 a zo ur chadenn skinwel.")).length);
//
// France 3 → 0 requires Breton XmlRuleDisambiguator: disambiguation.xml rulegroup FRANCE
// immunizes "France" when followed by [23]|Bleu|Culture|…; TopoReplaceRule skips immunized starts.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagbr "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/br"
	"github.com/stretchr/testify/require"
)

// wireBretonLT ports new JLanguageTool(new Breton()) for analysis:
// BretonWordTokenizer (via language code) + XmlRuleDisambiguator(Breton).
// Tagger is optional for TopoReplaceRule (surface match + immunization only).
func wireBretonLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	d := tagbr.BretonXmlRuleDisambiguator()
	require.NotNil(t, d, "Breton XmlRuleDisambiguator (disambiguation.xml) required for France 3 immunization")
	lt := languagetool.NewJLanguageTool("br")
	lt.Disambiguator = d
	return lt
}

func TestTopoReplaceRule_Rule(t *testing.T) {
	// Java TopoReplaceRuleTest.testRule
	rule := NewTopoReplaceRule(nil)
	lt := wireBretonLT(t)

	// Wrong sentences.
	matches := rule.Match(lt.GetAnalyzedSentence("France a zo ur vro."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Frañs", "bro-C’hall"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "anvioù lec’h", matches[0].GetShortMessage())
	require.Contains(t, matches[0].Message, "France")
	require.Contains(t, matches[0].Message, "<suggestion>Frañs</suggestion>")
	require.Contains(t, matches[0].Message, " pe ")
	require.Contains(t, matches[0].Message, "<suggestion>bro-C’hall</suggestion>")

	// Correct sentences — "France 3" TV channel: FRANCE disambig immunizes "France".
	matches = rule.Match(lt.GetAnalyzedSentence("France 3 a zo ur chadenn skinwel."))
	require.Equal(t, 0, len(matches))
}

func TestTopoReplaceRule_CaseSensitive(t *testing.T) {
	// Java isCaseSensitive()==true: lowercase key not in map.
	rule := NewTopoReplaceRule(nil)
	lt := wireBretonLT(t)
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("france a zo ur vro."))))
}

func TestTopoReplaceRule_Metadata(t *testing.T) {
	rule := NewTopoReplaceRule(nil)
	require.Equal(t, "BR_TOPO", rule.GetID())
	require.Equal(t, BretonTopo, rule.GetID())
	require.Equal(t, "anvioù-lec’h e brezhoneg", rule.GetDescription())
	require.Equal(t, "anvioù lec’h", rule.GetShort())
	require.Equal(t, " zo un anv lec’h gallek. Ha fellout a rae deoc’h skrivañ ", rule.GetSuggestion())
	require.Equal(t, " pe ", rule.GetSuggestionsSeparator())
	require.True(t, rule.IsCaseSensitive())
	require.Equal(t, "br", rule.GetLocale())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.CategoryMisc))

	ww := rule.GetWrongWords()
	require.NotEmpty(t, ww)
	// France is a single-word key in map index 0
	require.Equal(t, "Frañs|bro-C’hall", ww[0]["France"])
}

func TestTopoReplaceRule_TopoTxtOfficial(t *testing.T) {
	// Byte-identity with official rules/br/topo.txt (Java getFromRulesDirAsStream).
	official := discoverOfficialTopoTxt(t)
	got, err := topoFS.ReadFile("data/topo.txt")
	require.NoError(t, err)
	want, err := os.ReadFile(official)
	require.NoError(t, err)
	require.Equal(t, want, got, "embedded topo.txt must byte-match official resource")
}

func discoverOfficialTopoTxt(t *testing.T) string {
	t.Helper()
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "br",
		"src", "main", "resources", "org", "languagetool", "rules", "br", "topo.txt")
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("official topo.txt not found")
	return ""
}
