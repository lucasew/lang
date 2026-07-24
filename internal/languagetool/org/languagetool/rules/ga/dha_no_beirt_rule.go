package ga

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/people.txt
var peopleFS embed.FS

var (
	peopleOnce sync.Once
	peopleSet  map[string]bool
)

var numberReplacements = map[string]string{
	"dhá":      "beirt",
	"trí":      "triúr",
	"ceathair": "ceathrar",
	"cúig":     "cúigear",
	"sé":       "seisear",
	"seacht":   "seachtar",
	"ocht":     "ochtar",
	"naoi":     "naonúr",
	"deich":    "deichniúr",
}

func loadPeople() map[string]bool {
	peopleOnce.Do(func() {
		f, err := peopleFS.Open("data/people.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m := map[string]bool{}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := tools.JavaStringTrim(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			if line[0] == '*' {
				m[line[1:]] = true
			} else {
				m[strings.ToLower(line)] = true
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		peopleSet = m
	})
	return peopleSet
}

// DhaNoBeirtRule ports org.languagetool.rules.ga.DhaNoBeirtRule.
// isPerson: surface toLower in people.txt OR lemma readings (Java) — no de-lenition invent.
type DhaNoBeirtRule struct {
	messages map[string]string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewDhaNoBeirtRule(messages map[string]string) *DhaNoBeirtRule {
	_ = loadPeople()
	r := &DhaNoBeirtRule{messages: messages}
	// Java: dhá → beirt
	r.AddExamplePair(
		rules.Wrong("Tá <marker>dhá</marker> dheartháireacha agam."),
		rules.Fixed("Tá <marker>beirt</marker> dheartháireacha agam."),
	)
	return r
}

func (r *DhaNoBeirtRule) GetID() string { return "GA_DHA_NO_BEIRT" }

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

func (r *DhaNoBeirtRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		if !isNumber(tokens[i].GetToken()) || i+1 >= len(tokens) || !isPerson(tokens[i+1]) {
			continue
		}
		// dhá … déag → dháréag + delete déag
		if strings.EqualFold(tokens[i].GetToken(), "dhá") {
			foundDeag := false
			for j := i + 2; j < len(tokens); j++ {
				if strings.EqualFold(tokens[j].GetToken(), "déag") {
					rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(),
						"Ba chóir duit <suggestion>dháréag</suggestion> a scríobh")
					rm.ShortMessage = "Uimhir phearsanta"
					rm.SetSuggestedReplacement("dháréag")
					ruleMatches = append(ruleMatches, rm)
					rm2 := rules.NewRuleMatch(r, sentence, tokens[j].GetStartPos(), tokens[j].GetEndPos(),
						"Ba chóir duit \"déag\" a scriosadh.")
					rm2.ShortMessage = "Uimhir phearsanta"
					ruleMatches = append(ruleMatches, rm2)
					foundDeag = true
					break
				}
			}
			if foundDeag {
				continue
			}
		}
		replacement := numberReplacements[strings.ToLower(tokens[i].GetToken())]
		if replacement == "" {
			continue
		}
		msg := "Ba chóir duit <suggestion>" + replacement + " " + tokens[i+1].GetToken() + "</suggestion> a scríobh"
		rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
		rm.ShortMessage = "Uimhir phearsanta"
		rm.SetSuggestedReplacement(replacement + " " + tokens[i+1].GetToken())
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

func isNumber(tok string) bool {
	_, ok := numberReplacements[strings.ToLower(tok)]
	return ok
}

// isPerson ports DhaNoBeirtRule.isPerson:
// 1) surface toLower in people.txt
// 2) else any reading lemma in people.txt
// No de-lenition or substring invent (Java has neither).
func isPerson(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	people := loadPeople()
	if people[strings.ToLower(tok.GetToken())] {
		return true
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetLemma() == nil {
			continue
		}
		lem := *r.GetLemma()
		if people[lem] || people[strings.ToLower(lem)] {
			return true
		}
	}
	return false
}
