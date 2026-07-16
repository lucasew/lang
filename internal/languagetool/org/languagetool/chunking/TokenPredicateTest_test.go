package chunking

import "testing"

// Port of TokenPredicateTest.test — covered by TestTokenPredicate unit tests.
func TestTokenPredicate_Test(t *testing.T) {
	// simple string match
	tok := NewChunkTaggedToken("foo", nil, nil)
	requireTrue(t, NewTokenPredicate("foo", true).Apply(tok))
	requireFalse(t, NewTokenPredicate("bar", true).Apply(tok))
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
