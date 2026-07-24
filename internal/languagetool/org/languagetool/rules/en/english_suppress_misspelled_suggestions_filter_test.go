package en

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestEnglishSuppressMisspelledSuggestionsFilter_NoDictKeepsAll(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	// no dict: FilterDictIsMisspelled returns false (Java null SpellingCheckRule)
	kept, ok := f.FilterSuggestions([]string{"house", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"house", "xyz"}, kept)
}

func TestEnglishSuppressMisspelledSuggestionsFilter_Injected(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"house", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"house"}, kept)
	// all misspelled + suppressMatch → drop match
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
	// all misspelled + suppressMatch false → keep match with empty suggestions
	kept, ok = f.FilterSuggestions([]string{"xyz"}, false)
	require.True(t, ok)
	require.Empty(t, kept)
}

func TestEnglishSuppressMisspelledSuggestionsFilter_AcceptRuleMatch(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"house", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"house"}, out.GetSuggestedReplacements())

	// all misspelled + suppressMatch → drop
	m2 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m2.SetSuggestedReplacements([]string{"xyz"})
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"suppressMatch": "true"}, 0, nil, nil))

	// all misspelled + suppressMatch false → keep match with empty suggestions
	m3 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m3.SetSuggestedReplacements([]string{"xyz"})
	out = f.AcceptRuleMatch(m3, map[string]string{"suppressMatch": "false"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
}

// Java abstract isMisspelled tokenizes via language.getWordTokenizer() (EN: EnglishWordTokenizer)
// then checks each token with the default spelling rule.
func TestEnglishSuppressMisspelledSuggestionsFilter_TokenizesSuggestion(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	// per-token: "xyz" misspelled; "house" / "!" not
	f.IsMisspelled = func(w string) bool { return w == "xyz" || w == "teh" }
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"house!", "teh house", "xyz", "good house"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	// "house!" → tokens house + ! → keep
	// "teh house" has "teh" → drop
	// "xyz" → drop
	// "good house" → keep
	require.Equal(t, []string{"house!", "good house"}, out.GetSuggestedReplacements())
}

func TestEnglishSuppressMisspelledSuggestionsFilter_WithENDict(t *testing.T) {
	ClearEnglishFilterSpeller()
	t.Cleanup(ClearEnglishFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	dict := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	if !WireEnglishFilterSpeller(dict) {
		t.Skipf("en_US.dict not openable: %s", dict)
	}
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"house", "xyzzyqqq"}, true)
	require.True(t, ok)
	require.Contains(t, kept, "house")
	require.NotContains(t, kept, "xyzzyqqq")
	// multi-token via EnglishWordTokenizer: "house!" → house + ! ("!" is in en_US.dict)
	kept, ok = f.FilterSuggestions([]string{"house!", "xyzzyqqq!"}, true)
	require.True(t, ok)
	require.Contains(t, kept, "house!")
	require.NotContains(t, kept, "xyzzyqqq!")
	// space-separated: Java keeps " " as a token; en_US ignore_punctuation=false →
	// Speller.isMisspelled(" ") is true → whole suggestion dropped (bug-for-bug).
	_, ok = f.FilterSuggestions([]string{"good house"}, true)
	require.False(t, ok)
	// all misspelled + suppress → drop match
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"xyzzyqqq"})
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil))
}

func TestEnglishSuppressMisspelledSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.EnglishSuppressMisspelledSuggestionsFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.EnglishSuppressMisspelledSuggestionsFilter")
	require.NotNil(t, f)
	_, ok := f.(*EnglishSuppressMisspelledSuggestionsFilter)
	require.True(t, ok)
}
