package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// leftOAdjInvalid ports CompoundTagger.LEFT_O_ADJ_INVALID — directional / degree
// prefixes that normally write solid; hyphen form is :bad unless both parts are
// capitalized and the full lower compound is already an adj in the dict.
var leftOAdjInvalid = map[string]struct{}{
	"багато": {}, "мало": {}, "високо": {}, "низько": {}, "старо": {}, "важко": {},
	"зовнішньо": {}, "внутрішньо": {}, "ново": {}, "середньо": {},
	"південно": {}, "північно": {}, "західно": {}, "східно": {}, "центрально": {},
	"ранньо": {}, "пізньо": {},
}

// leftOAdj ports CompoundTagger.LEFT_O_ADJ — left stems allowed without left-side dict tag.
var leftOAdj = map[string]struct{}{
	"австро": {}, "адиго": {}, "американо": {}, "англо": {}, "афро": {},
	"еко": {}, "індо": {}, "іспано": {}, "італо": {}, "історико": {},
	"києво": {}, "марокано": {}, "угро": {}, "японо": {}, "румуно": {},
}

// Directional left stems commonly produced by Java oAdjMatch for geo/direction compounds.
var directionalLeft = map[string]struct{}{
	"південно": {}, "північно": {}, "східно": {}, "західно": {},
	"центрально": {}, "ранньо": {}, "пізньо": {},
}

// oToYj ports CompoundTagger.oToYj (кричуще → кричущий, ясно → ясний).
func oToYj(leftWord string) string {
	low := leftWord
	if strings.HasSuffix(low, "ьо") {
		return low[:len(low)-len("ьо")] + "ій"
	}
	if low == "" {
		return low
	}
	// drop last rune (о/е) + "ий"
	rs := []rune(low)
	if len(rs) < 2 {
		return low
	}
	return string(rs[:len(rs)-1]) + "ий"
}

// leftOAdjDictOK ports oAdjMatch left-side wordTagger checks when left ∉ LEFT_O_ADJ.
func leftOAdjDictOK(leftWord string, tagWord func(string) []tagging.TaggedWord) bool {
	if tagWord == nil {
		return false
	}
	if _, ok := leftOAdj[strings.ToLower(leftWord)]; ok {
		return true
	}
	return len(oAdjLeftTaggedWords(leftWord, tagWord)) > 0
}

// oAdjLeftTaggedWords ports oAdjMatch's left-side tagBothCases ladder (Java):
//
//	tagBothCases(left, ^adv.*|.*?numr.*)
//	tagBothCases(oToYj(left), ^adj.*)
//	if left.length()>4: tagBothCases(leftBase, ^noun.*)
//	tagBothCases(leftBase+"а", (noun:inanim:f:v_naz|numr).*)
//
// Used both for acceptance and for :bad evidence (hasPosTagPart2(taggedWords, ":bad")).
// When left ∈ LEFT_O_ADJ, returns nil (no evidence list; caller treats membership separately).
func oAdjLeftTaggedWords(leftWord string, tagWord func(string) []tagging.TaggedWord) []tagging.TaggedWord {
	if tagWord == nil || leftWord == "" {
		return nil
	}
	leftLow := strings.ToLower(leftWord)
	// tagBothCases(leftWord, Pattern.compile("^adv.*|.*?numr.*"))
	if tws := filterTagBothCases(leftWord, tagWord, func(pos string) bool {
		return strings.HasPrefix(pos, "adv") || strings.Contains(pos, "numr")
	}); len(tws) > 0 {
		return tws
	}
	// oToYj as adj
	if tws := filterTagBothCases(oToYj(leftLow), tagWord, func(pos string) bool {
		return strings.HasPrefix(pos, "adj")
	}); len(tws) > 0 {
		return tws
	}
	rs := []rune(leftLow)
	if len(rs) < 2 {
		return nil
	}
	base := string(rs[:len(rs)-1])
	// Java: if taggedWords.isEmpty() && leftWord.length() > 4 → leftBase as noun
	if tagging.UTF16Len(leftWord) > 4 {
		if tws := filterTagBothCases(base, tagWord, func(pos string) bool {
			return strings.HasPrefix(pos, "noun")
		}); len(tws) > 0 {
			return tws
		}
	}
	// tagBothCases(leftBase+"а", Pattern.compile("(noun:inanim:f:v_naz|numr).*"))
	// always attempted after leftBase (even when length ≤ 4 and leftBase was skipped)
	if tws := filterTagBothCases(base+"а", tagWord, func(pos string) bool {
		if strings.HasPrefix(pos, "numr") {
			return true
		}
		return strings.HasPrefix(pos, "noun:inanim:f:v_naz")
	}); len(tws) > 0 {
		return tws
	}
	return nil
}

// filterTagBothCases ports CompoundTagger.tagBothCases + POS filter.
func filterTagBothCases(word string, tagWord func(string) []tagging.TaggedWord, ok func(string) bool) []tagging.TaggedWord {
	if word == "" || tagWord == nil {
		return nil
	}
	var out []tagging.TaggedWord
	seen := map[string]struct{}{}
	add := func(w string) {
		for _, tw := range tagWord(w) {
			if !ok(tw.PosTag) {
				continue
			}
			key := tw.Lemma + "|" + tw.PosTag
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, tw)
		}
	}
	add(word)
	low := strings.ToLower(word)
	if low != word {
		add(low)
	} else {
		// Java: capitalize when already lower
		rs := []rune(word)
		if len(rs) > 0 {
			rs[0] = unicode.ToUpper(rs[0])
			cap := string(rs)
			if cap != word {
				add(cap)
			}
		}
	}
	return out
}

func hasPOSPrefixOrContains(tagWord func(string) []tagging.TaggedWord, w string, ok func(string) bool) bool {
	if w == "" {
		return false
	}
	for _, form := range []string{w, strings.ToLower(w)} {
		for _, tw := range tagWord(form) {
			if ok(tw.PosTag) {
				return true
			}
		}
		if form == strings.ToLower(w) {
			break
		}
	}
	return false
}

// DynamicDirectionalAdjReadings ports CompoundTagger.oAdjMatch for directional lefts.
// Requires tagWord hits on the right part as adj (Java wordTagger); fail-closed without dict.
// Does not invent case endings from surface alone.
func DynamicDirectionalAdjReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	if dash <= 0 || dash == len(token)-1 {
		return nil
	}
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if tagging.UTF16Len(leftWord) < 3 {
		return nil
	}
	// Java doGuessCompoundTag guards before tryOWithAdj: right part|conj|:pron → null
	// (unless both numr). Blocks лісо-та, відносини-коли, etc.
	leftTagsG := tagAsIsAndWithLowerCase(leftWord, tagWord)
	rightTagsG := tagEitherCase(rightWord, tagWord)
	if !strings.EqualFold(leftWord, rightWord) && hasRightPartConjOrPron(rightTagsG) {
		if !(hasPosStartInTags(leftTagsG, "numr") && hasPosStartInTags(rightTagsG, "numr")) {
			return nil
		}
	}
	leftLow := strings.ToLower(leftWord)
	// Gate shape: known directional left, LEFT_O_ADJ, or ends with о/е like Java O_ADJ_PATTERN.
	if _, ok := directionalLeft[leftLow]; !ok {
		if _, ok2 := leftOAdj[leftLow]; !ok2 {
			if !strings.HasSuffix(leftLow, "о") && !strings.HasSuffix(leftLow, "е") {
				return nil
			}
		}
	}
	// Java oAdjMatch: if left ∉ LEFT_O_ADJ, require left-side dict evidence (adv/adj/noun/numr).
	if _, ok := leftOAdj[leftLow]; !ok {
		if !leftOAdjDictOK(leftWord, tagWord) {
			return nil
		}
	}

	tws := tagWord(rightWord)
	if len(tws) == 0 {
		low := strings.ToLower(rightWord)
		if low != rightWord {
			tws = tagWord(low)
		}
	}
	if len(tws) == 0 {
		return nil
	}

	// Java analyzeAllCapitamizedAdj: both parts capitalized → if full lower is adj, skip :bad.
	skipBadForInvalid := false
	if isCapitalizedWord(leftWord) && isCapitalizedWord(rightWord) {
		for _, tw := range tagWord(strings.ToLower(token)) {
			if strings.HasPrefix(tw.PosTag, "adj") {
				skipBadForInvalid = true
				break
			}
		}
	}

	_, leftInvalid := leftOAdjInvalid[leftLow]
	extraBad := leftInvalid && !skipBadForInvalid

	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" || !strings.HasPrefix(pos, "adj") {
			continue
		}
		// Java: drop :comp. from right before combining
		if i := strings.Index(pos, ":comp"); i >= 0 {
			end := i + len(":comp")
			for end < len(pos) && pos[end] != ':' {
				end++
			}
			pos = pos[:i] + pos[end:]
		}
		if extraBad {
			pos = strings.ReplaceAll(pos, ":arch", "")
			if !strings.Contains(pos, ":bad") {
				pos = pos + ":bad"
			}
		}
		rightLemma := tw.Lemma
		if rightLemma == "" {
			rightLemma = strings.ToLower(rightWord)
		}
		lemma := leftLow + "-" + rightLemma
		key := lemma + "|" + pos
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}

func isCapitalizedWord(s string) bool {
	if s == "" {
		return false
	}
	rs := []rune(s)
	if !unicode.IsUpper(rs[0]) {
		return false
	}
	for _, r := range rs[1:] {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// MissingHyphenCandidates ports UkrainianTagger MISSING_HYPHEN only:
// Pattern.compile("([а-яіїєґ']+)(небудь)") — not inventing dash after anti/mini/….
// Solid no-dash prefixes (мінітест, …) are handled by guessOtherTagsInternal / DynamicNoDashPrefixReadings.
func MissingHyphenCandidates(token string) []string {
	if strings.Contains(token, "-") {
		return nil
	}
	lower := strings.ToLower(token)
	if strings.HasSuffix(lower, "небудь") && tagging.UTF16Len(lower) > tagging.UTF16Len("небудь")+2 {
		rs := []rune(token)
		suf := []rune("небудь")
		stem := string(rs[:len(rs)-len(suf)])
		return []string{stem + "-небудь"}
	}
	return nil
}

// CompoundNumrReadings tags forms like "2-х", "3-ом" soft.
var reCompoundNumr = regexp.MustCompile(`^(\d+)([-–])?(х|ом|им|и|а|е|го|му)?$`)

func CompoundNumrPOS(token string) string {
	if reCompoundNumr.MatchString(token) && strings.ContainsAny(token, "0123456789") {
		hasLetter := false
		for _, r := range token {
			if unicode.IsLetter(r) {
				hasLetter = true
				break
			}
		}
		if hasLetter {
			return "numr"
		}
	}
	return ""
}
