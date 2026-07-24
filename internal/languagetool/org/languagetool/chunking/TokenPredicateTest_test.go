package chunking

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/chunking/TokenPredicateTest.java

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of TokenPredicateTest.test — Java-visible match / no-match outcomes.
func TestTokenPredicate_Test(t *testing.T) {
	chunkTags := []ChunkTag{NewChunkTag("CHUNK1"), NewChunkTag("CHUNK2")}
	pos := "MYPOS"
	readings := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("mytoken", &pos, strPtr("mylemma")))
	chunkTaggedToken := NewChunkTaggedToken("mytoken", chunkTags, readings)

	assertMatch := func(expr string) {
		t.Helper()
		p := NewTokenPredicate(expr, false)
		require.True(t, p.Apply(chunkTaggedToken), "expected match for %q", expr)
	}
	assertNoMatch := func(expr string) {
		t.Helper()
		p := NewTokenPredicate(expr, false)
		require.False(t, p.Apply(chunkTaggedToken), "expected no match for %q", expr)
	}

	assertMatch("mytoken")
	assertNoMatch("mytoken2")
	assertMatch("string=mytoken")
	assertNoMatch("string=mytoken2")
	assertMatch("regex=my[abct]oken")
	assertNoMatch("regex=my[abc]oken")
	assertMatch("chunk=CHUNK1")
	assertMatch("chunk=CHUNK2")
	assertNoMatch("chunk=OTHERCHUNK")
	assertMatch("pos=MYPOS")
	assertNoMatch("pos=OTHER")
	assertMatch("posre=M.POS")
	assertNoMatch("posre=O.HER")

	// invalid=token → RuntimeException in Java; panic in Go
	require.Panics(t, func() {
		_ = NewTokenPredicate("invalid=token", false).Apply(chunkTaggedToken)
	})
}

func strPtr(s string) *string { return &s }
