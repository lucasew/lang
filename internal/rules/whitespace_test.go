package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/messages"
)

func TestMultipleWhitespace_LTGoldens(t *testing.T) {
	// Positions from MultipleWhitespaceRuleTest.java (LanguageTool core).
	msg := messages.Bundle{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	}

	tests := []struct {
		name    string
		text    string
		wantN   int
		wantPos [][2]int // from,to
	}{
		{"good simple", "This is a test sentence.", 0, nil},
		{"double space", "This  is a test sentence.", 1, [][2]int{{4, 6}}},
		{"leading nl triple space then double", "\n   This  is a test sentence.", 1, [][2]int{{8, 10}}},
		{"triple mid", "This is a test   sentence.", 1, [][2]int{{14, 17}}},
		{"three matches", "This is   a  test   sentence.", 3, [][2]int{{7, 10}, {11, 13}, {17, 20}}},
		{"tabs ok multi", "Multiple tabs\t\tare okay", 0, nil},
		{"nbsp", "This \u00A0is a test sentence.", 1, [][2]int{{4, 6}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MultipleWhitespace(tc.text, "t", "en", msg)
			if len(got) != tc.wantN {
				t.Fatalf("matches: got %d want %d (%+v)", len(got), tc.wantN, got)
			}
			for i, p := range tc.wantPos {
				if got[i].Offset != p[0] || got[i].EndOffset != p[1] {
					t.Errorf("match %d pos: got [%d,%d) want [%d,%d)", i, got[i].Offset, got[i].EndOffset, p[0], p[1])
				}
				if got[i].Rule != RuleWhitespace {
					t.Errorf("rule id: got %s", got[i].Rule)
				}
				if got[i].Message != msg.Get("whitespace_repetition") {
					t.Errorf("message: got %q", got[i].Message)
				}
				if got[i].Severity != SeverityWhitespace {
					t.Errorf("severity: got %s", got[i].Severity)
				}
			}
		})
	}
}
