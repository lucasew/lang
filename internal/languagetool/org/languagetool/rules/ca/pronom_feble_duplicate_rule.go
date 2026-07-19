package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// verbConjugatRE ports PronomFebleDuplicateRule.VERB_CONJUGAT = V.[SI].*
var verbConjugatRE = regexp.MustCompile(`^V.[SI].*$`)

// PronomFebleDuplicateRule ports org.languagetool.rules.ca.PronomFebleDuplicateRule (simplified).
// Flags weak pronouns both before and after a verb group without whitespace after.
type PronomFebleDuplicateRule struct {
	Messages map[string]string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewPronomFebleDuplicateRule(messages map[string]string) *PronomFebleDuplicateRule {
	r := &PronomFebleDuplicateRule{Messages: messages}
	// Java: S'ha de fer-se → S'ha de fer
	r.AddExamplePair(
		rules.Wrong("<marker>S'ha de fer-se</marker>."),
		rules.Fixed("<marker>S'ha de fer</marker>."),
	)
	return r
}

func (r *PronomFebleDuplicateRule) GetID() string { return "PRONOMS_FEBLES_DUPLICATS" }

func (r *PronomFebleDuplicateRule) GetDescription() string {
	return "Pronoms febles duplicats"
}

// AddExamplePair ports Rule.addExamplePair.
func (r *PronomFebleDuplicateRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *PronomFebleDuplicateRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *PronomFebleDuplicateRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *PronomFebleDuplicateRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	msg := "Combinació incorrecta de pronoms febles. Deixeu els de davant o els de darrere del verb."
	short := "Combinació incorrecta de pronoms febles."

	var before []int // indices of pronouns before verb
	verbStart, verbEnd := -1, -1
	var after []int

	flush := func(end int) {
		if len(before) > 0 && len(after) > 0 && verbStart >= 0 {
			from := tokens[before[0]].GetStartPos()
			to := tokens[end].GetEndPos()
			rm := rules.NewRuleMatch(r, sentence, from, to, msg)
			rm.ShortMessage = short
			// suggest drop before or drop after
			var keepAfter, keepBefore strings.Builder
			for i := before[len(before)-1] + 1; i <= verbEnd; i++ {
				if i > before[len(before)-1]+1 {
					keepAfter.WriteByte(' ')
				}
				keepAfter.WriteString(tokens[i].GetToken())
			}
			for _, j := range after {
				keepAfter.WriteString(tokens[j].GetToken())
			}
			for i := before[0]; i <= verbEnd; i++ {
				if i > before[0] {
					keepBefore.WriteByte(' ')
				}
				keepBefore.WriteString(tokens[i].GetToken())
			}
			s1 := tools.PreserveCase(keepAfter.String(), tokens[before[0]].GetToken())
			s2 := tools.PreserveCase(keepBefore.String(), tokens[before[0]].GetToken())
			rm.SetSuggestedReplacements([]string{s1, s2})
			out = append(out, rm)
		}
		before, after = nil, nil
		verbStart, verbEnd = -1, -1
	}

	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if isPronomFebleToken(tok) {
			if verbStart < 0 {
				// before verb
				if len(before) == 0 || tok.IsWhitespaceBefore() || i == 1 {
					before = append(before, i)
				} else if !tok.IsWhitespaceBefore() && verbEnd >= 0 {
					after = append(after, i)
				} else {
					flush(i - 1)
					before = []int{i}
				}
			} else if !tok.IsWhitespaceBefore() {
				after = append(after, i)
			} else {
				flush(i - 1)
				before = []int{i}
			}
			continue
		}
		// verb-like: Java VERB_CONJUGAT V.[SI].* (GV chunk path incomplete without chunker)
		if isConjugatedVerb(tok) {
			if len(before) > 0 {
				if verbStart < 0 {
					verbStart = i
				}
				verbEnd = i
				continue
			}
		}
		// break group
		if len(before) > 0 || verbStart >= 0 {
			flush(i - 1)
		}
	}
	if len(before) > 0 && len(after) > 0 {
		flush(len(tokens) - 1)
	}
	return out
}

func isPronomFebleToken(tok *languagetool.AnalyzedTokenReadings) bool {
	// Java getLemmaOfPronomFeble: PRONOM_FEBLE on POS only (fail closed without tags).
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if PPronomFeble.MatchString(*r.GetPOSTag()) {
			return true
		}
	}
	return false
}

func isConjugatedVerb(tok *languagetool.AnalyzedTokenReadings) bool {
	// Java VERB_CONJUGAT = V.[SI].* (also GV chunk path incomplete without chunker).
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if verbConjugatRE.MatchString(*r.GetPOSTag()) {
			return true
		}
	}
	return false
}
