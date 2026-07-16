package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ConjunctionAtBeginOfSentenceRule is a surface stand-in for the DE statistic rule:
// when MinPercent is 0 (always), flags sentences starting with listed conjunctions
// (Java uses KON POS tags and percentage thresholds).
type ConjunctionAtBeginOfSentenceRule struct {
	Messages   map[string]string
	MinPercent int // 0 = flag all matching sentence starts
	// conjunctions lowercased
	conjunctions map[string]struct{}
}

func NewConjunctionAtBeginOfSentenceRule(messages map[string]string) *ConjunctionAtBeginOfSentenceRule {
	// from commented fillerWords list in Java source (common sentence-start conjunctions)
	words := []string{
		"aber", "als", "also", "andererseits", "anschließend", "anschliessend", "anstatt",
		"außer", "ausserdem", "bevor", "beziehungsweise", "bis", "da", "dadurch", "dafür",
		"dagegen", "damit", "danach", "dann", "darauf", "darum", "dass", "davor", "dazu",
		"denn", "deshalb", "dessen", "desto", "desungeachtet", "deswegen", "doch", "ehe",
		"eh", "entweder", "falls", "ferner", "folglich", "genauso", "geschweige", "immerhin",
		"indem", "indes", "indessen", "insofern", "insoweit", "inzwischen", "je", "jedoch",
		"nachdem", "ob", "obgleich", "obschon", "obwohl", "obzwar", "oder", "respektive",
		"seit", "seitdem", "so", "sodass", "sofern", "solang", "solange", "sondern", "sooft",
		"soviel", "soweit", "sowie", "sowohl", "später", "statt", "trotzdem", "um", "umso",
		"und", "ungeachtet", "vorher", "während", "währenddem", "währenddessen", "weder",
		"weil", "wenn", "wenngleich", "wennschon", "wie", "wiewohl", "wobei", "wohingegen",
		"zumal", "zuvor", "zwar", "allerdings", "auch",
	}
	m := map[string]struct{}{}
	for _, w := range words {
		m[w] = struct{}{}
	}
	return &ConjunctionAtBeginOfSentenceRule{
		Messages:     messages,
		MinPercent:   0,
		conjunctions: m,
	}
}

func (r *ConjunctionAtBeginOfSentenceRule) GetID() string {
	return "CONJUNCTION_BEGIN_SENTENCE_DE"
}

func (r *ConjunctionAtBeginOfSentenceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r.MinPercent != 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	if len(tokens) < 3 {
		return nil
	}
	num := 1
	// skip opening quote
	if isOpeningQuoteDE(tokens[num].GetToken()) {
		num++
	}
	if num >= len(tokens) {
		return nil
	}
	tok := tokens[num]
	word := tok.GetToken()
	lc := strings.ToLower(word)
	if _, ok := r.conjunctions[lc]; !ok {
		return nil
	}
	// Java exceptions
	if word == "Wie" || word == "Seit" || word == "Allerdings" {
		return nil
	}
	if word == "Aber" && num+1 < len(tokens) && tokens[num+1].GetToken() == "auch" {
		return nil
	}
	if word == "Auch" && num+1 < len(tokens) && tokens[num+1].GetToken() == "wenn" {
		return nil
	}
	if word == "Um" {
		for i := num + 1; i < len(tokens); i++ {
			if tokens[i].GetToken() == "," || tokens[i].GetToken() == "herum" {
				return nil
			}
		}
	}
	if word == "Sondern" {
		return nil
	}
	msg := "Viele Sätze beginnen mit einer Konjunktion. Variieren Sie den Satzanfang."
	rm := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
	rm.ShortMessage = "Konjunktion am Satzanfang"
	return []*rules.RuleMatch{rm}
}

func isOpeningQuoteDE(s string) bool {
	switch s {
	case "\"", "„", "«", "»", "'", "‚", "‘":
		return true
	}
	return false
}
