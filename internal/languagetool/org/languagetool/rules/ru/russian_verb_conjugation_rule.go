package ru

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianVerbConjugationRule ports org.languagetool.rules.ru.RussianVerbConjugationRule.
// Checks personal pronoun + verb agreement (present/future and past).
type RussianVerbConjugationRule struct {
	Messages map[string]string
	// Category ports setCategory(Categories.GRAMMAR).
	Category *rules.Category
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// Java: private static final Pattern PRONOUN / FUT_REAL_VERB / PAST_VERB
// Matcher.find() semantics (no full-string anchors).
var (
	ruPronoun  = regexp.MustCompile(`PNN:(.*):Nom:(.*)`)
	ruFutReal  = regexp.MustCompile(`VB:(Fut|Real):(.*):(.*):(.*):(.*)`)
	ruPastVerb = regexp.MustCompile(`VB:Past:(.*):(.*):(.*)`)
)

func NewRussianVerbConjugationRule(messages map[string]string) *RussianVerbConjugationRule {
	r := &RussianVerbConjugationRule{
		Messages: messages,
		Category: rules.CatGrammar.GetCategory(messages),
	}
	// Java: Я идёт → Я иду
	r.AddExamplePair(
		rules.Wrong("<marker>Я идёт</marker>."),
		rules.Fixed("<marker>Я иду</marker>."),
	)
	return r
}

func (r *RussianVerbConjugationRule) GetID() string { return "RU_VERB_CONJUGATION" }

func (r *RussianVerbConjugationRule) GetDescription() string {
	return "Согласование личных местоимений с глаголами"
}

// GetCategory ports Rule.getCategory (Categories.GRAMMAR).
func (r *RussianVerbConjugationRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// getShort ports RussianVerbConjugationRule.getShort (protected in Java).
func (r *RussianVerbConjugationRule) getShort() string {
	return "Неверное спряжение глагола"
}

// GetShort is the exported alias used by tests / RuleMatch short message.
func (r *RussianVerbConjugationRule) GetShort() string { return r.getShort() }

// AddExamplePair ports Rule.addExamplePair.
func (r *RussianVerbConjugationRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *RussianVerbConjugationRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *RussianVerbConjugationRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports RussianVerbConjugationRule.match.
func (r *RussianVerbConjugationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokenReadings := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	// Java: for (int i = 1; i < tokenReadings.length - 1; i++)
	for i := 1; i < len(tokenReadings)-1; i++ {
		previousReading := tokenReadings[i-1]
		currentReading := tokenReadings[i]
		nextReading := tokenReadings[i+1]
		if previousReading == nil || currentReading == nil || nextReading == nil {
			continue
		}
		prevReadings := previousReading.GetReadings()
		curReadings := currentReading.GetReadings()
		if len(prevReadings) == 0 || len(curReadings) == 0 {
			continue
		}
		// Java: readings.get(0)
		previousLemmaTok := prevReadings[0]
		currentLemmaTok := curReadings[0]
		if previousLemmaTok == nil || currentLemmaTok == nil {
			continue
		}
		previousToken := previousLemmaTok.GetToken()
		currentToken := currentLemmaTok.GetToken()
		currentPosTag := currentLemmaTok.GetPOSTag()
		if currentToken != "" && currentPosTag != nil && *currentPosTag != "" {
			// Java: PRONOUN.matcher(currentPosTag); pronounMatcher.find()
			pm := ruPronoun.FindStringSubmatch(*currentPosTag)
			if pm != nil && previousToken != "и" {
				// Java: pronounPair = (group1, group2)
				pronounLeft, pronounRight := pm[1], pm[2]
				nextReadings := nextReading.GetReadings()
				if len(nextReadings) == 0 || nextReadings[0] == nil {
					continue
				}
				nextLemmaTok := nextReadings[0]
				next2Token := ""
				if i < len(tokenReadings)-2 {
					next2Reading := tokenReadings[i+2]
					if next2Reading != nil {
						n2 := next2Reading.GetReadings()
						if len(n2) > 0 && n2[0] != nil {
							next2Token = n2[0].GetToken()
						}
					}
				}
				nextToken := nextLemmaTok.GetToken()
				nextPosTag := nextLemmaTok.GetPOSTag()
				// Java: nextPosTag != null && !empty && !(next2=="быть" && next=="может") && !(next=="целую")
				if nextPosTag != nil && *nextPosTag != "" &&
					!(next2Token == "быть" && nextToken == "может") &&
					nextToken != "целую" {
					if vm := ruFutReal.FindStringSubmatch(*nextPosTag); vm != nil {
						// Java: verbPair = (group4, group5)
						verbLeft, verbRight := vm[4], vm[5]
						if isConjugationInPresentOrFutureWrong(pronounLeft, pronounRight, verbLeft, verbRight) {
							ruleMatches = append(ruleMatches, r.addRuleMatch(sentence, currentReading, nextReading))
						}
					} else if vm := ruPastVerb.FindStringSubmatch(*nextPosTag); vm != nil {
						// Java: isConjugationInPastWrong(pronounMatcher.group(1), verbMatcher.group(3))
						if isConjugationInPastWrong(pronounLeft, vm[3]) {
							ruleMatches = append(ruleMatches, r.addRuleMatch(sentence, currentReading, nextReading))
						}
					}
				}
			}
		}
	}
	return ruleMatches
}

// addRuleMatch ports private addRuleMatch.
func (r *RussianVerbConjugationRule) addRuleMatch(
	sentence *languagetool.AnalyzedSentence,
	currentReading, nextReading *languagetool.AnalyzedTokenReadings,
) *rules.RuleMatch {
	// Java: new RuleMatch(this, sentence, current.start, next.end, msg, getShort())
	rm := rules.NewRuleMatch(r, sentence, currentReading.GetStartPos(), nextReading.GetEndPos(),
		"Неверное спряжение глагола или неверное местоимение")
	rm.ShortMessage = r.getShort()
	return rm
}

// isConjugationInPresentOrFutureWrong ports private isConjugationInPresentOrFutureWrong(Pair, Pair).
func isConjugationInPresentOrFutureWrong(pronounLeft, pronounRight, verbLeft, verbRight string) bool {
	if pronounRight != verbRight {
		return true
	}
	// Java: Arrays.asList("Masc", "Fem", "Neut").contains(pronoun.getLeft())
	switch pronounLeft {
	case "Masc", "Fem", "Neut":
		return verbLeft == "PL"
	}
	return pronounLeft != verbLeft
}

// isConjugationInPastWrong ports private isConjugationInPastWrong.
func isConjugationInPastWrong(pronoun, verb string) bool {
	if pronoun == "Sin" {
		return verb == "PL" || verb == "Neut"
	}
	return pronoun != verb
}

// Aliases kept for older unit helpers in russian_verb_conjugation_rule_test.go.
func isConjugationPresentFutureWrong(pronLeft, pronRight, verbLeft, verbRight string) bool {
	return isConjugationInPresentOrFutureWrong(pronLeft, pronRight, verbLeft, verbRight)
}

func isConjugationPastWrong(pronoun, verb string) bool {
	return isConjugationInPastWrong(pronoun, verb)
}
