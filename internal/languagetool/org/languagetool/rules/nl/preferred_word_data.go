package nl

import (
	"bufio"
	"fmt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"io"
	"strings"
)

// PreferredWordRuleWithSuggestion ports org.languagetool.rules.nl.PreferredWordRuleWithSuggestion.
// Rule field is left as any — full PatternRule wiring is optional until the pattern engine is ready.
type PreferredWordRuleWithSuggestion struct {
	Rule    any
	OldWord string
	NewWord string
}

// PreferredWordData ports org.languagetool.rules.nl.PreferredWordData as a CSV pair loader.
type PreferredWordData struct {
	Rules []PreferredWordRuleWithSuggestion
}

// LoadPreferredWordData parses lines "old;new" (skip # comments).
func LoadPreferredWordData(r io.Reader, filePathHint string) (*PreferredWordData, error) {
	d := &PreferredWordData{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			hint := filePathHint
			if hint == "" {
				hint = "preferred words"
			}
			return nil, fmt.Errorf("unexpected format in file %s: %s", hint, line)
		}
		oldW := tools.JavaStringTrim(parts[0])
		newW := tools.JavaStringTrim(parts[1])
		d.Rules = append(d.Rules, PreferredWordRuleWithSuggestion{
			OldWord: oldW,
			NewWord: newW,
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *PreferredWordData) Get() []PreferredWordRuleWithSuggestion {
	return d.Rules
}
