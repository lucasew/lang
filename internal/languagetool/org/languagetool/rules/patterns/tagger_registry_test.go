package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of MatchState.toFinalString: lemma==null && hasNoTag on first reading only.
func TestIsUnknownToTagger_JavaLemmaAndHasNoTag(t *testing.T) {
	// empty result → unknown
	RegisterLanguageTagger("xx-unk", func(token string) []languagetool.TokenTag {
		return nil
	})
	require.True(t, IsUnknownToTagger("xx-unk", "zzz"))

	// POS set, lemma empty → known (hasNoTag false)
	RegisterLanguageTagger("xx-pos", func(token string) []languagetool.TokenTag {
		return []languagetool.TokenTag{{POS: "NN", Lemma: ""}}
	})
	require.False(t, IsUnknownToTagger("xx-pos", "word"))

	// POS empty, lemma set → known (Java: lemma != null short-circuits)
	RegisterLanguageTagger("xx-lem", func(token string) []languagetool.TokenTag {
		return []languagetool.TokenTag{{POS: "", Lemma: "dog"}}
	})
	require.False(t, IsUnknownToTagger("xx-lem", "dogs"), "lemma set → not MISTAKE even without POS")

	// POS empty, lemma empty → unknown
	RegisterLanguageTagger("xx-both", func(token string) []languagetool.TokenTag {
		return []languagetool.TokenTag{{POS: "", Lemma: ""}}
	})
	require.True(t, IsUnknownToTagger("xx-both", "q"))

	// SENT_END only → hasNoTag true; lemma empty → unknown
	RegisterLanguageTagger("xx-end", func(token string) []languagetool.TokenTag {
		return []languagetool.TokenTag{{POS: languagetool.SentenceEndTagName, Lemma: ""}}
	})
	require.True(t, IsUnknownToTagger("xx-end", "."))

	// Empty word → unknown (before tagger)
	require.True(t, IsUnknownToTagger("xx-start", ""))

	// SENT_START only → hasNoTag false in Java → known (not MISTAKE)
	RegisterLanguageTagger("xx-start", func(token string) []languagetool.TokenTag {
		return []languagetool.TokenTag{{POS: languagetool.SentenceStartTagName, Lemma: ""}}
	})
	require.False(t, IsUnknownToTagger("xx-start", "x"), "SENT_START is not hasNoTag")

	// First reading untagged, second tagged → Java uses only first → unknown
	RegisterLanguageTagger("xx-first", func(token string) []languagetool.TokenTag {
		return []languagetool.TokenTag{
			{POS: "", Lemma: ""},
			{POS: "NN", Lemma: "x"},
		}
	})
	require.True(t, IsUnknownToTagger("xx-first", "x"), "only first reading consulted")

	// No tagger registered → false (do not invent misspell).
	// Use a code with no shared base prefix (RegisterLanguageTagger also maps base before '-').
	require.False(t, IsUnknownToTagger("zznone", "anything"))
}
