package ga

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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
			line := strings.TrimSpace(sc.Text())
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
// Surface person matching (no Irish tagger lemmas).
type DhaNoBeirtRule struct {
	messages map[string]string
}

func NewDhaNoBeirtRule(messages map[string]string) *DhaNoBeirtRule {
	_ = loadPeople()
	return &DhaNoBeirtRule{messages: messages}
}

func (r *DhaNoBeirtRule) GetID() string { return "GA_DHA_NO_BEIRT" }

func (r *DhaNoBeirtRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		if !isNumber(tokens[i].GetToken()) || i+1 >= len(tokens) || !isPerson(tokens[i+1].GetToken()) {
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

func isPerson(tok string) bool {
	people := loadPeople()
	for _, cand := range personCandidates(strings.ToLower(tok)) {
		if people[cand] {
			return true
		}
		for p := range people {
			if len(p) >= 3 && (strings.HasPrefix(cand, p) || strings.Contains(cand, p)) {
				return true
			}
		}
	}
	return false
}

// personCandidates yields the surface form plus a simple de-lenited form (C+h → C).
func personCandidates(t string) []string {
	out := []string{t}
	r := []rune(t)
	if len(r) > 2 && r[1] == 'h' {
		// dheartháireacha → deartháireacha
		out = append(out, string(r[0])+string(r[2:]))
	}
	return out
}
