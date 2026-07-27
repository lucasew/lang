package ga

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DhaNoBeirtRuleID ports DhaNoBeirtRule.getId() → "GA_DHA_NO_BEIRT".
const DhaNoBeirtRuleID = "GA_DHA_NO_BEIRT"

// DhaNoBeirtRule ports org.languagetool.rules.ga.DhaNoBeirtRule.
//
// Java field twin summary:
//   - id GA_DHA_NO_BEIRT
//   - description "'dhá' nó 'beirt'"
//   - Categories.MISC + ITSIssueType.Misspelling
//   - example: dhá → beirt
//   - data from DhaNoBeirtData (people.txt + number map)
//
// Match control flow is bug-for-bug with Java (sticky replacement, length-2 gate,
// exact-case map get, dual déag matches without break, prevTokenIndex span).
type DhaNoBeirtRule struct {
	messages map[string]string
	// Category ports super.setCategory(Categories.MISC.getCategory(messages)).
	Category *rules.Category
	// IssueType ports setLocQualityIssueType(ITSIssueType.Misspelling).
	IssueType rules.ITSIssueType
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewDhaNoBeirtRule ports DhaNoBeirtRule(ResourceBundle messages).
func NewDhaNoBeirtRule(messages map[string]string) *DhaNoBeirtRule {
	// Eager-load people.txt like Java static DhaNoBeirtData.daoine.
	_ = getDaoine()
	r := &DhaNoBeirtRule{
		messages:  messages,
		Category:  rules.CatMisc.GetCategory(messages),
		IssueType: rules.ITSMisspelling,
	}
	// Java: addExamplePair(wrong dhá, fixed beirt)
	r.AddExamplePair(
		rules.Wrong("Tá <marker>dhá</marker> dheartháireacha agam."),
		rules.Fixed("Tá <marker>beirt</marker> dheartháireacha agam."),
	)
	return r
}

// GetID ports DhaNoBeirtRule.getId().
func (r *DhaNoBeirtRule) GetID() string { return DhaNoBeirtRuleID }

// GetDescription ports DhaNoBeirtRule.getDescription().
func (r *DhaNoBeirtRule) GetDescription() string { return "'dhá' nó 'beirt'" }

// GetCategory ports Rule.getCategory.
func (r *DhaNoBeirtRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *DhaNoBeirtRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSMisspelling
	}
	return r.IssueType
}

// AddExamplePair ports Rule.addExamplePair.
func (r *DhaNoBeirtRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *DhaNoBeirtRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *DhaNoBeirtRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports DhaNoBeirtRule.match — bug-for-bug Java control flow:
//  1. i from 1; prevTokenIndex starts 0; end of iter prevTokenIndex = i
//  2. gate: isNumber(tokens[i]) && (i < len-2 && isPerson(tokens[i+1]))
//  3. dhá + later déag: two matches immediately; msg=null; replacement="dháréag"
//  4. simple replace only when replacement == null; map get exact token case
//  5. when msg != null, span is tokens[prevTokenIndex+1], short "Uimhir phearsanta"
//  6. replacement never cleared after a successful non-null assignment (sticky)
func (r *DhaNoBeirtRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	markDeag := 1
	prevTokenIndex := 0
	// Java String replacement = null / msg = null — use pointers for null fidelity.
	var replacement *string
	var msg *string
	for i := 1; i < len(tokens); i++ {
		// Java: isNumber(tokens[i]) && (i < tokens.length - 2 && isPerson(tokens[i + 1]))
		// Requires a token after the person (not merely person present).
		if isNumber(tokens[i]) && (i < len(tokens)-2 && isPerson(tokens[i+1])) {
			// Java: "dhá".equalsIgnoreCase(...) && (i < tokens.length - 2)
			// Second length check is redundant given outer gate but kept for twin.
			if strings.EqualFold(tokens[i].GetToken(), "dhá") && (i < len(tokens)-2) {
				// No break: every déag hit adds two matches (Java for-loop continues).
				for j := i + 2; j < len(tokens); j++ {
					if strings.EqualFold(tokens[j].GetToken(), "déag") {
						markDeag = j
						rep := "dháréag"
						replacement = &rep
						m1 := "Ba chóir duit <suggestion>" + rep + "</suggestion> a scríobh"
						msg = &m1
						// Java span: tokens[prevTokenIndex+1] (equals number token under normal loop)
						match := rules.NewRuleMatch(r, sentence,
							tokens[prevTokenIndex+1].GetStartPos(), tokens[prevTokenIndex+1].GetEndPos(), m1)
						match.ShortMessage = "Uimhir phearsanta"
						ruleMatches = append(ruleMatches, match)
						m2 := "Ba chóir duit \"déag\" a scriosadh."
						msg = &m2
						match2 := rules.NewRuleMatch(r, sentence,
							tokens[markDeag].GetStartPos(), tokens[markDeag].GetEndPos(), m2)
						match2.ShortMessage = "Uimhir phearsanta"
						ruleMatches = append(ruleMatches, match2)
						msg = nil
					}
				}
			}
			// Java: if (replacement == null) { replacement = map.get(exact token); ... }
			if replacement == nil {
				// Exact-case get (keys are lowercase). Miss → null (sticky stays null).
				if rep, ok := getNumberReplacements()[tokens[i].GetToken()]; ok {
					replacement = &rep
					if msg == nil {
						// Java string concat of non-null replacement
						m := "Ba chóir duit <suggestion>" + rep + " " + tokens[i+1].GetToken() + "</suggestion> a scríobh"
						msg = &m
					}
				} else if msg == nil {
					// Java: "..." + null + "..." → literal "null" in message
					m := "Ba chóir duit <suggestion>null " + tokens[i+1].GetToken() + "</suggestion> a scríobh"
					msg = &m
				}
			}
		}
		if msg != nil {
			match := rules.NewRuleMatch(r, sentence,
				tokens[prevTokenIndex+1].GetStartPos(), tokens[prevTokenIndex+1].GetEndPos(), *msg)
			match.ShortMessage = "Uimhir phearsanta"
			ruleMatches = append(ruleMatches, match)
			msg = nil
		}
		prevTokenIndex = i
	}
	return ruleMatches
}

// isNumber ports DhaNoBeirtRule.isNumber: keySet equalsIgnoreCase on token surface.
func isNumber(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	surface := tok.GetToken()
	for num := range getNumberReplacements() {
		if strings.EqualFold(num, surface) {
			return true
		}
	}
	return false
}

// isPerson ports DhaNoBeirtRule.isPerson:
//  1. surface toLowerCase in daoine set
//  2. else any reading lemma exact contains (Java does not toLower the lemma)
// No de-lenition invent.
func isPerson(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	daoine := getDaoine()
	// Java: getDaoine().contains(tok.getToken().toLowerCase())
	if daoine[strings.ToLower(tok.GetToken())] {
		return true
	}
	// Java: getDaoine().contains(reading.getLemma()) — exact, no toLower
	for _, reading := range tok.GetReadings() {
		if reading == nil {
			continue
		}
		lem := reading.GetLemma()
		if lem == nil {
			// Java Set.contains(null) → false
			continue
		}
		if daoine[*lem] {
			return true
		}
	}
	return false
}
