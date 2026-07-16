package rules

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// MaxCompoundTerms ports AbstractCompoundRule.MAX_TERMS.
const MaxCompoundTerms = 5

// CompoundRuleData ports org.languagetool.rules.CompoundRuleData.
type CompoundRuleData struct {
	IncorrectCompounds        map[string]struct{}
	JoinedSuggestion          map[string]struct{}
	JoinedLowerCaseSuggestion map[string]struct{}
	DashSuggestion            map[string]struct{}
	HasDigitPatterns          bool
}

// NewCompoundRuleData loads compound lists from a reader.
func NewCompoundRuleData(r io.Reader, path string) (*CompoundRuleData, error) {
	d := &CompoundRuleData{
		IncorrectCompounds:        make(map[string]struct{}),
		JoinedSuggestion:          make(map[string]struct{}),
		JoinedLowerCaseSuggestion: make(map[string]struct{}),
		DashSuggestion:            make(map[string]struct{}),
	}
	if err := d.loadCompoundFile(r, path); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *CompoundRuleData) loadCompoundFile(r io.Reader, path string) error {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || line[0] == '#' {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		} else {
			line = strings.TrimSpace(line)
		}
		if line == "" {
			continue
		}
		expLine := strings.ReplaceAll(line, "-", " ")
		if err := d.validateLine(path, expLine); err != nil {
			return err
		}
		switch {
		case strings.HasSuffix(expLine, "+"):
			expLine = expLine[:len(expLine)-1]
			d.JoinedSuggestion[expLine] = struct{}{}
		case strings.HasSuffix(expLine, "*"):
			expLine = expLine[:len(expLine)-1]
			d.DashSuggestion[expLine] = struct{}{}
		case strings.HasSuffix(expLine, "?"):
			expLine = expLine[:len(expLine)-1]
			d.JoinedSuggestion[expLine] = struct{}{}
			d.JoinedLowerCaseSuggestion[expLine] = struct{}{}
		case strings.HasSuffix(expLine, "$"):
			expLine = expLine[:len(expLine)-1]
			d.JoinedSuggestion[expLine] = struct{}{}
			d.DashSuggestion[expLine] = struct{}{}
			d.JoinedLowerCaseSuggestion[expLine] = struct{}{}
		default:
			d.JoinedSuggestion[expLine] = struct{}{}
			d.DashSuggestion[expLine] = struct{}{}
		}
		d.IncorrectCompounds[expLine] = struct{}{}
		if strings.Contains(expLine, `\d`) {
			d.HasDigitPatterns = true
		}
	}
	return sc.Err()
}

func (d *CompoundRuleData) validateLine(path, line string) error {
	parts := strings.Split(line, " ")
	if len(parts) == 1 {
		return fmt.Errorf("Not a compound in file %s: %s", path, line)
	}
	if len(parts) > MaxCompoundTerms {
		return fmt.Errorf("Too many compound parts in file %s: %s, maximum allowed: %d", path, line, MaxCompoundTerms)
	}
	// Java: incorrectCompounds.contains(line.toLowerCase()) — set stores post-suffix forms,
	// so same raw line with trailing marker rarely collides; keep parity.
	if _, ok := d.IncorrectCompounds[strings.ToLower(line)]; ok {
		return fmt.Errorf("Duplicated word in file %s: %s", path, line)
	}
	return nil
}

func (d *CompoundRuleData) ContainsIncorrect(s string) bool {
	_, ok := d.IncorrectCompounds[s]
	return ok
}

func (d *CompoundRuleData) ContainsDash(s string) bool {
	_, ok := d.DashSuggestion[s]
	return ok
}

func (d *CompoundRuleData) ContainsJoined(s string) bool {
	_, ok := d.JoinedSuggestion[s]
	return ok
}

// JoinedLowerCaseAnyMatch ports stream anyMatch(s -> stringToCheck.contains(s)).
func (d *CompoundRuleData) JoinedLowerCaseAnyMatch(stringToCheck string) bool {
	for s := range d.JoinedLowerCaseSuggestion {
		if strings.Contains(stringToCheck, s) {
			return true
		}
	}
	return false
}
