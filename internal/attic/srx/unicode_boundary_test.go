package srx

import "testing"

// Java SRX uses UNICODE_CHARACTER_CLASS; RE2 \b is ASCII-only.
func TestUnicodeWordBoundary_RomanianSamd(t *testing.T) {
	doc, err := DefaultDocument()
	if err != nil {
		t.Fatal(err)
	}
	parts := doc.Split("A spus șamd. șamd.", "ro", "_two")
	if len(parts) != 1 {
		t.Fatalf("șamd. must not end sentence (Java segment.srx), got %#v", parts)
	}
}
