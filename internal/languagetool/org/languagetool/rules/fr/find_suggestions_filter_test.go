package fr

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrFR(s string) *string { return &s }

func TestFRFindSuggestionsRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.fr.FindSuggestionsFilter"))
}

func TestFRCleanSuggestion(t *testing.T) {
	require.Equal(t, "manger", frCleanSuggestion("le manger"))
	require.Equal(t, "manger", frCleanSuggestion("s'manger"))
	require.Equal(t, "faire", frCleanSuggestion("nous faire quelque chose"))
}

func TestFRGetSpellingSuggestions_Variants(t *testing.T) {
	f := NewFindSuggestionsFilter()
	f.SpellingMatch = func(word string) []string {
		// return word-specific
		switch word {
		case "mänge":
			return []string{"mange"}
		case "manges":
			return []string{"mange"}
		case "mange":
			return []string{"mange"}
		default:
			return nil
		}
	}
	// tagged token → MakeWrong first
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("mange", ptrFR("V"), ptrFR("manger")))
	// IsTagged should be true with POS
	sugs := f.getSpellingSuggestions(tok)
	require.Contains(t, sugs, "mange")
}

func TestFRFindSuggestionsAccept(t *testing.T) {
	f := NewFindSuggestionsFilter()
	f.SpellingMatch = func(word string) []string {
		if word == "maisön" || word == "maison" {
			return []string{"maison"}
		}
		return nil
	}
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFR("N f s"), ptrFR("maison")))
	}
	// untagged misspelling
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("maisön", nil, nil))
	tok.SetStartPos(0)
	f.EnsureSpellingHook()
	m := rules.NewRuleMatch(nil, nil, 0, 6, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "maison")
}

func TestFRFindSuggestionsNoSpeller(t *testing.T) {
	f := NewFindSuggestionsFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 1, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, nil, nil))
}

func TestFRFindSuggestions_WithFRDict(t *testing.T) {
	ClearFrenchFilterSpeller()
	t.Cleanup(ClearFrenchFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Java MorfologikFrenchSpellerRule: /fr/french.dict
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/fr/src/main/resources/org/languagetool/resource/fr/french.dict"),
		filepath.Join(root, "third_party/fr/french.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireFrenchFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("french.dict not in tree (Java FindSuggestionsFilter default spelling rule)")
	}
	f := NewFindSuggestionsFilter()
	// Soft POS: accept any suggestion without FrenchTagger
	f.MatchesDesiredPostag = func(suggestion, desiredPostag string) bool { return true }
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("maisön", nil, nil))
	tok.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("FS"), nil, 0, 6, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	// Dict may suggest maison; do not invent if empty
	if len(FilterDictSuggest("maisön")) > 0 {
		require.NotNil(t, out)
		require.NotEmpty(t, out.GetSuggestedReplacements())
	}
}
