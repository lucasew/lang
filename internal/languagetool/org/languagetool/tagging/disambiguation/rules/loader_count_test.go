package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOfficialENDisambiguationCount(t *testing.T) {
	candidates := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
			"src", "main", "resources", "org", "languagetool", "resource", "en", "disambiguation.xml"),
		filepath.Join("testdata", "upstream", "en", "resource", "disambiguation.xml"),
	}
	wd, _ := os.Getwd()
	var path string
	for dir := wd; ; dir = filepath.Dir(dir) {
		for _, rel := range candidates {
			p := filepath.Join(dir, rel)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				path = p
				break
			}
		}
		if path != "" {
			break
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	if path == "" {
		t.Skip("no official en disambiguation.xml")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rules, _, err := NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "en", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("path=%s rules=%d", path, len(rules))
	if len(rules) < 800 {
		t.Fatalf("expected ~900+ EN disambig rules after rulegroup+and parse, got %d", len(rules))
	}
	found := false
	for _, r := range rules {
		if r == nil || r.GetID() != "INSTAL_INSTALL" {
			continue
		}
		found = true
		if len(r.Tokens) != 1 {
			t.Fatalf("INSTAL_INSTALL: want 1 and-group token, got %d", len(r.Tokens))
		}
		if r.Tokens[0] == nil || len(r.Tokens[0].AndGroup) < 1 {
			t.Fatalf("INSTAL_INSTALL: missing AndGroup members")
		}
		break
	}
	if !found {
		t.Fatal("INSTAL_INSTALL not loaded (and/rulegroup parse failed)")
	}
	// EXCEPT_NOT_VERB: <marker> must keep "except" token (not empty exception-only).
	found = false
	for _, r := range rules {
		if r == nil || r.GetID() != "EXCEPT_NOT_VERB" {
			continue
		}
		found = true
		if len(r.Tokens) < 2 {
			t.Fatalf("EXCEPT_NOT_VERB: want >=2 tokens (marker+follow), got %d", len(r.Tokens))
		}
		if r.Tokens[0].Token != "except" {
			t.Fatalf("EXCEPT_NOT_VERB: first token want except, got %q", r.Tokens[0].Token)
		}
		if !r.Tokens[0].InsideMarker {
			t.Fatal("EXCEPT_NOT_VERB: marked token must be InsideMarker")
		}
		break
	}
	if !found {
		t.Fatal("EXCEPT_NOT_VERB not loaded (marker parse failed)")
	}
}
