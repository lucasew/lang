package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Wire-path smoke kept separate; full outcome matrix is in
// english_partial_pos_tag_filter_test.go.

func TestEnglishPartialPosTagFilter_TagAndDisambiguate(t *testing.T) {
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)

	// Fail-closed without both hooks.
	f := NewEnglishPartialPosTagFilter(nil)
	ok, err := f.Accept("running", "^(run).*", "VB.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	// Wire tagger-like function + pass-through disambiguator.
	// Accept extracts group 1 ("run") from "running" via ^(run).*, then tags that partial.
	tw := func(token string) []languagetool.TokenTag {
		if token == "run" {
			return []languagetool.TokenTag{{POS: "VB", Lemma: "run"}, {POS: "NN", Lemma: "run"}}
		}
		return nil
	}
	filterTagMu.Lock()
	filterTagWord = tw
	filterTagMu.Unlock()
	WireEnglishFilterDisambiguator(passThroughDisambig{})

	// Match any of the POS tags.
	ok, err = f.Accept("running", "^(run).*", "VB.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)

	// Fail when postag does not match.
	ok, err = f.Accept("running", "^(run).*", "JJ.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	// Only tagger, no disambig → fail closed.
	ClearEnglishFilterDisambiguator()
	ok, err = f.Accept("running", "^(run).*", "VB.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}
