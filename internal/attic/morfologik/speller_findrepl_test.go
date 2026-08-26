package morfologik

import (
	"path/filepath"
	"testing"
)

func enUSDictPath(t *testing.T) string {
	t.Helper()
	// third_party path used by LT port
	candidates := []string{
		"third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict",
		"../third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict",
		"../../third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict",
	}
	// also try from repo root via walking up
	for _, rel := range candidates {
		if p, err := filepath.Abs(rel); err == nil {
			if d, err := OpenDictionary(p); err == nil && d != nil {
				_ = d
				return p
			}
		}
	}
	// discover from cwd
	for _, root := range []string{".", "..", "../..", "../../.."} {
		p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
		if d, err := OpenDictionary(p); err == nil && d != nil {
			return p
		}
	}
	t.Skip("en_US.dict not found")
	return ""
}

func TestSpellerFSA_FindReplacements_Recieve(t *testing.T) {
	path := enUSDictPath(t)
	d, err := OpenDictionary(path)
	if err != nil || d == nil {
		t.Skip(err)
	}
	sp := NewSpellerFSA(d, 1)
	// known word → empty
	if c := sp.FindReplacementCandidates("receive"); len(c) != 0 {
		t.Fatalf("receive should be empty, got %v", c)
	}
	cands := sp.FindReplacementCandidates("recieve")
	if len(cands) == 0 {
		t.Fatal("expected suggestions for recieve")
	}
	found := false
	for _, c := range cands {
		if c.Word == "receive" {
			found = true
			if c.OrigDistance != 1 {
				t.Fatalf("receive origDistance=%d want 1 (weight=%d)", c.OrigDistance, c.Distance)
			}
		}
	}
	if !found {
		t.Fatalf("receive not in %v", cands)
	}
}

func TestSpellerFSA_FindReplacements_Edit2(t *testing.T) {
	path := enUSDictPath(t)
	d, err := OpenDictionary(path)
	if err != nil {
		t.Skip(err)
	}
	sp1 := NewSpellerFSA(d, 1)
	// garentee needs distance 2 for guarantee (classic LT case)
	c1 := sp1.FindReplacementCandidates("garentee")
	for _, c := range c1 {
		if c.Word == "guarantee" {
			t.Fatalf("edit-1 must not return guarantee; got %v", c1)
		}
	}
	sp2 := NewSpellerFSA(d, 2)
	c2 := sp2.FindReplacementCandidates("garentee")
	found := false
	for _, c := range c2 {
		if c.Word == "guarantee" || c.Word == "guaranteed" || c.Word == "guarantees" {
			found = true
		}
	}
	if !found {
		t.Fatalf("edit-2 should suggest guarantee*; got %v", c2)
	}
}

func TestSpellerFSA_HMatrixReset(t *testing.T) {
	path := enUSDictPath(t)
	d, err := OpenDictionary(path)
	if err != nil {
		t.Skip(err)
	}
	sp := NewSpellerFSA(d, 1)
	// Java: HMatrix must be reset each findReplacementCandidates call
	a := sp.FindReplacementCandidates("recieve")
	b := sp.FindReplacementCandidates("recieve")
	if len(a) == 0 || len(b) == 0 {
		t.Fatal("empty")
	}
	// same top result after reuse
	if a[0].Word != b[0].Word {
		t.Fatalf("reuse diverged: %v vs %v", a, b)
	}
}

// Twin of Speller anyToOne/anyToTwo: f→ph and kw→qu via HMatrix path inside findRepl.
func TestSpellerFSA_AnyToOneTwo_EN(t *testing.T) {
	path := enUSDictPath(t)
	d, err := OpenDictionary(path)
	if err != nil {
		t.Skip(err)
	}
	sp := NewSpellerFSA(d, 1)
	// en_US.info style short pairs: f ph, ph f, kw qu
	sp.LoadReplacementPairs([]struct{ From, To string }{
		{"f", "ph"},
		{"ph", "f"},
		{"kw", "qu"},
		{"qu", "kw"},
	})
	// fone → phone (f in word matches pattern for dict "ph" via anyToTwo)
	cands := sp.FindReplacementCandidates("fone")
	foundPhone := false
	for _, c := range cands {
		if c.Word == "phone" {
			foundPhone = true
			// pure short replacement keeps HMatrix depth cost 0 (Java anyToTwo path)
			if c.OrigDistance != 0 {
				t.Fatalf("phone origDistance=%d want 0 (weight=%d)", c.OrigDistance, c.Distance)
			}
		}
	}
	if !foundPhone {
		t.Fatalf("expected phone for fone; got %v", cands)
	}
	// kwality → quality
	cands2 := sp.FindReplacementCandidates("kwality")
	foundQ := false
	for _, c := range cands2 {
		if c.Word == "quality" {
			foundQ = true
			if c.OrigDistance != 0 {
				t.Fatalf("quality origDistance=%d want 0", c.OrigDistance)
			}
		}
	}
	if !foundQ {
		t.Fatalf("expected quality for kwality; got %v", cands2)
	}
}
