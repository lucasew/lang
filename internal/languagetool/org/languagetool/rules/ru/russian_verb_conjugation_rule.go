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
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewRussianVerbConjugationRule(messages map[string]string) *RussianVerbConjugationRule {
	r := &RussianVerbConjugationRule{Messages: messages}
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

func (r *RussianVerbConjugationRule) GetShort() string {
	return "Неверное спряжение глагола"
}

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

var (
	ruPronoun   = regexp.MustCompile(`^PNN:(.*):Nom:(.*)$`)
	ruFutReal   = regexp.MustCompile(`^VB:(Fut|Real):(.*):(.*):(.*):(.*)$`)
	ruPastVerb  = regexp.MustCompile(`^VB:Past:(.*):(.*):(.*)$`)
)

func (r *RussianVerbConjugationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	for i := 1; i < len(tokens)-1; i++ {
		prev := tokens[i-1]
		cur := tokens[i]
		next := tokens[i+1]
		if cur == nil || next == nil {
			continue
		}
		readings := cur.GetReadings()
		if len(readings) == 0 {
			continue
		}
		curTok := readings[0]
		pos := curTok.GetPOSTag()
		if pos == nil || *pos == "" {
			continue
		}
		pm := ruPronoun.FindStringSubmatch(*pos)
		if pm == nil {
			continue
		}
		if prev != nil && prev.GetToken() == "и" {
			continue
		}
		next2 := ""
		if i+2 < len(tokens) && tokens[i+2] != nil {
			next2 = tokens[i+2].GetToken()
		}
		nreadings := next.GetReadings()
		if len(nreadings) == 0 {
			continue
		}
		nTok := nreadings[0]
		nPos := nTok.GetPOSTag()
		if nPos == nil || *nPos == "" {
			continue
		}
		if next2 == "быть" && next.GetToken() == "может" {
			continue
		}
		if next.GetToken() == "целую" {
			continue
		}
		// pronoun groups: 1=gender/person slot (Masc|Fem|Neut|P1|P2|P3…), 2=Sin|PL
		pronLeft, pronRight := pm[1], pm[2]
		if vm := ruFutReal.FindStringSubmatch(*nPos); vm != nil {
			// groups: 1 Fut|Real, 2.., 4 person/num left, 5 Sin|PL right
			verbLeft, verbRight := vm[4], vm[5]
			if isConjugationPresentFutureWrong(pronLeft, pronRight, verbLeft, verbRight) {
				out = append(out, r.newMatch(sentence, cur, next))
			}
		} else if vm := ruPastVerb.FindStringSubmatch(*nPos); vm != nil {
			// groups: 1.., 3 gender/number of past
			if isConjugationPastWrong(pronLeft, vm[3]) {
				out = append(out, r.newMatch(sentence, cur, next))
			}
		}
	}
	return out
}

func (r *RussianVerbConjugationRule) newMatch(
	sentence *languagetool.AnalyzedSentence,
	cur, next *languagetool.AnalyzedTokenReadings,
) *rules.RuleMatch {
	rm := rules.NewRuleMatch(r, sentence, cur.GetStartPos(), next.GetEndPos(),
		"Неверное спряжение глагола или неверное местоимение")
	rm.ShortMessage = r.GetShort()
	return rm
}

func isConjugationPresentFutureWrong(pronLeft, pronRight, verbLeft, verbRight string) bool {
	if pronRight != verbRight {
		return true
	}
	switch pronLeft {
	case "Masc", "Fem", "Neut":
		return verbLeft == "PL"
	}
	return pronLeft != verbLeft
}

func isConjugationPastWrong(pronoun, verb string) bool {
	if pronoun == "Sin" {
		return verb == "PL" || verb == "Neut"
	}
	return pronoun != verb
}
