package de

import (
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SimilarNameRule ports org.languagetool.rules.de.SimilarNameRule.
// Without a German name tagger, capitalized tokens of length ≥4 are treated as candidate names.
type SimilarNameRule struct {
	Messages map[string]string
}

const (
	similarNameMinLen  = 4
	similarNameMaxDiff = 1
)

func NewSimilarNameRule(messages map[string]string) *SimilarNameRule {
	return &SimilarNameRule{Messages: messages}
}

func (r *SimilarNameRule) GetID() string { return "DE_SIMILAR_NAMES" }

var similarNameExclude = map[string]struct{}{
	"Dein": {}, "Deine": {}, "Deinen": {}, "Deiner": {}, "Deines": {}, "Deinem": {},
	"Ihr": {}, "Ihre": {}, "Ihren": {}, "Ihrer": {}, "Ihres": {}, "Ihrem": {},
}

func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	la, lb := len(ar), len(br)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	// two-row DP
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
			ins := cur[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			cur[j] = min3(ins, del, sub)
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func absInt(x int) int {
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
		// genitive / dative endings
		nameEndsWithS := endsWith(name, "s") && !endsWith(nameHere, "s")
		otherEndsWithS := !endsWith(name, "s") && endsWith(nameHere, "s")
		nameEndsWithN := endsWith(name, "n") && !endsWith(nameHere, "n")
		otherEndsWithN := !endsWith(name, "n") && endsWith(nameHere, "n")
		if nameEndsWithS || otherEndsWithS || nameEndsWithN || otherEndsWithN {
			continue
		}
		lenDiff := absInt(utf8.RuneCountInString(name) - utf8.RuneCountInString(nameHere))
		if lenDiff <= similarNameMaxDiff && levenshtein(name, nameHere) <= similarNameMaxDiff {
			return name
		}
	}
	return ""
}

func endsWith(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}

func isMaybeNameSurface(word string) bool {
	if utf8.RuneCountInString(word) < similarNameMinLen {
		return false
	}
	if _, ok := similarNameExclude[word]; ok {
		return false
	}
	if !tools.StartsWithUppercase(word) {
		return false
	}
	// require mostly letters
	for _, r := range word {
		if !unicode.IsLetter(r) && r != '-' {
			return false
		}
	}
	return true
}

func (r *SimilarNameRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	namesSoFar := map[string]struct{}{}
	var ruleMatches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		for _, token := range sentence.GetTokensWithoutWhitespace() {
			word := token.GetToken()
			if !isMaybeNameSurface(word) {
				continue
			}
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
