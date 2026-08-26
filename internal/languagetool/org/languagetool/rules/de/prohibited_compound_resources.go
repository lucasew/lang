package de

import (
	"bufio"
	"embed"
	"sort"
	"strings"
	"sync"

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
	prohibitedResOnce  sync.Once
	prohibitedAllPairs []prohibitedPair
	// prohibitedPairMap ports pairMap from setupAhoCorasickSearch (part → pairs).
	prohibitedPairMap map[string][]prohibitedPair
	// prohibitedSearchStems are Aho-Corasick dictionary keys (unique pair parts).
	prohibitedSearchStems []string
	prohibitedExceptions  map[string]struct{}
)

// ProhibitedCompoundExceptions returns words from compound_exceptions.txt
// (Java LoadingCache path /de/compound_exceptions.txt).
func ProhibitedCompoundExceptions() map[string]struct{} {
	initProhibitedResources()
	return prohibitedExceptions
}

// AllProhibitedPairs returns lowercasePairs + case variants + confusion_sets pairs
// (Java static init list before Aho-Corasick build).
func AllProhibitedPairs() []prohibitedPair {
	initProhibitedResources()
	return prohibitedAllPairs
}

// prohibitedPairLookup returns pairs registered for stem (Java pairMap.get(hit.value)).
func prohibitedPairLookup(stem string) []prohibitedPair {
	initProhibitedResources()
	return prohibitedPairMap[stem]
}

// prohibitedStems returns dictionary keys for multi-pattern search
// (Java AhoCorasickDoubleArrayTrie map keys).
func prohibitedStems() []string {
	initProhibitedResources()
	return prohibitedSearchStems
}

func initProhibitedResources() {
	prohibitedResOnce.Do(func() {
		prohibitedExceptions = loadCompoundExceptions()
		pairs := make([]prohibitedPair, 0, len(lowercaseProhibitedPairs)*2+64)
		// Java addUpperCaseVariants → addAllCaseVariants
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
		// Java setupAhoCorasickSearch(pairs, pairMap)
		prohibitedPairMap, prohibitedSearchStems = setupProhibitedAhoCorasick(pairs)
	})
}

// setupProhibitedAhoCorasick ports setupAhoCorasickSearch: TreeMap of parts + pairMap lists.
// Search uses stem Contains (equivalent candidate set to AhoCorasick parseText hits).
func setupProhibitedAhoCorasick(pairs []prohibitedPair) (map[string][]prohibitedPair, []string) {
	pairMap := make(map[string][]prohibitedPair)
	// Java TreeMap keys — stable stem order for determinism
	stemSet := map[string]struct{}{}
	for _, pair := range pairs {
		stemSet[pair.part1] = struct{}{}
		stemSet[pair.part2] = struct{}{}
		pairMap[pair.part1] = append(pairMap[pair.part1], pair)
		pairMap[pair.part2] = append(pairMap[pair.part2], pair)
	}
	stems := make([]string, 0, len(stemSet))
	for s := range stemSet {
		stems = append(stems, s)
	}
	// Java TreeMap key order
	sort.Strings(stems)
	return pairMap, stems
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
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip trailing comment (Java: line.replaceFirst("#.*", "").trim())
		if i := strings.Index(line, "#"); i >= 0 {
			line = tools.JavaStringTrim(line[:i])
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
	// Deduplicate pair strings we emit (map has both directions)
	seen := map[string]struct{}{}
	var out []prohibitedPair
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
			// allMatch(k -> startsWithUppercase(k.getString()) && !ignoreWords.contains(k.getString()))
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
	// Java Character.isUpperCase(s.charAt(0)) — UTF-16 first unit
	return tools.StartsWithUppercase(s)
}
