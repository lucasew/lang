package rules

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// CleanOverlappingFilter ports org.languagetool.rules.CleanOverlappingFilter.
// Removes overlapping matches according to rule priorities.
// Input must be ordered by start position.
type CleanOverlappingFilter struct {
	// PriorityForID returns language priority for a rule id (like Language.getRulePriority).
	PriorityForID func(id string) int
	// HidePremiumMatches when true demotes premium rules to min priority.
	HidePremiumMatches bool
	// IsPremiumRule identifies premium rules (overridable; default: id contains "PREMIUM").
	IsPremiumRule func(rm *RuleMatch) bool
}

const negativeConstant = math.MinInt32 + 10000

func NewCleanOverlappingFilter(priorityForID func(string) int, hidePremium bool) *CleanOverlappingFilter {
	f := &CleanOverlappingFilter{
		PriorityForID:      priorityForID,
		HidePremiumMatches: hidePremium,
	}
	f.IsPremiumRule = func(rm *RuleMatch) bool {
		return strings.Contains(ruleIDOf(rm.Rule), "PREMIUM")
	}
	return f
}

func (f *CleanOverlappingFilter) Filter(ruleMatches []*RuleMatch) []*RuleMatch {
	var cleanList []*RuleMatch
	var prev *RuleMatch
	for _, ruleMatch := range ruleMatches {
		if prev == nil {
			prev = ruleMatch
			continue
		}
		if ruleMatch.FromPos < prev.FromPos {
			panic(fmt.Sprintf("The list of rule matches is not ordered. Make sure it is sorted by start position. RuleMatch from=%d; previous from=%d",
				ruleMatch.FromPos, prev.FromPos))
		}

		isDuplicateSuggestion := false
		if len(ruleMatch.SuggestedReplacements) > 0 && len(prev.SuggestedReplacements) > 0 {
			suggestion := ruleMatch.SuggestedReplacements[0]
			prevSuggestion := prev.SuggestedReplacements[0]
			if ruleMatch.FromPos == prev.ToPos {
				if strings.HasSuffix(prevSuggestion, ",") && strings.HasPrefix(suggestion, ", ") {
					isDuplicateSuggestion = true
				}
			}
			// Java: indexOf(" ") > 0 — space must not be at index 0
			if strings.Index(suggestion, " ") > 0 && strings.Index(prevSuggestion, " ") > 0 &&
				ruleMatch.FromPos == prev.ToPos+1 {
				parts := strings.Split(suggestion, " ")
				partsPrev := strings.Split(prevSuggestion, " ")
				if len(partsPrev) > 1 && len(parts) > 1 && partsPrev[1] == parts[0] {
					isDuplicateSuggestion = true
				}
			}
		}

		// no overlapping (juxtaposed errors are not removed)
		if ruleMatch.FromPos >= prev.ToPos && !isDuplicateSuggestion {
			cleanList = append(cleanList, prev)
			prev = ruleMatch
			continue
		}
		// overlapping
		currentPriority := f.priorityOf(ruleMatch)
		prevPriority := f.priorityOf(prev)

		currentIsPunctuationOnly := f.isPunctuationOnlyChange(ruleMatch)
		prevIsPunctuationOnly := f.isPunctuationOnlyChange(prev)
		if currentIsPunctuationOnly && prevIsPunctuationOnly {
			curAll := includedAllAtOnce(ruleMatch.Rule)
			prevAll := includedAllAtOnce(prev.Rule)
			if curAll != prevAll {
				if curAll {
					if currentPriority < prevPriority {
						currentPriority = prevPriority + 1
					}
				} else if prevPriority < currentPriority {
					prevPriority = currentPriority + 1
				}
			}
		}
		if currentPriority == prevPriority {
			// take the longest error
			currentPriority = ruleMatch.ToPos - ruleMatch.FromPos
			prevPriority = prev.ToPos - prev.FromPos
		}
		if currentPriority == prevPriority {
			currentPriority++ // take the last one
		}
		if currentPriority > prevPriority {
			prev = ruleMatch
		}
	}
	if prev != nil {
		cleanList = append(cleanList, prev)
	}
	return cleanList
}

func (f *CleanOverlappingFilter) priorityOf(rm *RuleMatch) int {
	id := ruleIDOf(rm.Rule)
	p := 0
	if f.PriorityForID != nil {
		p = f.PriorityForID(id)
	}
	if f.IsPremiumRule != nil && f.IsPremiumRule(rm) && f.HidePremiumMatches {
		p = math.MinInt32
	}
	if hasPickyTag(rm.Rule) && p != math.MinInt32 {
		p += negativeConstant
	}
	return p
}

func hasPickyTag(rule any) bool {
	if r, ok := rule.(interface{ HasTag(Tag) bool }); ok {
		return r.HasTag(TagPicky)
	}
	if r, ok := rule.(RuleWithTags); ok {
		for _, t := range r.GetTags() {
			if t == TagPicky {
				return true
			}
		}
	}
	return false
}

func includedAllAtOnce(rule any) bool {
	if r, ok := rule.(interface{ IsIncludedInErrorsCorrectedAllAtOnce() bool }); ok {
		return r.IsIncludedInErrorsCorrectedAllAtOnce()
	}
	return false
}

func (f *CleanOverlappingFilter) isPunctuationOnlyChange(match *RuleMatch) bool {
	if match == nil {
		return false
	}
	suggestions := match.SuggestedReplacements
	if len(suggestions) == 0 {
		return false
	}
	replacement := suggestions[0]
	var original string
	if match.Sentence != nil {
		sentenceStr := match.Sentence.GetText()
		from, to := match.FromPos, match.ToPos
		// UTF-16 positions
		if from > -1 && to > -1 && to <= utf16Len(sentenceStr) && from < to {
			original = utf16Substring(sentenceStr, from, to)
		} else {
			return false
		}
	} else {
		return false
	}
	if replacement == original {
		return false
	}
	return keepLettersAndDigits(original) == keepLettersAndDigits(replacement)
}

func keepLettersAndDigits(s string) string {
	var b strings.Builder
	for _, ch := range s {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}
