package commandline

import (
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/bitext"
)

// CheckBitextFile runs bitext rules on tab-separated source\ttarget lines.
// Returns total match count.
func CheckBitextFile(w io.Writer, contents string, rulesList []bitext.BitextRule) (int, error) {
	if w == nil {
		w = io.Discard
	}
	total := 0
	for i, line := range strings.Split(contents, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			// also allow | separator
			parts = strings.SplitN(line, "|", 2)
		}
		if len(parts) != 2 {
			continue
		}
		src, trg := parts[0], parts[1]
		matches := bitext.CheckBitext(src, trg, rulesList)
		for _, m := range matches {
			total++
			fmt.Fprintf(w, "%d.) Line %d, Rule ID: %s\nMessage: %s\n", total, i+1, m.RuleID, m.Message)
		}
	}
	return total, nil
}
