package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikPolishSpellerRuleID ports MorfologikPolishSpellerRule.getId().
	// Java returns "MORFOLOGIK_RULE_PL_PL" (not MORFOLOGIK_RULE_PL).
	MorfologikPolishSpellerRuleID = "MORFOLOGIK_RULE_PL_PL"
	// PolishSpellerDict ports MorfologikPolishSpellerRule.getFileName() → RESOURCE_FILENAME.
	// Java: "/pl/hunspell/pl_PL.dict"
	PolishSpellerDict = "/pl/hunspell/pl_PL.dict"
)

// Java MorfologikPolishSpellerRule.tokenizingPattern: (?:[Qq]uasi|[Nn]iby)-
var polishTokenizingPattern = regexp.MustCompile(`(?:[Qq]uasi|[Nn]iby)-`)

// Java prefixes set (static block) — never split these in compound accept.
var polishPrefixes = map[string]struct{}{
	"arcy": {}, "neo": {}, "pre": {}, "anty": {}, "eks": {}, "bez": {}, "beze": {},
	"ekstra": {}, "hiper": {}, "infra": {}, "kontr": {}, "maksi": {}, "midi": {},
	"między": {}, "mini": {}, "nad": {}, "nade": {}, "około": {}, "ponad": {},
	"post": {}, "pro": {}, "przeciw": {}, "pseudo": {}, "super": {}, "śród": {},
	"ultra": {}, "wice": {}, "wokół": {}, "wokoło": {},
}

// Java bannedSuffixes — prune space-split suggestions whose 2nd token is in set.
var polishBannedSuffixes = map[string]struct{}{
	"ami": {}, "ach": {}, "e": {}, "ego": {}, "em": {}, "emu": {}, "ie": {},
	"im": {}, "m": {}, "om": {}, "owie": {}, "owi": {}, "ze": {},
}

// MorfologikPolishSpellerRule ports rules.pl.MorfologikPolishSpellerRule.
// getRuleMatches: isNotCompound, lower-only suggestion, pruneSuggestions;
// tokenizingPattern for niby-/quasi- (checked as remainder only).
type MorfologikPolishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// TagPOS optional Polish tagger for isNotCompound adj compounds (adja / num:comp / adj:).
	// Fail-closed when nil: only prefix-based compounds accepted without inventing POS.
	TagPOS func(word string) []string
	// incorrectExamples / correctExamples port Rule.addExamplePair (not on SpellingCheckRule:
	// import cycle with rules package — same pattern as AbstractEnglishSpellerRule / RU).
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewMorfologikPolishSpellerRule() *MorfologikPolishSpellerRule {
	r := &MorfologikPolishSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikPolishSpellerRuleID, "pl", PolishSpellerDict, nil),
	}
	// Java tokenizingPattern(): (?:[Qq]uasi|[Nn]iby)- — base Match splits segments.
	r.TokenizingPattern = polishTokenizingPattern
	// Java: bledem → błędem (wrong example omits trailing period, same as upstream)
	r.AddExamplePair(
		rules.Wrong("To jest zdanie z <marker>bledem</marker>"),
		rules.Fixed("To jest zdanie z <marker>błędem</marker>."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *MorfologikPolishSpellerRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *MorfologikPolishSpellerRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *MorfologikPolishSpellerRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports parent Match with Polish getRuleMatches arms:
// tokenizingPattern split, isNotCompound drop, lower-only sug, pruneSuggestions.
func (r *MorfologikPolishSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil, nil
	}
	base, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil || len(base) == 0 {
		return base, err
	}
	out := make([]*rules.RuleMatch, 0, len(base))
	for _, m := range base {
		if m == nil {
			continue
		}
		word := matchSurfacePL(m, sentence)
		if word == "" {
			out = append(out, m)
			continue
		}
		// tokenizingPattern: if niby-/quasi- present, only remainder is spell-checked in Java.
		// Parent already flagged whole word; drop if remainder is accepted (or empty remainder after prefix).
		if polishTokenizingPattern.MatchString(word) {
			if r.acceptTokenizedRemainder(word) {
				continue
			}
		}
		// isNotCompound: false → compound accepted → suppress match
		if !r.isNotCompound(word) {
			continue
		}
		// if lower form accepted → only lower as suggestion
		low := strings.ToLower(word)
		if low != word && !r.wordIsMisspelled(low) {
			m.SetSuggestedReplacements([]string{low})
			out = append(out, m)
			continue
		}
		sugs := m.GetSuggestedReplacements()
		if len(sugs) == 0 && FilterDictAvailable() {
			sugs = FilterDictSuggest(word)
		}
		if len(sugs) > 0 {
			m.SetSuggestedReplacements(prunePolishSuggestions(sugs))
		}
		out = append(out, m)
	}
	return out, nil
}

// acceptTokenizedRemainder ports tokenizingPattern loop: if pattern matches,
// trailing segment after last match is the part checked (prefix segment often empty).
func (r *MorfologikPolishSpellerRule) acceptTokenizedRemainder(word string) bool {
	idxs := polishTokenizingPattern.FindAllStringIndex(word, -1)
	if len(idxs) == 0 {
		return false
	}
	// Java: after last find, check word[index:end]; also checks between matches.
	// For typical "Niby-artysta" single match at start: only "artysta" is checked.
	index := 0
	allOK := true
	anyPart := false
	for _, span := range idxs {
		part := word[index:span[0]]
		if part != "" {
			anyPart = true
			if r.wordIsMisspelled(part) {
				allOK = false
			}
		}
		index = span[1]
	}
	if index < len(word) {
		part := word[index:]
		if part != "" {
			anyPart = true
			if r.wordIsMisspelled(part) {
				allOK = false
			}
		}
	}
	// If we only saw the tokenizing match and no remainder, do not auto-accept whole word.
	return anyPart && allOK
}

// isNotCompound ports MorfologikPolishSpellerRule.isNotCompound.
// Returns true when the word is NOT a recognized compound (should still be flagged if misspelled).
// Returns false when a prefix or adj compound is recognized (suppress misspell).
func (r *MorfologikPolishSpellerRule) isNotCompound(word string) bool {
	if r == nil || word == "" {
		return true
	}
	runes := []rune(word)
	n := len(runes)
	for i := 2; i < n; i++ {
		first := string(runes[:i])
		second := string(runes[i:])
		// prefix path: prefix in set, second not misspelled, second longer than first
		if _, ok := polishPrefixes[strings.ToLower(first)]; ok {
			if !r.wordIsMisspelled(second) && tokenizers.UTF16Len(second) > tokenizers.UTF16Len(first) {
				return false
			}
		}
		// adj compound path needs TagPOS
		if r.TagPOS != nil {
			t0 := r.TagPOS(first)
			t1 := r.TagPOS(second)
			if hasExactTagPL(t0, "adja") || (hasExactTagPL(t0, "num:comp") && !hasExactTagPL(t0, "adv")) {
				if hasPartialTagPL(t1, "adj:") {
					return false
				}
			}
		}
	}
	return true
}

// prunePolishSuggestions ports pruneSuggestions: drop "word suffix" where suffix is banned.
func prunePolishSuggestions(suggestions []string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	out := make([]string, 0, len(suggestions))
	for _, sug := range suggestions {
		if !strings.Contains(sug, " ") {
			out = append(out, sug)
			continue
		}
		parts := strings.Split(sug, " ")
		if len(parts) > 1 {
			if _, banned := polishBannedSuffixes[parts[1]]; banned {
				continue
			}
		}
		out = append(out, sug)
	}
	return out
}

func (r *MorfologikPolishSpellerRule) wordIsMisspelled(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if r.IsMisspelled != nil {
		return r.IsMisspelled(word)
	}
	if r.Speller != nil {
		return r.Speller.IsMisspelled(word)
	}
	return false
}

func hasExactTagPL(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func hasPartialTagPL(tags []string, prefix string) bool {
	for _, t := range tags {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	return false
}

func matchSurfacePL(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	text := sent.GetText()
	from, to := m.GetFromPos(), m.GetToPos()
	if from < 0 || from >= to {
		return ""
	}
	runes := []rune(text)
	if to <= len(runes) {
		return string(runes[from:to])
	}
	if to <= len(text) {
		return text[from:to]
	}
	return ""
}
