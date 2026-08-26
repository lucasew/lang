package pipeline

import (
	"strings"
	"testing"
)

func TestEnglishContractions(t *testing.T) {
	toks := EnglishWordTokenize("We don't believe so.")
	var texts []string
	for _, t := range toks {
		if !t.Whitespace {
			texts = append(texts, t.Text)
		}
	}
	joined := strings.Join(texts, "|")
	t.Log(joined)
	// expect do + n't split from don't
	if !strings.Contains(joined, "do|n't") && !strings.Contains(joined, "do|n|'|t") {
		// pattern 2: (do)(n't)
		ok := false
		for i := 0; i+1 < len(texts); i++ {
			if texts[i] == "do" && (texts[i+1] == "n't" || texts[i+1] == "n’t") {
				ok = true
			}
		}
		if !ok {
			t.Fatalf("contraction not split: %v", texts)
		}
	}
}
