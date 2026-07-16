package finding

import "testing"

func TestSARIFLevel(t *testing.T) {
	cases := map[string]string{
		"misspelling":   SeverityError,
		"Misspelling":   SeverityError,
		"grammar":       SeverityError,
		"style":         SeverityNote,
		"whitespace":    SeverityWarning,
		"duplication":   SeverityWarning,
		"typographical": SeverityWarning,
		"other":         SeverityWarning,
		"":              SeverityWarning,
	}
	for in, want := range cases {
		if got := SARIFLevel(in); got != want {
			t.Errorf("SARIFLevel(%q)=%q want %q", in, got, want)
		}
	}
	typ, sev := WithType("whitespace")
	if typ != "whitespace" || sev != SeverityWarning {
		t.Fatalf("WithType whitespace: %q %q", typ, sev)
	}
}
