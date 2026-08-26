package morfologik

import (
	"path/filepath"
	"testing"
)

func TestIsCamelCase_JavaTwin(t *testing.T) {
	// Java: first upper, second lower, not capitalized (internal upper), not all upper
	if !isCamelCase("iPhone") {
		// iPhone: first is lower 'i' → Java false
		// "GreatElephant": first upper, second lower, internal upper → true
	}
	if !isCamelCase("GreatElephant") {
		t.Fatal("GreatElephant should be camel case")
	}
	if !isCamelCase("Waschmaschinen-Test") {
		t.Fatal("dash compound camel case")
	}
	if isCamelCase("Water") {
		t.Fatal("Water is capitalized, not camel")
	}
	if isCamelCase("WATER") {
		t.Fatal("all upper not camel")
	}
	if isCamelCase("water") {
		t.Fatal("lower not camel")
	}
	if isCamelCase("iPhone") {
		t.Fatal("iPhone first lower → not Java camel")
	}
}

func TestIsAllUppercase_JavaTwin(t *testing.T) {
	if !isAllUppercase("WATER") {
		t.Fatal("WATER")
	}
	if !isAllUppercase("123") {
		t.Fatal("digits-only is all-upper in Java (no lowercase letter)")
	}
	if !isAllUppercase("") {
		t.Fatal("empty")
	}
	if isAllUppercase("Water") {
		t.Fatal("Water")
	}
}

func TestDictionary_IsMisspelled_EN(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	// known
	if d.IsMisspelled("software") {
		t.Fatal("software")
	}
	if d.IsMisspelled("Water") { // convertCase
		t.Fatal("Water via convertCase")
	}
	if d.IsMisspelled("WATER") {
		t.Fatal("WATER via convertCase")
	}
	// ignore-numbers default true
	if d.IsMisspelled("123454") {
		t.Fatal("numbers ignored")
	}
	// true misspellings
	if !d.IsMisspelled("bicylce") {
		t.Fatal("bicylce should be misspelled")
	}
	if !d.IsMisspelled("sdadsadas") {
		t.Fatal("garbage misspelled")
	}
}

func TestDictionary_Locale_EN(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	if d.Locale != "en_US" {
		t.Fatalf("locale=%q want en_US", d.Locale)
	}
	if d.ToLower("WATER") != "water" {
		t.Fatalf("ToLower WATER → %q", d.ToLower("WATER"))
	}
	if d.ToUpper("water") != "WATER" {
		t.Fatalf("ToUpper water → %q", d.ToUpper("water"))
	}
}

func TestToLower_TurkishLocale(t *testing.T) {
	// Java Locale("tr") : I → ı (dotless)
	d := &Dictionary{}
	d.setLocale("tr_TR")
	got := d.ToLower("I")
	if got != "ı" {
		t.Fatalf("Turkish ToLower(I)=%q want ı", got)
	}
}

func TestIsInDictionary_EN(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	if !d.IsInDictionary("house") {
		t.Fatal("house in dict")
	}
	if !d.IsInDictionary("software") {
		t.Fatal("software")
	}
	if d.IsInDictionary("xyzzyqqqnotaword") {
		t.Fatal("garbage not in dict")
	}
	if d.IsInDictionary("") {
		t.Fatal("empty")
	}
}
