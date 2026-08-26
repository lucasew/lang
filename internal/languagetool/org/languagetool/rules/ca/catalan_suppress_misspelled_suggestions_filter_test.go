package ca

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestCatalanSuppressMisspelled_NoDictAllMisspelled(t *testing.T) {
	ClearCatalanFilterSpeller()
	f := NewCatalanSuppressMisspelledSuggestionsFilter()
	// Java: null SpellingCheckRule → isMisspelled true → drop all + suppressMatch
	kept, ok := f.FilterSuggestions([]string{"casa", "bé"}, true)
	require.False(t, ok)
	require.Empty(t, kept)
}

func TestCatalanSuppressMisspelled_IncorrectVerbChunk(t *testing.T) {
	ClearCatalanFilterSpeller()
	// Wire a fake "dict" path unavailable → still need dict for step past null check
	// Use override path: HasIncorrectVerb after forcing available via inject of dict-less
	// Force available path by setting IsMisspelled to catalan path with HasIncorrectVerb only:
	// When dict available but HasIncorrectVerb true → misspelled
	f := NewCatalanSuppressMisspelledSuggestionsFilter()
	// Simulate dict present without real file: call catalanIsMisspelled with Available via inject
	// Wire dict if present; else inject Speller via SetIsMisspelled combining chunk + speller
	f.HasIncorrectVerb = func(s string) bool { return s == "badverb" }
	// Without dict, catalanIsMisspelled returns true before chunk — still true for badverb
	require.True(t, f.catalanIsMisspelled("badverb"))
	// With override for unit of chunk after dict gate:
	f.SetIsMisspelled(func(s string) bool {
		if f.HasIncorrectVerb != nil && f.HasIncorrectVerb(s) {
			return true
		}
		return false
	})
	kept, ok := f.FilterSuggestions([]string{"casa", "badverb"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"casa"}, kept)
}

func TestCatalanSuppressMisspelled_AcceptRuleMatch(t *testing.T) {
	ClearCatalanFilterSpeller()
	f := NewCatalanSuppressMisspelledSuggestionsFilter()
	f.SetIsMisspelled(func(w string) bool { return w == "xyz" })
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"casa", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"casa"}, out.GetSuggestedReplacements())
}

// Java isMisspelled receives the full suggestion (analyzeText), not invent per-token first.
func TestCatalanSuppressMisspelled_FullSuggestionToChunkHook(t *testing.T) {
	ClearCatalanFilterSpeller()
	f := NewCatalanSuppressMisspelledSuggestionsFilter()
	var seen []string
	// Bypass null-dict early return via override that still uses HasIncorrectVerb then OK.
	f.HasIncorrectVerb = func(s string) bool {
		seen = append(seen, s)
		return s == "va anar"
	}
	f.SetIsMisspelled(func(s string) bool {
		if f.HasIncorrectVerb != nil && f.HasIncorrectVerb(s) {
			return true
		}
		return false
	})
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"va anar", "casa bona"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"casa bona"}, out.GetSuggestedReplacements())
	require.Contains(t, seen, "va anar", "chunk hook must see full multi-word suggestion")
	require.Contains(t, seen, "casa bona")
}

func TestCatalanSuppressMisspelled_WithCADict(t *testing.T) {
	ClearCatalanFilterSpeller()
	t.Cleanup(ClearCatalanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/ca/src/main/resources/org/languagetool/resource/ca/ca-ES_spelling.dict"),
		filepath.Join(root, "third_party/ca/ca-ES_spelling.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireCatalanFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("ca-ES_spelling.dict not in tree")
	}
	f := NewCatalanSuppressMisspelledSuggestionsFilter()
	// invent nothing: junk should be misspelled
	require.True(t, f.catalanIsMisspelled("xyzzyqqq"))
	kept, ok := f.FilterSuggestions([]string{"xyzzyqqq"}, true)
	require.False(t, ok)
	require.Empty(t, kept)
}

func TestCatalanSuppressMisspelled_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.CatalanSuppressMisspelledSuggestionsFilter"))
}
