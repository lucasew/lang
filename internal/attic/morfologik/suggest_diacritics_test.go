package morfologik

import (
	"testing"
)

func TestStripDiacritic(t *testing.T) {
	if stripDiacritic('é') != 'e' {
		t.Fatalf("é → %c", stripDiacritic('é'))
	}
	if stripDiacritic('ñ') != 'n' {
		t.Fatalf("ñ → %c", stripDiacritic('ñ'))
	}
	if stripDiacritic('a') != 'a' {
		t.Fatalf("a → %c", stripDiacritic('a'))
	}
}

func TestRunesEqualUnderOpts_Diacritics(t *testing.T) {
	opt := SuggestOpts{IgnoreDiacritics: true}
	if !runesEqualUnderOpts('e', 'é', opt) {
		t.Fatal("e ~ é with ignore-diacritics")
	}
	if runesEqualUnderOpts('e', 'é', SuggestOpts{}) {
		t.Fatal("e !~ é without flag")
	}
}

func TestRunesEqualUnderOpts_EquivalentChars(t *testing.T) {
	opt := SuggestOpts{EquivalentChars: map[rune][]rune{'l': {'ł'}, 'u': {'ó'}}}
	if !runesEqualUnderOpts('l', 'ł', opt) {
		t.Fatal("l ~ ł")
	}
	if !runesEqualUnderOpts('ł', 'l', opt) {
		t.Fatal("ł ~ l reverse")
	}
}

func TestEdit1Candidates_DiacriticAlphabet(t *testing.T) {
	cands := edit1CandidatesOpts("cafe", SuggestOpts{IgnoreDiacritics: true})
	found := false
	for _, c := range cands {
		if c == "café" || c == "cafè" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected café-style replace among %d cands", len(cands))
	}
}
