package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// leftPivNvRE ports noun:inanim:p:v_...:nv.* (Matcher.matches on left of півгодини-годину).
var leftPivNvRE = regexp.MustCompile(`^noun:inanim:p:v_...:nv.*$`)

// DynamicPivNvDualReadings ports CompoundTagger півгодини-годину:
// word starts with "пів", left tagged noun:inanim:p:v_…:nv, right noun:inanim → force :p: gender.
func DynamicPivNvDualReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	// Java word.startsWith("пів") is case-sensitive
	surface := normalizeDash(token)
	if !strings.HasPrefix(surface, "пів") {
		return nil
	}
	if strings.Count(surface, "-") != 1 {
		return nil
	}
	i := strings.Index(surface, "-")
	left, right := surface[:i], surface[i+1:]
	if left == "" || right == "" {
		return nil
	}

	leftTags := tagAsIsAndWithLowerCase(left, tagWord)
	if !HasPosTag2(leftTags, leftPivNvRE) {
		return nil
	}
	rightTags := tagEitherCase(right, tagWord)
	if len(rightTags) == 0 {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, tw := range rightTags {
		if !strings.HasPrefix(tw.PosTag, "noun:inanim:") {
			continue
		}
		// replace first :m: / :f: / :n: with :p:
		pos := tw.PosTag
		pos = strings.Replace(pos, ":m:", ":p:", 1)
		if pos == tw.PosTag {
			pos = strings.Replace(pos, ":f:", ":p:", 1)
		}
		if pos == tw.PosTag {
			pos = strings.Replace(pos, ":n:", ":p:", 1)
		}
		// Java: lemma = word (full surface)
		p, l := pos, surface
		out = append(out, languagetool.NewAnalyzedToken(token, &p, &l))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// tagAsIsAndWithLowerCase ports CompoundTagger.tagAsIsAndWithLowerCase.
func tagAsIsAndWithLowerCase(word string, tagWord func(string) []tagging.TaggedWord) []tagging.TaggedWord {
	if tagWord == nil || word == "" {
		return nil
	}
	var out []tagging.TaggedWord
	seen := map[string]struct{}{}
	add := func(w string) {
		for _, tw := range tagWord(w) {
			key := tw.Lemma + "|" + tw.PosTag
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, tw)
		}
	}
	add(word)
	if low := strings.ToLower(word); low != word {
		add(low)
	}
	return out
}

// DynamicUpperRightCompoundReadings ports doGuessCompoundTag upper-right arm
// (after dashPrefixMatch fails): lower adj:bad compound; tryOWithAdj; noun-noun flow.
// Only when right side is capitalized.
func DynamicUpperRightCompoundReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	t := normalizeDash(token)
	if strings.Count(t, "-") != 1 || strings.HasPrefix(t, "-") || strings.HasSuffix(t, "-") {
		return nil
	}
	i := strings.LastIndex(t, "-")
	left, right := t[:i], t[i+1:]
	if left == "" || right == "" {
		return nil
	}
	rr := []rune(right)
	if !unicode.IsUpper(rr[0]) {
		return nil
	}
	// пів- upper already handled by FixedPartReadings
	if strings.HasPrefix(t, "пів-") || strings.HasPrefix(strings.ToLower(t), "пів-") {
		return nil
	}
	// skip if this is a known dash prefix left (handled elsewhere)
	if IsDashPrefix(left) || IsDashPrefixInvalid(left) {
		return nil
	}

	// Кримсько-Татарський — full lower compound has adj…:bad
	lowerCompound := tagAsIsAndWithLowerCase(strings.ToLower(t), tagWord)
	for _, tw := range lowerCompound {
		if strings.HasPrefix(tw.PosTag, "adj") && strings.Contains(tw.PosTag, "bad") {
			// return as surface token with those tags
			return taggedWordsToSurfaceTokens(token, lowerCompound)
		}
	}

	// tryOWithAdj when right capitalized OR left ends with о OR right is adj
	rightTags2 := tagAsIsAndWithLowerCase(right, tagWord)
	rightToks2 := taggedWordsToTokens(right, rightTags2)
	useOAdj := toolsIsCapitalized(right) || strings.HasSuffix(left, "о") || hasAdjTag(rightTags2)
	if useOAdj {
		if match := tryOWithAdjTokens(token, left, rightToks2, tagWord); len(match) > 0 {
			return match
		}
	}

	// Жінка-Актриса: both non-prop nouns → allow fall-through (return nil so FullTagMatch can run)
	// other cases with upper right → fail closed (Java return null)
	leftTags := tagAsIsAndWithLowerCase(left, tagWord)
	rightTags := tagEitherCase(right, tagWord)
	if hasNonPropNoun(leftTags) && hasNonPropNoun(rightTags) {
		// flow-through: let FullTagMatch handle
		return nil
	}
	// signal "blocked" upper-right non-noun: use empty special? Java returns null meaning untagged.
	// We return a sentinel only when we should suppress further tagMatch — hard in layered Tag().
	// Caller runs this before FullTagMatch; returning nil allows FullTagMatch. Matching Java
	// "return null" for non-noun upper pairs means we should suppress FullTagMatch.
	// Use empty non-nil? Go convention: use DynamicUpperRightBlocks.
	return nil
}

// DynamicUpperRightBlocks reports whether Java would return null for upper-right compounds
// (suppress further compound tagging).
func DynamicUpperRightBlocks(token string, tagWord func(string) []tagging.TaggedWord) bool {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return false
	}
	t := normalizeDash(token)
	if strings.Count(t, "-") != 1 {
		return false
	}
	i := strings.LastIndex(t, "-")
	left, right := t[:i], t[i+1:]
	if left == "" || right == "" {
		return false
	}
	rr := []rune(right)
	if !unicode.IsUpper(rr[0]) {
		return false
	}
	if strings.HasPrefix(t, "пів-") || IsDashPrefix(left) || IsDashPrefixInvalid(left) {
		return false
	}
	// if lower adj:bad or oAdj would match, not a block
	if rs := DynamicUpperRightCompoundReadings(token, tagWord); len(rs) > 0 {
		return false
	}
	leftTags := tagAsIsAndWithLowerCase(left, tagWord)
	rightTags := tagEitherCase(right, tagWord)
	if hasNonPropNoun(leftTags) && hasNonPropNoun(rightTags) {
		return false // flow-through
	}
	// upper right and not dual non-prop noun → block
	return true
}

func hasAdjTag(tags []tagging.TaggedWord) bool {
	for _, tw := range tags {
		if strings.HasPrefix(tw.PosTag, "adj") {
			return true
		}
	}
	return false
}

func hasNonPropNoun(tags []tagging.TaggedWord) bool {
	// Java: noun(?!.prop).* — noun without prop
	for _, tw := range tags {
		if strings.HasPrefix(tw.PosTag, "noun") && !strings.Contains(tw.PosTag, "prop") {
			return true
		}
	}
	return false
}

func toolsIsCapitalized(s string) bool {
	rs := []rune(s)
	if len(rs) == 0 || !unicode.IsUpper(rs[0]) {
		return false
	}
	for _, r := range rs[1:] {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// DynamicNoDashSolidBlocks: if solid (no hyphen) form tags, Java skips tagMatch for hyphen form.
// When noDash has hits and left is not intj, suppress FullTagMatch (Java only tagMatches when empty).
func DynamicNoDashSolidHasTags(token string, tagWord func(string) []tagging.TaggedWord) bool {
	if tagWord == nil || !strings.Contains(token, "-") {
		return false
	}
	t := normalizeDash(token)
	i := strings.Index(t, "-")
	if i <= 0 {
		return false
	}
	left := t[:i]
	leftTags := tagAsIsAndWithLowerCase(left, tagWord)
	hasIntj := false
	for _, tw := range leftTags {
		if strings.HasPrefix(tw.PosTag, "intj") {
			hasIntj = true
			break
		}
	}
	if hasIntj {
		return false
	}
	noDash := strings.ReplaceAll(t, "-", "")
	if tagging.UTF16Len(noDash) < 2 {
		return false
	}
	return len(tagAsIsAndWithLowerCase(noDash, tagWord)) > 0
}
