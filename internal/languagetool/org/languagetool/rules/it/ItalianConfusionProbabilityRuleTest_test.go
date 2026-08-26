package it

// Twin of ArabicConfusionProbabilityRuleTest matrix for Italian empty subclass.
// Java has no dedicated ItalianConfusionProbabilityRuleTest; Java-visible meta from
// ItalianConfusionProbabilityRule + ConfusionProbabilityRule:
//   ctor super(messages, lm, language) only; example pair commented TODO;
//   no getId / getFilenames override; official /it/confusion_sets.txt header-only.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

// TestItalianConfusionProbabilityRule_Constructor mirrors AR constructor/meta matrix
// for this empty Italian subclass (FakeLanguageModel stand-in + id/filenames/examples/pairs).
func TestItalianConfusionProbabilityRule_Constructor(t *testing.T) {
	// Java: new ItalianConfusionProbabilityRule(messages, languageModel, language);
	// super only — constructing with FakeLanguageModel stand-in does not panic.
	lm := ngrams.UniformLanguageModel(0.5, 1) // FakeLanguageModel stand-in
	r := NewItalianConfusionProbabilityRule(lm)
	require.NotNil(t, r)
	require.NotNil(t, r.ConfusionProbabilityRule)
	require.Equal(t, ngrams.ConfusionRuleID, r.GetID())
	require.Equal(t, "CONFUSION_RULE", r.GetID())

	// Java-visible constructor meta (not invent): no override of getFilenames
	require.Equal(t, []string{"confusion_sets.txt"}, r.GetFilenames())
	require.Equal(t, ItalianConfusionFilenames(), r.GetFilenames())

	// Java ItalianConfusionProbabilityRule example pair is commented TODO only.
	require.Empty(t, r.GetIncorrectExamples(), "Java has no incorrect examples")
	require.Empty(t, r.GetCorrectExamples(), "Java has no correct examples")
	require.Len(t, r.GetIncorrectExamples(), 0)
	require.Len(t, r.GetCorrectExamples(), 0)

	// Official /it/confusion_sets.txt is header comment only → empty WordToPairs after parse.
	// Do not invent confusion data rows.
	require.Empty(t, r.WordToPairs, "header-only confusion_sets.txt must yield empty WordToPairs")
	require.Len(t, r.WordToPairs, 0)
}

// TestItalianConfusionSets_MatchOfficialResource asserts the embedded confusion_sets.txt
// is byte-identical to the official Java resource Java loads for Italian.
func TestItalianConfusionSets_MatchOfficialResource(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// rules/it → walk up to repo root via known layout: .../rules/it → 5 levels to workspace
	// thisFile is .../internal/languagetool/org/languagetool/rules/it/ItalianConfusionProbabilityRuleTest_test.go
	itDir := filepath.Dir(thisFile)
	embeddedPath := filepath.Join(itDir, "data", "confusion_sets.txt")
	embedded, err := os.ReadFile(embeddedPath)
	require.NoError(t, err)

	// inspiration path relative to package: 5× .. from rules/it → workspace root
	repoRoot := filepath.Clean(filepath.Join(itDir, "..", "..", "..", "..", "..", ".."))
	officialPath := filepath.Join(repoRoot,
		"inspiration", "languagetool", "languagetool-language-modules", "it",
		"src", "main", "resources", "org", "languagetool", "resource", "it", "confusion_sets.txt")
	official, err := os.ReadFile(officialPath)
	if err != nil {
		t.Skipf("official Italian confusion_sets.txt not available: %v", err)
	}
	require.Equal(t, official, embedded, "embedded it/data/confusion_sets.txt must match official resource bytes")
}
