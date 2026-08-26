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
	// Java Speller.areEqual: only map[from].contains(to), not reverse
	opt := SuggestOpts{EquivalentChars: map[rune][]rune{'l': {'ł'}, 'u': {'ó'}}}
	if !runesEqualUnderOpts('l', 'ł', opt) {
		t.Fatal("l ~ ł")
	}
	if runesEqualUnderOpts('ł', 'l', opt) {
		t.Fatal("ł !~ l without SymmetricEquivalent (Java one-way)")
	}
	// invent edit-gen may enable reverse
	opt.SymmetricEquivalent = true
	if !runesEqualUnderOpts('ł', 'l', opt) {
		t.Fatal("ł ~ l with SymmetricEquivalent")
	}
}

func TestRunesEqualUnderOpts_DiacriticsConvertCase(t *testing.T) {
	// Java: IgnoreDiacritics + ConvertCase → E ~ é via NFD first + toLower
	opt := SuggestOpts{IgnoreDiacritics: true, ConvertCase: true}
	if !runesEqualUnderOpts('E', 'é', opt) {
		t.Fatal("E ~ é with ignore-diacritics+convert-case")
	}
	optNo := SuggestOpts{IgnoreDiacritics: true}
	if runesEqualUnderOpts('E', 'é', optNo) {
		t.Fatal("E !~ é without convert-case (NFD first is E vs e)")
	}
}


