package uk

import (
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// AllowFullTagMatch ports doGuessCompoundTag guards immediately before tagMatch:
//  - left has "pron" without "numr" → no
//  - right is part|conj|…:pron (unless both numr) and left≠right → no
//  - left empty dict → no
//  - left length ≤ 2 and not intj → no (unless intj)
//  - solid no-dash already tagged → no (caller may also use DynamicNoDashSolidHasTags)
func AllowFullTagMatch(token string, tagWord func(string) []tagging.TaggedWord) bool {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return false
	}
	t := normalizeDash(token)
	if strings.Count(t, "-") != 1 {
		// multi-dash handled elsewhere
		return true
	}
	i := strings.Index(t, "-")
	left, right := t[:i], t[i+1:]
	if left == "" || right == "" {
		return false
	}

	leftTags := tagAsIsAndWithLowerCase(left, tagWord)
	rightTags := tagEitherCase(right, tagWord)
	if len(rightTags) == 0 {
		return false
	}

	// left pron without numr → null
	if hasPosPartInTags(leftTags, "pron") && !hasPosPartInTags(leftTags, "numr") {
		return false
	}

	// right part|conj|…:pron (and left≠right) unless both numr
	if !strings.EqualFold(left, right) && hasRightPartConjOrPron(rightTags) {
		if !(hasPosStartInTags(leftTags, "numr") && hasPosStartInTags(rightTags, "numr")) {
			return false
		}
	}

	// noDash solid already tagged → skip tagMatch (Java)
	if DynamicNoDashSolidHasTags(token, tagWord) {
		return false
	}

	// upper-right non-noun block
	if DynamicUpperRightBlocks(token, tagWord) {
		return false
	}

	if len(leftTags) == 0 {
		return false
	}
	hasIntj := hasPosStartInTags(leftTags, "intj")
	if utf8.RuneCountInString(left) <= 2 && !hasIntj {
		return false
	}
	return true
}

func hasPosPartInTags(tags []tagging.TaggedWord, part string) bool {
	for _, tw := range tags {
		if strings.Contains(tw.PosTag, part) {
			return true
		}
	}
	return false
}

func hasPosStartInTags(tags []tagging.TaggedWord, prefix string) bool {
	for _, tw := range tags {
		if strings.HasPrefix(tw.PosTag, prefix) {
			return true
		}
	}
	return false
}

// hasRightPartConjOrPron ports Pattern.compile("(part|conj).*|.*?:pron.*")
func hasRightPartConjOrPron(tags []tagging.TaggedWord) bool {
	for _, tw := range tags {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		if strings.HasPrefix(pos, "part") || strings.HasPrefix(pos, "conj") {
			return true
		}
		if strings.Contains(pos, ":pron") {
			return true
		}
	}
	return false
}

// DynamicFinalOAdjReadings ports final tryOWithAdj after failed tagMatch in doGuessCompoundTag.
func DynamicFinalOAdjReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	t := normalizeDash(token)
	if strings.Count(t, "-") != 1 {
		return nil
	}
	i := strings.LastIndex(t, "-")
	left, right := t[:i], t[i+1:]
	if left == "" || right == "" {
		return nil
	}
	// already handled as dash prefix / invalid
	if IsDashPrefix(left) || IsDashPrefixInvalid(left) {
		return nil
	}
	rightTags := tagEitherCase(right, tagWord)
	if len(rightTags) == 0 {
		return nil
	}
	return tryOWithAdjTokens(token, left, taggedWordsToTokens(right, rightTags), tagWord)
}
