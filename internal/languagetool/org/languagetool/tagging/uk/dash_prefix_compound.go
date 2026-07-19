package uk

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// DASH_PREFIX_LAT_PATTERN: Latin (3+) or single Greek letter as left prefixes.
var dashPrefixLatRE = regexp.MustCompile(`^(?:[a-zA-Z]{3,}|[α-ωΑ-Ω])$`)

// left single letter / Latin-Greek token for adj compounds (Java: [А-ЯІЇЄҐa-zA-Zα-ωΑ-Ω]|[a-zA-Z-]+)
var leftLatOrLetterRE = regexp.MustCompile(`^(?:[А-ЯІЇЄҐA-Za-zα-ωΑ-Ω]|[a-zA-Z-]+)$`)

var (
	// drop :comp. (one char after comp)
	compDropRE = regexp.MustCompile(`:comp.`)
)

// noun|adj without pron (Java (noun|adj)(?!.*pron).*)
func isNounOrAdjNoPron(pos string) bool {
	if !strings.HasPrefix(pos, "noun") && !strings.HasPrefix(pos, "adj") {
		return false
	}
	return !strings.Contains(pos, "pron")
}

// adj without pron/bad/slang/arch (Java adj(?!.*(pron|bad|slang|arch)).*)
func isAdjClean(pos string) bool {
	if !strings.HasPrefix(pos, "adj") {
		return false
	}
	for _, bad := range []string{"pron", "bad", "slang", "arch"} {
		if strings.Contains(pos, bad) {
			return false
		}
	}
	return true
}

// GenerateTokensWithRightInflected ports CompoundTagger.generateTokensWithRighInflected.
func GenerateTokensWithRightInflected(word, leftWord string, rightTokens []*languagetool.AnalyzedToken, posTagStart, addTag string, dropRE *regexp.Regexp) []*languagetool.AnalyzedToken {
	var out []*languagetool.AnalyzedToken
	for _, at := range rightTokens {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		pos := *at.GetPOSTag()
		if !strings.HasPrefix(pos, posTagStart) || strings.Contains(pos, "v_kly") {
			continue
		}
		if dropRE != nil {
			pos = dropRE.ReplaceAllString(pos, "")
		}
		if addTag != "" {
			pos = AddIfNotContains(pos, addTag)
		}
		lemma := leftWord + "-" + lemmaOfToken(at)
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// DynamicSingleLetterRedupReadings ports з-зателефоную path:
// left length 1, right starts with same letter (lower), retag right + :alt.
func DynamicSingleLetterRedupReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	t := normalizeDash(token)
	if strings.Count(t, "-") != 1 {
		return nil
	}
	i := strings.Index(t, "-")
	left, right := t[:i], t[i+1:]
	if utf8.RuneCountInString(left) != 1 || utf8.RuneCountInString(right) <= 3 {
		return nil
	}
	leftLow := strings.ToLower(left)
	if !strings.HasPrefix(strings.ToLower(right), leftLow) {
		return nil
	}
	wd := tagWord(right)
	if len(wd) == 0 {
		return nil
	}
	wd = AddIfNotContainsWords(wd, ":alt", "")
	return taggedWordsToSurfaceTokens(token, wd)
}

// DynamicInvalidDashPrefixReadings ports dashPrefixesInvalid left:
// filter noun|adj (!pron); lemma left+"-"+right lemma; :bad unless right capitalized.
func DynamicInvalidDashPrefixReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	t := normalizeDash(token)
	if strings.Count(t, "-") != 1 {
		return nil
	}
	i := strings.Index(t, "-")
	left, right := t[:i], t[i+1:]
	if left == "" || right == "" {
		return nil
	}
	if !IsDashPrefixInvalid(left) {
		return nil
	}
	rightWd := tagEitherCase(right, tagWord)
	var filtered []tagging.TaggedWord
	for _, tw := range rightWd {
		if isNounOrAdjNoPron(tw.PosTag) {
			filtered = append(filtered, tw)
		}
	}
	rightWd = filtered
	if len(rightWd) == 0 {
		return nil
	}
	extra := ":bad"
	if rs := []rune(right); len(rs) > 0 && unicode.IsUpper(rs[0]) {
		extra = ""
	}
	// Java adjust(rightWdList, leftWord+"-", null, extraTag)
	adj := Adjust(rightWd, left+"-", "", extra)
	return taggedWordsToSurfaceTokens(token, adj)
}

// DynamicDashPrefixReadings ports main single-dash dashPrefixes / LAT / top-numr / adj compounds.
func DynamicDashPrefixReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	t := normalizeDash(token)
	if strings.Count(t, "-") != 1 {
		return nil
	}
	// trailing/leading dash rejected by Java doGuessCompoundTag
	if strings.HasPrefix(t, "-") || strings.HasSuffix(t, "-") {
		return nil
	}
	i := strings.LastIndex(t, "-")
	left, right := t[:i], t[i+1:]
	if left == "" || right == "" {
		return nil
	}
	leftLow := strings.ToLower(left)

	// invalid prefixes handled separately (call order)
	if IsDashPrefixInvalid(left) {
		return nil
	}

	latMatch := dashPrefixLatRE.MatchString(left)
	dashPrefixMatch := IsDashPrefix(left) || latMatch
	if !dashPrefixMatch {
		// still allow single-letter/Latin adj compounds without being in dash_prefixes
		return latinLetterAdjCompounds(token, left, right, tagWord)
	}

	rightWd := tagEitherCase(right, tagWord)
	if len(rightWd) == 0 {
		return nil
	}
	rightTokens := taggedWordsToTokens(right, rightWd)

	// adj compounds for single letter / Latin left
	var adjCompounds []*languagetool.AnalyzedToken
	if leftLatOrLetterRE.MatchString(left) {
		var adjRight []*languagetool.AnalyzedToken
		for _, r := range rightTokens {
			if r != nil && r.GetPOSTag() != nil && isAdjClean(*r.GetPOSTag()) {
				adjRight = append(adjRight, r)
			}
		}
		adjCompounds = GenerateTokensWithRightInflected(token, left, adjRight, "adj", "", compDropRE)
	}

	// skip міді-бронза
	if strings.EqualFold(left, "міді") && hasLemmaAny(rightTokens, "бронза") {
		return nil
	}

	extraTag := ""
	lowerCased := false
	if v, ok := DashPrefixExtraTag(left); ok {
		extraTag = v
		// if only lower key hit and cyrillic lower → lowerCased
		loadDashPrefixResources()
		if _, exact := dashPrefixes[left]; !exact {
			if _, low := dashPrefixes[leftLow]; low {
				if cyrillicWord(leftLow) {
					lowerCased = true
				}
			}
		}
	}

	leftKey := left
	if lowerCased {
		leftKey = leftLow
	}

	var newTokens []*languagetool.AnalyzedToken
	if nounToks := GetNvPrefixNounMatch(token, rightTokens, leftKey, extraTag); len(nounToks) > 0 {
		newTokens = append(newTokens, nounToks...)
	}

	// топ-десять
	if strings.EqualFold(left, "топ") && hasPosPartInList(rightTokens, "numr:") {
		return GenerateTokensWithRightInflected(token, left, rightTokens, "numr:", ":bad", nil)
	}

	if len(newTokens) == 0 {
		newTokens = append(newTokens, adjCompounds...)
	} else if len(adjCompounds) > 0 {
		// Java may keep both; keep noun first then adj
		newTokens = append(newTokens, adjCompounds...)
	}

	if len(newTokens) == 0 {
		return nil
	}
	return dedupeTokens(newTokens)
}

func latinLetterAdjCompounds(token, left, right string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if !leftLatOrLetterRE.MatchString(left) {
		return nil
	}
	rightWd := tagEitherCase(right, tagWord)
	if len(rightWd) == 0 {
		return nil
	}
	rightTokens := taggedWordsToTokens(right, rightWd)
	var adjRight []*languagetool.AnalyzedToken
	for _, r := range rightTokens {
		if r != nil && r.GetPOSTag() != nil && isAdjClean(*r.GetPOSTag()) {
			adjRight = append(adjRight, r)
		}
	}
	return GenerateTokensWithRightInflected(token, left, adjRight, "adj", "", compDropRE)
}

func normalizeDash(token string) string {
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")
	t = strings.ReplaceAll(t, "\u2011", "-")
	return t
}

func hasLemmaAny(toks []*languagetool.AnalyzedToken, lemma string) bool {
	for _, t := range toks {
		if lemmaOfToken(t) == lemma {
			return true
		}
	}
	return false
}

func cyrillicWord(s string) bool {
	for _, r := range s {
		if r == '\'' {
			continue
		}
		if !isUkrLetter(r) {
			return false
		}
	}
	return s != ""
}
