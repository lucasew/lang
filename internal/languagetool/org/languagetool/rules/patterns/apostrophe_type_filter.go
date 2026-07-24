package patterns

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ApostropheTypeFilter ports org.languagetool.rules.patterns.ApostropheTypeFilter.
type ApostropheTypeFilter struct{}

func (ApostropheTypeFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	wordFrom := GetRequired("wordFrom", arguments)
	hasTypographical := strings.EqualFold(GetRequired("hasTypographicalApostrophe", arguments), "true")
	if wordFrom == "" {
		return nil
	}
	posWord := 0
	if wordFrom == "marker" {
		for posWord < len(patternTokens) && patternTokens[posWord].GetStartPos() < match.GetFromPos() {
			posWord++
		}
		posWord++
	} else {
		var err error
		posWord, err = strconv.Atoi(wordFrom)
		if err != nil {
			panic(err)
		}
	}
	if posWord < 1 || posWord > len(patternTokens) {
		panic(fmt.Sprintf("ApostropheTypeFilter: Index out of bounds, wordFrom: %d", posWord))
	}
	atrWord := patternTokens[posWord-1]
	if hasTypographical == atrWord.HasTypographicApostrophe() {
		return match
	}
	return nil
}
