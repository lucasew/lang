package morfologik

import "testing"

// Twins of morfologik SpellerTest edit-distance helpers (no FSA dict required).

// Port of SpellerTest edit-distance assertions (abka family) with threshold 2.
func TestSpellerED_EditDistance(t *testing.T) {
	spell := NewSpellerED(2)
	// Java: getEditDistance(spell, "abka", "abakan") == 2
	if got := spell.GetEditDistance("abka", "abakan"); got != 2 {
		t.Fatalf("abka/abakan: got %d want 2", got)
	}
	if got := spell.GetEditDistance("abka", "abaki"); got != 2 {
		t.Fatalf("abka/abaki: got %d want 2", got)
	}
	// same word
	if got := spell.GetEditDistance("abka", "abka"); got != 0 {
		t.Fatalf("abka/abka: got %d want 0", got)
	}
	// single substitution
	if got := spell.GetEditDistance("abka", "abke"); got != 1 {
		t.Fatalf("abka/abke: got %d want 1", got)
	}
}

// Port of SpellerTest.testCutOffEditDistance (Oflazer repo/reprter).
func TestSpellerED_CutOffEditDistance(t *testing.T) {
	spell2 := NewSpellerED(2)
	// Java: getCutOffDistance(spell2, "repo", "reprter") == 1
	if got := spell2.GetCutOffDistance("repo", "reprter"); got != 1 {
		t.Fatalf("repo/reprter cuted: got %d want 1", got)
	}
	// Java: getCutOffDistance(spell2, "reporter", "reporter") == 0
	if got := spell2.GetCutOffDistance("reporter", "reporter"); got != 0 {
		t.Fatalf("reporter/reporter cuted: got %d want 0", got)
	}
}

func TestSpellerED_DiacriticAreEqual(t *testing.T) {
	spell := NewSpellerED(2)
	spell.IgnoreDiacritics = true
	// with ignore-diacritics, e ~ é so abke ~ abké is free at that position
	// full strings abka/abką: last char a vs ą — free → distance from other diffs only
	d := spell.GetEditDistance("abka", "abką")
	// Java assertTrue(getEditDistance(spell, "abka", "abaką") == 2) with Polish dict flags
	// Here only last-char free: abka vs abką same length, one free match → dist 0
	if d != 0 {
		t.Fatalf("abka/abką with ignore-diacritics: got %d want 0", d)
	}
}
