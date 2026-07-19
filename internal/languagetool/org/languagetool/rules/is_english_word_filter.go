package rules

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// EnglishTaggerFunc tags a single word into AnalyzedTokenReadings (pluggable EN tagger).
type EnglishTaggerFunc func(word string) *languagetool.AnalyzedTokenReadings

// IsEnglishWordFilter ports org.languagetool.rules.IsEnglishWordFilter with a pluggable tagger.
type IsEnglishWordFilter struct {
	Tag EnglishTaggerFunc
}

func NewIsEnglishWordFilter(tag EnglishTaggerFunc) *IsEnglishWordFilter {
	return &IsEnglishWordFilter{Tag: tag}
}

var (
	enWordTagMu sync.RWMutex
	defaultENWordTag EnglishTaggerFunc
)

// SetDefaultEnglishWordTagger wires EN tagger for IsEnglishWordFilter
// (Java: Languages.getLanguageForShortCode("en-US").createDefaultTagger()).
func SetDefaultEnglishWordTagger(tag EnglishTaggerFunc) {
	enWordTagMu.Lock()
	defer enWordTagMu.Unlock()
	defaultENWordTag = tag
}

// AcceptRuleMatch keeps the match when all formPositions refer to English-tagged words.
// tokenPositions is used for skip-corrected refs (same as RuleFilterEvaluator).
// Without a tagger: fail-closed (Java returns null when tagger is null).
func (f *IsEnglishWordFilter) AcceptRuleMatch(match *RuleMatch, args map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *RuleMatch {
	tag := EnglishTaggerFunc(nil)
	if f != nil {
		tag = f.Tag
	}
	if tag == nil {
		enWordTagMu.RLock()
		tag = defaultENWordTag
		enWordTagMu.RUnlock()
	}
	if tag == nil {
		return nil
	}
	// use local filter with resolved tag
	f = &IsEnglishWordFilter{Tag: tag}
	formPosStr, ok := args["formPositions"]
	if !ok {
		panic("Missing key 'formPositions'")
	}
	parts := strings.Split(formPosStr, ",")
	var forms []string
	for _, fp := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(fp))
		if err != nil {
			panic(err)
		}
		idx := skipCorrectedRef(tokenPositions, n)
		if idx < 0 || idx >= len(patternTokens) {
			panic(fmt.Sprintf("formPositions out of bounds: %d", n))
		}
		forms = append(forms, patternTokens[idx].GetToken())
	}
	isEnglish := true
	if postagsStr, ok := args["postags"]; ok {
		postags := strings.Split(postagsStr, ",")
		if len(postags) != len(forms) {
			panic("The number of forms and postags has to be the same in disambiguation rule with filter IsEnglishWordFilter.")
		}
		for i := range postags {
			isEnglish = isEnglish && f.wordIsTaggedWith(forms[i], strings.TrimSpace(postags[i]))
		}
	} else {
		for _, form := range forms {
			isEnglish = isEnglish && f.wordIsTagged(form)
		}
	}
	if isEnglish {
		return match
	}
	return nil
}

func (f *IsEnglishWordFilter) wordIsTaggedWith(word, postag string) bool {
	atr := f.Tag(word)
	if atr == nil {
		return false
	}
	return atr.MatchesPosTagRegex(postag)
}

func (f *IsEnglishWordFilter) wordIsTagged(word string) bool {
	atr := f.Tag(word)
	if atr == nil {
		return false
	}
	return atr.IsTagged()
}

// skipCorrectedRef mirrors RuleFilterEvaluator skip correction (1-based ref).
func skipCorrectedRef(tokenPositions []int, refNumber int) int {
	if len(tokenPositions) == 0 {
		return refNumber - 1
	}
	correctedRef := 0
	i := 0
	for _, tokenPosition := range tokenPositions {
		if i >= refNumber {
			break
		}
		i++
		correctedRef += tokenPosition
	}
	return correctedRef - 1
}
