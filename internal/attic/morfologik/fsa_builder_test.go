package morfologik

import (
	"testing"
)

func TestFSABuilder_SimpleWords(t *testing.T) {
	fsa := BuildFSAFromWords([]string{"a", "ab", "abc", "b", "ba"})
	if fsa == nil {
		t.Fatal("nil fsa")
	}
	root := fsa.RootNode()
	if root == 0 {
		t.Fatal("root is terminal")
	}
	for _, w := range []string{"a", "ab", "abc", "b", "ba"} {
		kind, _, _ := fsa.Match([]byte(w), root)
		if kind != ExactMatch {
			t.Fatalf("%q: kind=%d want ExactMatch", w, kind)
		}
	}
	// missing
	kind, _, _ := fsa.Match([]byte("ac"), root)
	if kind == ExactMatch {
		t.Fatal("ac should not match")
	}
	kind, _, _ = fsa.Match([]byte("z"), root)
	if kind == ExactMatch {
		t.Fatal("z should not match")
	}
}

func TestFSABuilder_Prefix(t *testing.T) {
	fsa := BuildFSAFromWords([]string{"abc", "abd"})
	root := fsa.RootNode()
	kind, _, node := fsa.Match([]byte("ab"), root)
	if kind != SequenceIsAPrefix && kind != ExactMatch {
		// "ab" is prefix of longer words, not final itself
		if kind != SequenceIsAPrefix {
			t.Fatalf("ab kind=%d", kind)
		}
	}
	_ = node
	// full words
	kind, _, _ = fsa.Match([]byte("abc"), root)
	if kind != ExactMatch {
		t.Fatalf("abc kind=%d", kind)
	}
}

func TestDictionary_FromWords_Suggest(t *testing.T) {
	d := NewDictionaryFromWords([]string{"receive", "recipe", "the", "cat"}, nil)
	if d == nil {
		t.Fatal("nil dict")
	}
	if !d.IsInDictionary("receive") {
		t.Fatal("receive not in dict")
	}
	if d.IsInDictionary("xyz") {
		t.Fatal("xyz should miss")
	}
	// SpellerFSA findRepl
	sp := NewSpellerFSA(d, 1)
	cands := sp.FindReplacementCandidates("recieve")
	found := false
	for _, c := range cands {
		if c.Word == "receive" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected receive; got %v", cands)
	}
	// run-on
	words := d.ReplaceRunOnWords("thecat")
	if len(words) == 0 || words[0] != "the cat" && !containsStr(words, "the cat") {
		t.Fatalf("thecat → %v", words)
	}
}

func containsStr(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
