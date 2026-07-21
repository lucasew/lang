package de

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SpellingData ports org.languagetool.rules.de.SpellingData as old→new spelling maps.
// German synthesizer expansion for ß→ss forms is left to a pluggable ExpandForms hook.
type SpellingData struct {
	// Map is the body-mode coherency map (old → new).
	Map map[string]string
	// SentenceStartMap includes uppercase-first variants of lowercase pairs.
	SentenceStartMap map[string]string
	// ExpandForms optional: for old spellings with ß that become ss, add forms.
	ExpandForms func(oldSpelling string) []string
}

// LoadSpellingData parses "old;new" CSV lines (comments with #).
func LoadSpellingData(r io.Reader, pathHint string) (*SpellingData, error) {
	return LoadSpellingDataWithExpand(r, pathHint, nil)
}

// LoadSpellingDataWithExpand ports SpellingData.getCoherencyMap for body + sentence-start maps.
// Buffers the full CSV so both modes can be built (Java builds two tries from the same file).
func LoadSpellingDataWithExpand(r io.Reader, pathHint string, expand func(string) []string) (*SpellingData, error) {
	if pathHint == "" {
		pathHint = "spelling data"
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return LoadSpellingDataBoth(string(data), pathHint, expand)
}

// LoadSpellingDataBoth builds body and sentence-start maps from the same content string.
func LoadSpellingDataBoth(content, pathHint string, expand func(string) []string) (*SpellingData, error) {
	body, err := getCoherencyMap(strings.NewReader(content), pathHint, false, expand)
	if err != nil {
		return nil, err
	}
	sent, err := getCoherencyMap(strings.NewReader(content), pathHint, true, expand)
	if err != nil {
		return nil, err
	}
	return &SpellingData{Map: body, SentenceStartMap: sent, ExpandForms: expand}, nil
}

func getCoherencyMap(r io.Reader, filePath string, sentStartMode bool, expand func(string) []string) (map[string]string, error) {
	coherencyMap := map[string]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			return nil, fmt.Errorf("unexpected format in file %s: %s", filePath, line)
		}
		oldSpelling := parts[0]
		newSpelling := parts[1]
		if err := sanityChecks(filePath, line, oldSpelling, newSpelling, coherencyMap); err != nil {
			return nil, err
		}
		if sentStartMode && startsWithLowercase(oldSpelling) && startsWithLowercase(newSpelling) {
			coherencyMap[tools.UppercaseFirstChar(oldSpelling)] = tools.UppercaseFirstChar(newSpelling)
		} else {
			coherencyMap[oldSpelling] = newSpelling
		}
		if strings.Contains(oldSpelling, "ß") && strings.ReplaceAll(oldSpelling, "ß", "ss") == newSpelling {
			if expand != nil {
				for _, form := range expand(oldSpelling) {
					if !strings.Contains(form, "ss") {
						coherencyMap[form] = strings.ReplaceAll(form, "ß", "ss")
					}
				}
			}
		}
	}
	return coherencyMap, sc.Err()
}

func startsWithLowercase(s string) bool {
	// Java SpellingData: StringTools.startsWithLowercase (UTF-16 charAt(0))
	return tools.StartsWithLowercase(s)
}

func sanityChecks(filePath, line, oldSpelling, newSpelling string, coherencyMap map[string]string) error {
	if oldSpelling == newSpelling {
		return fmt.Errorf("old and new spelling are the same in %s: %s", filePath, line)
	}
	if lookup, ok := coherencyMap[newSpelling]; ok && lookup == oldSpelling {
		return fmt.Errorf("contradictory entry in %s: '%s' suggests '%s' and vice versa", filePath, oldSpelling, lookup)
	}
	if prev, ok := coherencyMap[oldSpelling]; ok && prev != newSpelling {
		return fmt.Errorf("duplicate key in %s: %s, val: %s vs. %s", filePath, oldSpelling, prev, newSpelling)
	}
	return nil
}

func (d *SpellingData) Lookup(old string) (string, bool) {
	if d == nil || d.Map == nil {
		return "", false
	}
	v, ok := d.Map[old]
	return v, ok
}
