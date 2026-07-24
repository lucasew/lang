package pattern

import (
	"os"
	"path/filepath"
	"testing"
)

func repoGrammar(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		p := filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "rules", "en", "grammar.xml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("grammar.xml not found")
		}
		dir = parent
	}
}

func TestLoadEnglishGrammar(t *testing.T) {
	path := repoGrammar(t)
	rules, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) < 1000 {
		t.Fatalf("expected thousands of rules, got %d", len(rules))
	}
	var unicode *Rule
	var noPOS, withPOS int
	for _, r := range rules {
		if r.RequiresPOS {
			withPOS++
		} else {
			noPOS++
		}
		if r.ID == "UNICODE_CASING" {
			unicode = r
		}
	}
	t.Logf("rules total=%d noPOS=%d withPOS=%d", len(rules), noPOS, withPOS)
	if unicode == nil {
		t.Fatal("UNICODE_CASING not found")
	}
	if unicode.RequiresPOS {
		t.Fatal("UNICODE_CASING should not require POS")
	}
}

func TestUNICODE_CASING_Match(t *testing.T) {
	path := repoGrammar(t)
	rules, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var unicode *Rule
	for _, r := range rules {
		if r.ID == "UNICODE_CASING" {
			unicode = r
			break
		}
	}
	if unicode == nil {
		t.Fatal("missing rule")
	}
	text := "The unicode standard defines almost 150,000 characters."
	ctx := NewMatchContext("t", "en", text, 0, nil)
	got := MatchRule(unicode, ctx, false)
	if len(got) != 1 {
		t.Fatalf("want 1 match, got %d (%+v)", len(got), got)
	}
	if got[0].Rule != "UNICODE_CASING" {
		t.Errorf("rule %s", got[0].Rule)
	}
	if len(got[0].Suggestions) == 0 || got[0].Suggestions[0] != "Unicode" {
		t.Errorf("suggestions %+v", got[0].Suggestions)
	}
}

func TestOXFORD_COMMA_CASING(t *testing.T) {
	path := repoGrammar(t)
	rules, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rule *Rule
	for _, r := range rules {
		if r.ID == "OXFORD_COMMA_CASING" {
			rule = r
			break
		}
	}
	if rule == nil {
		t.Fatal("missing")
	}
	ctx := NewMatchContext("t", "en", "Are you using the oxford comma?", 0, nil)
	got := MatchRule(rule, ctx, false)
	if len(got) != 1 {
		t.Fatalf("got %d want 1: %+v", len(got), got)
	}
}
