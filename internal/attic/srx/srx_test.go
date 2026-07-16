package srx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadDoc(t *testing.T) *Document {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for {
		p := filepath.Join(dir, "inspiration", "languagetool", "languagetool-core", "src", "main", "resources", "org", "languagetool", "resource", "segment.srx")
		if _, err := os.Stat(p); err == nil {
			doc, err := Load(p)
			if err != nil {
				t.Fatal(err)
			}
			return doc
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("segment.srx not found")
		}
		dir = parent
	}
}

func TestLoadCounts(t *testing.T) {
	doc := loadDoc(t)
	if len(doc.LangRules) < 5 {
		t.Fatalf("lang rules %d", len(doc.LangRules))
	}
	if len(doc.LangRules["English"]) < 10 {
		t.Fatalf("English rules %d", len(doc.LangRules["English"]))
	}
	if len(doc.Maps) < 5 {
		t.Fatalf("maps %d", len(doc.Maps))
	}
	names := doc.ruleNames("en_two")
	if len(names) == 0 {
		t.Fatal("no rule names for en_two")
	}
	t.Logf("en_two -> %v", names)
}

func TestSplitEnglish(t *testing.T) {
	doc := loadDoc(t)
	parts := doc.Split("Hello world. How are you?", "en", "_two")
	if len(parts) < 2 {
		t.Fatalf("expected >=2 sentences, got %#v (rules en_two=%v English=%d)", parts, doc.ruleNames("en_two"), len(doc.LangRules["English"]))
	}
	if !strings.Contains(parts[0], "Hello") {
		t.Fatalf("parts %#v", parts)
	}
}
