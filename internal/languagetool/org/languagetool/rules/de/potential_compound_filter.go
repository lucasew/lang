package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PotentialCompoundFilter ports org.languagetool.rules.de.PotentialCompoundFilter.
//
// Java: GermanyGerman.getDefaultSpellingRule().match(joinedSentence); matches.length==0 → joined valid.
// JoinedIsValid overrides the process-wide filter speller when set (tests).
// Without override and without WireGermanFilterSpeller → hyphenated only
// (fail-closed: do not invent that a joined compound is spelled correctly).
type PotentialCompoundFilter struct {
	// JoinedIsValid optional override; nil uses !FilterDictIsMisspelled when dict wired.
	JoinedIsValid func(joined string) bool
}

func NewPotentialCompoundFilter() *PotentialCompoundFilter {
	return &PotentialCompoundFilter{}
}

// joinedValid ports Java spelling-rule match (empty matches ⇒ valid joined form).
func (f *PotentialCompoundFilter) joinedValid(joined string) bool {
	if f != nil && f.JoinedIsValid != nil {
		return f.JoinedIsValid(joined)
	}
	if FilterDictAvailable() {
		// Java: RuleMatch[] matches = getDefaultSpellingRule().match(...); matches.length == 0
		return !FilterDictIsMisspelled(joined)
	}
	return false
}

// Suggestions returns replacement strings for part1+part2 (Java twin).
func (f *PotentialCompoundFilter) Suggestions(part1, part2 string) []string {
	p1cap := capitalizeIfUniform(part1)
	p2low, p2cap := part2, part2
	if !isMixedOrAllUpper(part2) {
		p2low = strings.ToLower(part2)
		p2cap = tools.UppercaseFirstChar(strings.ToLower(part2))
	}
	if !isMixedOrAllUpper(part1) {
		p1cap = tools.UppercaseFirstChar(strings.ToLower(part1))
	}
	joined := p1cap + p2low
	hyphenated := p1cap + "-" + p2cap
	var out []string
	if f.joinedValid(joined) {
		// Java: if joinedWord.length() > 20 also suggest hyphenated first
		if utf8.RuneCountInString(joined) > 20 {
			out = append(out, hyphenated)
		}
		out = append(out, joined)
	} else {
		// misspelled / no speller → hyphenated only
		out = append(out, hyphenated)
	}
	return out
}

// AcceptRuleMatch ports PotentialCompoundFilter.acceptRuleMatch.
func (f *PotentialCompoundFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	reps := f.Suggestions(arguments["part1"], arguments["part2"])
	if len(reps) == 0 {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	out.SetSuggestedReplacements(reps)
	return out
}

func isMixedOrAllUpper(s string) bool {
	hasLower, hasUpper := false, false
	for _, r := range s {
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsUpper(r) {
			hasUpper = true
		}
	}
	if hasUpper && !hasLower {
		return true // all upper
	}
	return hasUpper && hasLower
}

func capitalizeIfUniform(s string) string {
	if isMixedOrAllUpper(s) {
		return s
	}
	return tools.UppercaseFirstChar(strings.ToLower(s))
}
