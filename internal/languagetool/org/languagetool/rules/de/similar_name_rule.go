package de

import (
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SimilarNameRule ports org.languagetool.rules.de.SimilarNameRule (default off).
// Java: (EIG: && !:COU) || isPosTagUnknown — no letter-class invent.
// Java: TYPOS, setDefaultOff().
type SimilarNameRule struct {
	Messages   map[string]string
	Category   *rules.Category
	DefaultOff bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

const (
	similarNameMinLen  = 4
	similarNameMaxDiff = 1
)

func NewSimilarNameRule(messages map[string]string) *SimilarNameRule {
	r := &SimilarNameRule{
		Messages:   messages,
		Category:   rules.CatTypos.GetCategory(messages),
		DefaultOff: true,
	}
	// Java: Miller → Müller
	r.AddExamplePair(
		rules.Wrong("Angela Müller ist CEO. <marker>Miller</marker> wurde in Hamburg geboren."),
		rules.Fixed("Angela Müller ist CEO. <marker>Müller</marker> wurde in Hamburg geboren."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *SimilarNameRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *SimilarNameRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *SimilarNameRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *SimilarNameRule) GetID() string { return "DE_SIMILAR_NAMES" }

func (r *SimilarNameRule) GetDescription() string {
	return "Mögliche Tippfehler in Namen finden"
}

func (r *SimilarNameRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *SimilarNameRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// MinToCheckParagraph ports SimilarNameRule.minToCheckParagraph (Java returns -1).
func (r *SimilarNameRule) MinToCheckParagraph() int { return -1 }

var similarNameExclude = map[string]struct{}{
	"Dein": {}, "Deine": {}, "Deinen": {}, "Deiner": {}, "Deines": {}, "Deinem": {},
	"Ihr": {}, "Ihre": {}, "Ihren": {}, "Ihrer": {}, "Ihres": {}, "Ihrem": {},
}

func levenshteinSimilar(a, b string) int {
	ar, br := []rune(a), []rune(b)
	la, lb := len(ar), len(br)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			ins, del, sub := cur[j-1]+1, prev[j]+1, prev[j-1]+cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}

func absIntSimilar(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (r *SimilarNameRule) similarName(nameHere string, namesSoFar map[string]struct{}) string {
	for name := range namesSoFar {
		if name == nameHere {
			continue
		}
		nameEndsWithS := endsWithStr(name, "s") && !endsWithStr(nameHere, "s")
		otherEndsWithS := !endsWithStr(name, "s") && endsWithStr(nameHere, "s")
		nameEndsWithN := endsWithStr(name, "n") && !endsWithStr(nameHere, "n")
		otherEndsWithN := !endsWithStr(name, "n") && endsWithStr(nameHere, "n")
		if nameEndsWithS || otherEndsWithS || nameEndsWithN || otherEndsWithN {
			continue
		}
		lenDiff := absIntSimilar(utf8.RuneCountInString(name) - utf8.RuneCountInString(nameHere))
		if lenDiff <= similarNameMaxDiff && levenshteinSimilar(name, nameHere) <= similarNameMaxDiff {
			return name
		}
	}
	return ""
}

func endsWithStr(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}

func isMaybeName(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	word := token.GetToken()
	if utf8.RuneCountInString(word) < similarNameMinLen {
		return false
	}
	if _, ok := similarNameExclude[word]; ok {
		return false
	}
	if !tools.StartsWithUppercase(word) {
		return false
	}
	// Java: (EIG: && !:COU) || isPosTagUnknown
	// isPosTagUnknown is not the same as !isTagged (multi-reading untagged ≠ unknown).
	if token.HasPartialPosTag("EIG:") && !token.HasPartialPosTag(":COU") {
		return true
	}
	return token.IsPosTagUnknown()
}

func (r *SimilarNameRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	namesSoFar := map[string]struct{}{}
	var ruleMatches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		for _, token := range sentence.GetTokensWithoutWhitespace() {
			if !isMaybeName(token) {
				continue
			}
			word := token.GetToken()
			if similar := r.similarName(word, namesSoFar); similar != "" {
				msg := "'" + word + "' ähnelt dem vorher benutzten '" + similar + "', handelt es sich evtl. um einen Tippfehler?"
				rm := rules.NewRuleMatch(r, sentence, pos+token.GetStartPos(), pos+token.GetEndPos(), msg)
				rm.SetSuggestedReplacement(similar)
				ruleMatches = append(ruleMatches, rm)
			}
			namesSoFar[word] = struct{}{}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
