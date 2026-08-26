package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/RemoteRuleCacheTest.java
//
// Full JLanguageTool + RemoteRule RPC deferred (rules import would cycle).
// Green slice: duplicate sentence identity, local 0–1 matches, document offset
// shift, and ResultCache remote put/get hits — same shape as Java test.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of languagetool-core/src/test/java/org/languagetool/RemoteRuleCacheTest.java :: RemoteRuleCacheTest.testDuplicateSentence
func TestRemoteRuleCache_DuplicateSentence(t *testing.T) {
	text := "Foo. Foo. Bar." // repeated first sentence
	sentences := AnalyzeTextLocal(text)
	require.GreaterOrEqual(t, len(sentences), 2)

	// Distinct by GetText (Java stream.distinct on AnalyzedSentence ≈ same surface)
	seen := map[string]struct{}{}
	var distinct []*AnalyzedSentence
	for _, s := range sentences {
		k := s.GetText()
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		distinct = append(distinct, s)
	}
	require.Equal(t, 2, len(distinct), "Foo. / Bar. should collapse to 2 distinct surfaces")

	const ruleID = "TEST_REMOTE_RULE"
	// Stand-in for TestRemoteRule: one local match [0,1) per sentence.
	type localMatch struct{ From, To int }
	matchSentence := func(s *AnalyzedSentence) []localMatch {
		if s == nil || s.GetText() == "" {
			return nil
		}
		return []localMatch{{From: 0, To: 1}}
	}

	var directFrom []int
	offset := 0
	for _, s := range sentences {
		ms := matchSentence(s)
		require.Len(t, ms, 1)
		for _, m := range ms {
			directFrom = append(directFrom, m.From+offset)
		}
		offset += len([]rune(s.GetText()))
	}
	require.Equal(t, 3, len(directFrom), "Test rule matches when called directly")
	// Foo.␠ (5) + Foo.␠ (5) → sentence starts 0, 5, 10
	require.Equal(t, []int{0, 5, 10}, directFrom)

	// Remote match cache: second pass hits ResultCache for repeated "Foo."
	cache := NewResultCache(1000)
	var cachedFrom []int
	offset = 0
	for _, s := range sentences {
		st := s.GetText()
		var sentenceMatches []localMatch
		if v, ok := cache.GetRemoteMatchesIfPresent(st, ruleID); ok {
			sentenceMatches, _ = v.([]localMatch)
		} else {
			sentenceMatches = matchSentence(s)
			cache.PutRemoteMatches(st, ruleID, sentenceMatches)
		}
		for _, m := range sentenceMatches {
			cachedFrom = append(cachedFrom, m.From+offset)
		}
		offset += len([]rune(st))
	}
	require.Equal(t, 3, len(cachedFrom), "Cached matches collected correctly")
	require.Equal(t, []int{0, 5, 10}, cachedFrom)
	require.GreaterOrEqual(t, cache.HitCount(), int64(1), "duplicate sentence should hit remote cache")
}
