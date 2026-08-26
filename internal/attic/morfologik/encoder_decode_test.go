package morfologik

import "testing"

// Twins of morfologik Trim*Encoder.decode examples / control bytes.

func TestDecodeTrimSuffix(t *testing.T) {
	// truncate 0 ('A'), append "xy" to "abc" → abcxy
	got := decodeTrimSuffix([]byte("abc"), []byte{'A', 'x', 'y'})
	if got != "abcxy" {
		t.Fatalf("got %q", got)
	}
	// truncate 2 ('C'), append "z" to "abcde" → abcz
	got = decodeTrimSuffix([]byte("abcde"), []byte{'C', 'z'})
	if got != "abcz" {
		t.Fatalf("got %q", got)
	}
	// REMOVE_EVERYTHING (255): encode stores (255+'A')&0xFF; decode recovers 255 → drop all source
	enc := []byte{byte((255 + 'A') & 0xFF), 'x', 'y'}
	got = decodeTrimSuffix([]byte("abc"), enc)
	if got != "xy" {
		t.Fatalf("REMOVE_EVERYTHING got %q want xy (enc[0]=%d)", got, enc[0])
	}
}

func TestDecodeTrimPrefixAndSuffix(t *testing.T) {
	// P=0,K=0 → full source + suffix
	got := decodeTrimPrefixAndSuffix([]byte("abc"), []byte{'A', 'A', 'x'})
	if got != "abcx" {
		t.Fatalf("got %q", got)
	}
	// P=1 ('B'), K=1 ('B'), source "xyzzy", suffix "Q" → yzz + Q
	got = decodeTrimPrefixAndSuffix([]byte("xyzzy"), []byte{'B', 'B', 'Q'})
	if got != "yzzQ" {
		t.Fatalf("got %q", got)
	}
}

func TestDecodeTrimInfixAndSuffix_JavaExamples(t *testing.T) {
	// From TrimInfixAndSuffixEncoder javadoc:
	// src: ayz  dst: abc  encoded: AACbc
	// X=0,L=0,K=2,suffix=bc → a + bc = abc
	got := decodeTrimInfixAndSuffix([]byte("ayz"), []byte{'A', 'A', 'C', 'b', 'c'})
	if got != "abc" {
		t.Fatalf("ayz/AACbc → %q want abc", got)
	}
	// src: aillent  dst: aller  encoded: BBCr
	// X=1,L=1,K=2,suffix=r → a + lle + r = aller
	got = decodeTrimInfixAndSuffix([]byte("aillent"), []byte{'B', 'B', 'C', 'r'})
	if got != "aller" {
		t.Fatalf("aillent/BBCr → %q want aller", got)
	}
}

func TestDecodeTrimInfixAndSuffix_RemoveEverything(t *testing.T) {
	// L or K = REMOVE_EVERYTHING → stem = encoded[3:] only
	// store 255 as (255+'A')&0xFF in one control byte
	rem := byte((255 + 'A') & 0xFF)
	enc := []byte{'A', rem, 'A', 'n', 'e', 'w'}
	got := decodeTrimInfixAndSuffix([]byte("oldstem"), enc)
	if got != "new" {
		t.Fatalf("REMOVE_EVERYTHING infixLength → %q want new", got)
	}
}
