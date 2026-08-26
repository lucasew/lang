package languagetool

// Twin of ProtoResultMatchCacheTest — ResultCache holds opaque match payloads
// (proto-free stand-in for CachedResultMatch lists).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtoResultMatchCache_RunTests(t *testing.T) {
	c := NewResultCache(5)
	sent := AnalyzePlain("test")
	key := NewInputSentence(sent, "en", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil)
	type cachedMatch struct {
		RuleID  string
		Message string
		FromPos int
		ToPos   int
	}
	m := cachedMatch{RuleID: "RULE", Message: "msg", FromPos: 1, ToPos: 4}
	c.PutMatches(key, []cachedMatch{m})
	got, ok := c.GetMatchesIfPresent(key)
	require.True(t, ok)
	list, ok := got.([]cachedMatch)
	require.True(t, ok)
	require.Equal(t, "RULE", list[0].RuleID)
	require.Equal(t, "msg", list[0].Message)
}
