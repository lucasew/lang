package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Port of TokenPredicateTest.test — covered by unit tests below.
func TestTokenPredicate_Test(t *testing.T) {
	tok := NewChunkTaggedToken("foo", nil, nil)
	requireTrue(t, NewTokenPredicate("foo", true).Apply(tok))
	requireFalse(t, NewTokenPredicate("bar", true).Apply(tok))
	// single-quoted unquote
	requireTrue(t, NewTokenPredicate("string='foo'", true).Apply(tok))
	// pos contains
	pos := "NN"
	reading := languagetool.NewAnalyzedToken("foo", &pos, nil)
	atr := languagetool.NewAnalyzedTokenReadings(reading)
	ct := NewChunkTaggedToken("foo", nil, atr)
	requireTrue(t, NewTokenPredicate("pos=NN", true).Apply(ct))
	requireTrue(t, NewTokenPredicate("posre=N.*", true).Apply(ct))
	// bad multi-equals panics
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = NewTokenPredicate("a=b=c", true)
}

func requireTrue(t *testing.T, v bool) {
	t.Helper()
	if !v {
		t.Fatal("expected true")
	}
}
func requireFalse(t *testing.T, v bool) {
	t.Helper()
	if v {
		t.Fatal("expected false")
	}
}
