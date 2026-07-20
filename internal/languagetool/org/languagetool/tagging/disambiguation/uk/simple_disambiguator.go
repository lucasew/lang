package uk

import (
	"bufio"
	"embed"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rulesuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

//go:embed data/disambig_remove.txt data/disambig_dups.txt
var disambigFS embed.FS

// particle suffix pattern for -то/-бо etc. (Java SimpleDisambiguator.PATTERN)
var reParticleSuffix = regexp.MustCompile(`.*-(то|от|таки|бо|но)$`)

// MatcherEntry is one lemma+pos pair to remove (Java MatcherEntry: full POS regex match).
type MatcherEntry struct {
	Lemma string
	POS   *regexp.Regexp // Java Pattern.compile(tagRegex).matcher(pos).matches()
}

// TokenMatcher holds entries to strip from a token's readings.
type TokenMatcher struct {
	Entries []MatcherEntry
}

func (m *TokenMatcher) Matches(tok *languagetool.AnalyzedToken) bool {
	if m == nil || tok == nil {
		return false
	}
	lemma := ""
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	pos := ""
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	if pos == "" {
		return false // Java !hasNoTag
	}
	for _, e := range m.Entries {
		if e.Lemma != "*" && e.Lemma != "" && lemma != e.Lemma {
			continue
		}
		// Java Pattern.matcher(posTag).matches() = entire string
		if e.POS == nil || !posMatchesFullDisambig(e.POS, pos) {
			continue
		}
		return true
	}
	return false
}

// posMatchesFullDisambig ports Matcher.matches() for disambig POS patterns.
func posMatchesFullDisambig(re *regexp.Regexp, pos string) bool {
	if re == nil || pos == "" {
		return false
	}
	loc := re.FindStringIndex(pos)
	return loc != nil && loc[0] == 0 && loc[1] == len(pos)
}

// SimpleDisambiguator ports tagging.disambiguation.uk.SimpleDisambiguator.
type SimpleDisambiguator struct {
	disambiguation.AbstractDisambiguator
	RemoveMap map[string]*TokenMatcher
	// DupsMap: if key lemma is present, remove readings whose lemma is in the value list.
	DupsMap map[string][]string
}

func NewSimpleDisambiguator() *SimpleDisambiguator {
	return NewSimpleDisambiguatorFull(LoadDisambigRemoveMap(), LoadDisambigDupsMap())
}

// NewSimpleDisambiguatorWith starts with an explicit remove map (tests).
func NewSimpleDisambiguatorWith(m map[string]*TokenMatcher) *SimpleDisambiguator {
	if m == nil {
		m = map[string]*TokenMatcher{}
	}
	return &SimpleDisambiguator{RemoveMap: m, DupsMap: map[string][]string{}}
}

// NewSimpleDisambiguatorFull sets remove + dups maps.
func NewSimpleDisambiguatorFull(remove map[string]*TokenMatcher, dups map[string][]string) *SimpleDisambiguator {
	if remove == nil {
		remove = map[string]*TokenMatcher{}
	}
	if dups == nil {
		dups = map[string][]string{}
	}
	return &SimpleDisambiguator{RemoveMap: remove, DupsMap: dups}
}

func (d *SimpleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	remove := map[string]*TokenMatcher(nil)
	dups := map[string][]string(nil)
	if d != nil {
		remove = d.RemoveMap
		dups = d.DupsMap
	}
	RemoveRareForms(input, remove)
	RemoveDuplicateLemmas(input, dups)
	return input
}

var (
	disambigRemoveOnce sync.Once
	disambigRemoveMap  map[string]*TokenMatcher
	disambigDupsOnce   sync.Once
	disambigDupsMap    map[string][]string
)

// LoadDisambigRemoveMap loads official /uk/disambig_remove.txt (embedded).
func LoadDisambigRemoveMap() map[string]*TokenMatcher {
	disambigRemoveOnce.Do(func() {
		disambigRemoveMap = loadDisambigRemoveFromFS()
	})
	return disambigRemoveMap
}

// LoadDisambigDupsMap loads official /uk/disambig_dups.txt (embedded).
func LoadDisambigDupsMap() map[string][]string {
	disambigDupsOnce.Do(func() {
		disambigDupsMap = loadDisambigDupsFromFS()
	})
	return disambigDupsMap
}

func loadDisambigRemoveFromFS() map[string]*TokenMatcher {
	f, err := disambigFS.Open("data/disambig_remove.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	out := map[string]*TokenMatcher{}
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.Index(line, "#"); i >= 0 {
			// strip trailing comment after space-hash (Java: replaceFirst(" *#.*", ""))
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key := parts[0]
		var entries []MatcherEntry
		for _, matcher := range strings.Split(parts[1], "|") {
			matcher = strings.TrimSpace(matcher)
			if matcher == "" {
				continue
			}
			mp := strings.SplitN(matcher, " ", 2)
			if len(mp) < 2 {
				continue
			}
			lemma, posRE := mp[0], mp[1]
			// Java matches() is full-string
			re, err := regexp.Compile("^(?:" + posRE + ")$")
			if err != nil {
				// fail-closed: skip bad pattern rather than invent
				continue
			}
			entries = append(entries, MatcherEntry{Lemma: lemma, POS: re})
		}
		if len(entries) > 0 {
			out[key] = &TokenMatcher{Entries: entries}
		}
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
	return out
}

func loadDisambigDupsFromFS() map[string][]string {
	f, err := disambigFS.Open("data/disambig_dups.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	out := map[string][]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		out[parts[0]] = append([]string(nil), parts[1:]...)
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
	return out
}

// RemoveDuplicateLemmas drops secondary lemmas when a preferred lemma is present.
func RemoveDuplicateLemmas(input *languagetool.AnalyzedSentence, dups map[string][]string) {
	if input == nil || len(dups) == 0 {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		readings := tokens[i].GetReadings()
		present := map[string]struct{}{}
		for _, r := range readings {
			if r != nil && r.GetLemma() != nil {
				present[*r.GetLemma()] = struct{}{}
			}
		}
		toRemove := map[string]struct{}{}
		for preferred, seconds := range dups {
			if _, ok := present[preferred]; !ok {
				continue
			}
			for _, s := range seconds {
				toRemove[s] = struct{}{}
			}
		}
		if len(toRemove) == 0 {
			continue
		}
		copyR := append([]*languagetool.AnalyzedToken(nil), readings...)
		for j := len(copyR) - 1; j >= 0; j-- {
			r := copyR[j]
			if r == nil || r.GetLemma() == nil {
				continue
			}
			if _, ok := toRemove[*r.GetLemma()]; ok {
				tokens[i].RemoveReading(r, "dis_remove_dups")
			}
		}
	}
}

// RemoveRareForms strips readings matching RemoveMap (in-place).
func RemoveRareForms(input *languagetool.AnalyzedSentence, removeMap map[string]*TokenMatcher) {
	if input == nil || len(removeMap) == 0 {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		// Java getCleanToken
		token := tokens[i].GetCleanToken()
		if token == "" {
			token = tokens[i].GetToken()
		}
		if token == "" {
			continue
		}
		// Java: if Character.isLowerCase(token.charAt(0)) then toLowerCase()
		if r := []rune(token); len(r) > 0 && unicode.IsLower(r[0]) {
			token = strings.ToLower(token)
		}
		tm := lookupMatcher(token, removeMap)
		if tm == nil {
			continue
		}
		readings := append([]*languagetool.AnalyzedToken(nil), tokens[i].GetReadings()...)
		for j := len(readings) - 1; j >= 0; j-- {
			if tm.Matches(readings[j]) {
				tokens[i].RemoveReading(readings[j], "dis_remove_rare")
			}
		}
	}
}

func lookupMatcher(token string, removeMap map[string]*TokenMatcher) *TokenMatcher {
	if tm := removeMap[token]; tm != nil {
		return tm
	}
	low := strings.ToLower(token)
	if tm := removeMap[low]; tm != nil {
		return tm
	}
	if reParticleSuffix.MatchString(low) {
		if idx := strings.LastIndex(low, "-"); idx > 0 {
			if tm := removeMap[low[:idx]]; tm != nil {
				return tm
			}
			// also try original-case base
			if idx2 := strings.LastIndex(token, "-"); idx2 > 0 {
				if tm := removeMap[token[:idx2]]; tm != nil {
					return tm
				}
			}
		}
	}
	return nil
}

// RemoveVmisReadings drops v_mis when another non-end reading remains (token-local helper).
func RemoveVmisReadings(atr *languagetool.AnalyzedTokenReadings) {
	if atr == nil || !canRemoveVmis(atr.GetReadings()) {
		return
	}
	readings := append([]*languagetool.AnalyzedToken(nil), atr.GetReadings()...)
	for _, r := range readings {
		if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), "v_mis") {
			atr.RemoveReading(r, "dis_v_mis")
		}
	}
}

func canRemoveVmis(analyzed []*languagetool.AnalyzedToken) bool {
	foundVmis, foundOther := false, false
	for _, token := range analyzed {
		if token == nil {
			continue
		}
		pos := token.GetPOSTag()
		if pos != nil && strings.Contains(*pos, "v_mis") {
			foundVmis = true
		} else if pos != nil && !strings.HasSuffix(*pos, "_END") {
			foundOther = true
		}
		if foundVmis && foundOther {
			return true
		}
	}
	return false
}

var (
	vMisPrepOnce sync.Once
	vMisPreps    map[string]struct{}
	// Java PATTERN_1 for removeVmis startCheck
	vMisStartWordRE = regexp.MustCompile(`^[а-яіїєґa-z0-9].*`)
)

// loadVMisPreps ports V_MIS_PREPS from CaseGovernmentHelper map + y/B.
func loadVMisPreps() map[string]struct{} {
	vMisPrepOnce.Do(func() {
		out := map[string]struct{}{}
		cg := rulesuk.LoadCaseGovernmentHelper()
		if cg != nil {
			for lemma, cases := range cg.Map {
				if _, ok := cases["v_mis"]; ok {
					out[strings.ToLower(lemma)] = struct{}{}
				}
			}
		}
		// Java static: Latin y/B often used instead of real prep
		out["y"] = struct{}{}
		out["b"] = struct{}{} // lower of "B"
		out["B"] = struct{}{}
		vMisPreps = out
	})
	return vMisPreps
}

// RemoveVmis ports UkrainianHybridDisambiguator.removeVmis (sentence-level).
// Starts after first prep or capitalized word; aborts when a v_mis prep is seen.
func RemoveVmis(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	preps := loadVMisPreps()
	tokens := input.GetTokensWithoutWhitespace()
	startCheck := false
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil || tok.GetToken() == "" {
			continue
		}
		lower := strings.ToLower(tok.GetToken())
		hasPrep := hasPOSPart(tok, "prep")

		if !startCheck {
			if hasPrep {
				startCheck = true
			} else if vMisStartWordRE.MatchString(lower) {
				// Java: all-lowercase non-prep → continue without startCheck
				if isAllLowerCaseToken(tok.GetToken()) {
					continue
				}
				startCheck = true
			}
		}

		if hasPrep {
			if _, ok := preps[lower]; ok {
				return // Java: leave rest of sentence alone
			}
		}

		if !startCheck {
			continue
		}
		if !canRemoveVmis(tok.GetReadings()) {
			continue
		}
		readings := append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...)
		for _, r := range readings {
			if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), "v_mis") {
				tok.RemoveReading(r, "dis_v_mis")
			}
		}
	}
}

func hasPOSPart(tok *languagetool.AnalyzedTokenReadings, part string) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), part) {
			return true
		}
	}
	return false
}

func isAllLowerCaseToken(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsLower(r) {
				return false
			}
		}
	}
	return hasLetter
}
