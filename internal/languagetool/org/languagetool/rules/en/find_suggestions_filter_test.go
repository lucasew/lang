package en

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFindSuggestionsFilter_WiresTagAndSynthesize(t *testing.T) {
	ClearEnglishFilterTagger()
	f := NewFindSuggestionsFilter()
	require.NotNil(t, f.Tag)
	require.NotNil(t, f.Synthesize)
	// Without WireEnglishFilterTagger, Tag returns nil (fail-closed)
	require.Nil(t, f.Tag("house"))
}

func TestFilterTagWord_WithDict(t *testing.T) {
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	pos := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/english.dict")
	if !WireEnglishFilterTagger(pos) {
		t.Skip("english.dict not openable")
	}
	atr := FilterTagWord("houses")
	require.NotNil(t, atr)
	// dict should yield NNS/NN-style readings
	require.True(t, atr.IsTagged() || atr.MatchesPosTagRegex("NN.*"),
		"expected tagged houses, got readings=%v", atr.GetReadings())
	require.True(t, FilterSuggestionMatchesPostag("houses", "NNS|NN"))
}

func TestFindSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.FindSuggestionsFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.FindSuggestionsFilter")
	require.NotNil(t, f)
	// New instance each GetFilter
	fs, ok := f.(*FindSuggestionsFilter)
	if ok {
		require.NotNil(t, fs.Tag)
		require.NotNil(t, fs.Synthesize)
	}
}

func TestFindSuggestionsFilter_NoDictFailClosed(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewFindSuggestionsFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 4, "msg")
	// no speller → Accept returns nil
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "NN.*",
	}, 0, nil, nil))
}
