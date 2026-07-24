package en

// Twin of EnglishNumberInWordFilterTest.
import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestEnglishNumberInWordFilter_FailClosedWithoutDict(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishNumberInWordFilter()
	// Soft invent removed: without speller dict, no candidates
	require.Empty(t, f.Suggestions("H0use"))
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "H0use"}, 0, nil, nil))
}

func TestEnglishNumberInWordFilter_WithInjectedSpeller(t *testing.T) {
	// Inject misspelled gate without full dict binary
	f := NewEnglishNumberInWordFilter()
	f.inner.IsMisspelled = func(w string) bool {
		// known good forms
		return w != "House" && w != "house" && w != "Good"
	}
	f.inner.GetSuggestions = nil
	// Force available path by using Suggestions via inner only (Suggestions also checks FilterDictAvailable)
	// Call inner.Suggestions directly for unit gate logic
	got := f.inner.Suggestions("H0use")
	require.Contains(t, got, "House")
}

// Twin of EnglishNumberInWordFilterTest.testFilter: Go0d → Good
func TestEnglishNumberInWordFilter_Filter(t *testing.T) {
	// Prefer real EN dict when present; else inject isMisspelled gate (same abstract filter path).
	ClearEnglishFilterSpeller()
	t.Cleanup(ClearEnglishFilterSpeller)
	f := NewEnglishNumberInWordFilter()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Try common en_US.dict locations under language module
	dictCandidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en/hunspell/en_US.dict"),
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en/en_US.dict"),
	}
	wired := false
	for _, d := range dictCandidates {
		if WireEnglishFilterSpeller(d) {
			wired = true
			break
		}
	}
	if !wired {
		// Abstract filter twin without full binary: 0→o gives Good if not misspelled
		f.inner.IsMisspelled = func(w string) bool { return w != "Good" && w != "good" }
		// Suggestions() may fail-closed without dict; AcceptRuleMatch uses same gate
		// Temporarily force Accept via inner (public Accept may require dict availability)
		m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 4, "fake msg")
		out := f.inner.AcceptRuleMatch(m, map[string]string{"word": "Go0d"}, 2, nil, nil)
		require.NotNil(t, out)
		require.Contains(t, out.GetSuggestedReplacements(), "Good")
		return
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 4, "fake msg")
	out := f.AcceptRuleMatch(m, map[string]string{"word": "Go0d"}, 2, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "Good")
}
