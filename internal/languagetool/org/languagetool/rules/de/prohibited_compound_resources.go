package de

import (
	"bufio"
	"embed"
	"strings"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/compound_exceptions.txt data/confusion_sets.txt
var prohibitedCompoundFS embed.FS

// Java ProhibitedCompoundRule.ignoreWords
var prohibitedIgnoreWords = map[string]struct{}{
	"Die": {}, "De": {},
}

var (
	prohibitedResOnce    sync.Once
	prohibitedAllPairs   []prohibitedPair
	prohibitedExceptions map[string]struct{}
)

// ProhibitedCompoundExceptions returns words from compound_exceptions.txt.
func ProhibitedCompoundExceptions() map[string]struct{} {
	initProhibitedResources()
	return prohibitedExceptions
}

// AllProhibitedPairs returns lowercasePairs + case variants + confusion_sets pairs (Java static init).
func AllProhibitedPairs() []prohibitedPair {
	initProhibitedResources()
	return prohibitedAllPairs
}

func initProhibitedResources() {
	prohibitedResOnce.Do(func() {
		prohibitedExceptions = loadCompoundExceptions()
		pairs := make([]prohibitedPair, 0, len(lowercaseProhibitedPairs)*2+64)
		// Java addUpperCaseVariants
		for _, lc := range lowercaseProhibitedPairs {
			pairs = append(pairs, lc)
			uc1, uc2 := tools.UppercaseFirstChar(lc.part1), tools.UppercaseFirstChar(lc.part2)
			if lc.part1 != uc1 || lc.part2 != uc2 {
				pairs = append(pairs, prohibitedPair{uc1, lc.desc1, uc2, lc.desc2})
			}
		}
		// Java addItemsFromConfusionSets(..., isUpperCase=true)
		pairs = append(pairs, loadConfusionPairsForProhibited()...)
		prohibitedAllPairs = pairs
	})
}

func loadCompoundExceptions() map[string]struct{} {
	m := map[string]struct{}{}
	f, err := prohibitedCompoundFS.Open("data/compound_exceptions.txt")
	if err != nil {
		return m
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip trailing comment
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		m[line] = struct{}{}
	}
	return m
}

func loadConfusionPairsForProhibited() []prohibitedPair {
	f, err := prohibitedCompoundFS.Open("data/confusion_sets.txt")
	if err != nil {
		return nil
	}
	defer f.Close()
	loaded, err := rules.NewConfusionSetLoader(nil).LoadConfusionPairs(f)
	if err != nil || loaded == nil {
		// fail closed: incomplete without inventing pairs
		return nil
	}
	// Deduplicate pair strings we emit
	seen := map[string]struct{}{}
	var out []prohibitedPair
	// Walk unique ConfusionPair objects (map has both directions)
	for _, list := range loaded {
		for _, p := range list {
			if p == nil {
				continue
			}
			terms := p.GetTerms()
			if len(terms) != 2 || terms[0] == nil || terms[1] == nil {
				continue
			}
			s1, s2 := terms[0].GetString(), terms[1].GetString()
			// Java: allUpper && not ignoreWords
			if !startsWithUppercaseDE(s1) || !startsWithUppercaseDE(s2) {
				continue
			}
			if _, ok := prohibitedIgnoreWords[s1]; ok {
				continue
			}
			if _, ok := prohibitedIgnoreWords[s2]; ok {
				continue
			}
			d1, d2 := "", ""
			if terms[0].GetDescription() != nil {
				d1 = *terms[0].GetDescription()
			}
			if terms[1].GetDescription() != nil {
				d2 = *terms[1].GetDescription()
			}
			// Java adds (part1, part2) and lowercaseFirstChar variants when isUpperCase
			key := s1 + "\x00" + s2
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				out = append(out, prohibitedPair{s1, d1, s2, d2})
			}
			lc1, lc2 := tools.LowercaseFirstChar(s1), tools.LowercaseFirstChar(s2)
			key2 := lc1 + "\x00" + lc2
			if _, ok := seen[key2]; !ok {
				seen[key2] = struct{}{}
				out = append(out, prohibitedPair{lc1, d1, lc2, d2})
			}
		}
	}
	return out
}

func startsWithUppercaseDE(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsUpper(r)
}
