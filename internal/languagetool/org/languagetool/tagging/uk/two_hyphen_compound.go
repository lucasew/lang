package uk

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// GetNvPrefixNounMatch ports CompoundTagger.getNvPrefixNounMatch.
// Right-side noun POS (no v_kly); lemma = left + "-" + right lemma; optional extraTag.
func GetNvPrefixNounMatch(word string, rightTokens []*languagetool.AnalyzedToken, leftWord, extraTag string) []*languagetool.AnalyzedToken {
	var out []*languagetool.AnalyzedToken
	for _, at := range rightTokens {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		pos := *at.GetPOSTag()
		if !strings.HasPrefix(pos, "noun") || strings.Contains(pos, "v_kly") {
			continue
		}
		// Java: if extraTag is not :alt OR lemma not capitalized, apply extraTag
		lemmaRight := lemmaOfToken(at)
		applyExtra := true
		if extraTag == ":alt" && lemmaRight != "" {
			rs := []rune(lemmaRight)
			if len(rs) > 0 && unicode.IsUpper(rs[0]) {
				applyExtra = false
			}
		}
		if applyExtra && extraTag != "" {
			pos = AddIfNotContains(pos, extraTag)
		}
		newLemma := leftWord + "-" + lemmaRight
		p, l := pos, newLemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// tagEitherCase ports CompoundTagger.tagEitherCase (as-is then lower if capitalized).
func tagEitherCase(word string, tagWord func(string) []tagging.TaggedWord) []tagging.TaggedWord {
	if tagWord == nil || word == "" {
		return nil
	}
	tws := tagWord(word)
	if len(tws) > 0 {
		return tws
	}
	rs := []rune(word)
	if len(rs) > 0 && unicode.IsUpper(rs[0]) {
		return tagWord(strings.ToLower(word))
	}
	return nil
}

// DynamicTwoHyphenReadings ports CompoundTagger.doGuessTwoHyphens (exactly two dashes).
// Paths: dash prefix on first-second; adj second+third via TagMatch; oAdj chain.
func DynamicTwoHyphenReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" {
		return nil
	}
	// normalize dash variants
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")
	t = strings.ReplaceAll(t, "\u2011", "-")
	if strings.Count(t, "-") != 2 {
		return nil
	}
	parts := strings.Split(t, "-")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil
	}

	rightWdList := tagEitherCase(parts[2], tagWord)
	if len(rightWdList) == 0 {
		return nil
	}
	rightTokens := taggedWordsToTokens(parts[2], rightWdList)

	firstAndSecond := parts[0] + "-" + parts[1]
	if extra, ok := DashPrefixExtraTag(firstAndSecond); ok {
		loadDashPrefixResources()
		leftKey := firstAndSecond
		if _, ok2 := dashPrefixes[firstAndSecond]; !ok2 {
			leftKey = strings.ToLower(firstAndSecond)
		}
		return GetNvPrefixNounMatch(token, rightTokens, leftKey, extra)
	}

	// try full match — only adj for second part
	secondWdList := tagEitherCase(parts[1], tagWord)
	if HasPosTagStart2(secondWdList, "adj") {
		secondTokens := taggedWordsToTokens(parts[1], secondWdList)
		// TagMatch surface = full word (Java uses full surface for all compound tokens)
		tagMatchSecondAndThird := TagMatch(token, secondTokens, rightTokens)
		if len(tagMatchSecondAndThird) > 0 {
			// Java also tries left+secondAndThird but discards the result
			if leftWd := tagEitherCase(parts[0], tagWord); len(leftWd) > 0 {
				_ = TagMatch(token, taggedWordsToTokens(parts[0], leftWd), tagMatchSecondAndThird)
			}
			return tagMatchSecondAndThird
		}
	}

	// try ірансько-нігерійсько-зімбабвійський — oAdj chain
	secondAndThird := tryOWithAdjTokens(token, parts[1], rightTokens, tagWord)
	if len(secondAndThird) > 0 {
		return tryOWithAdjTokens(token, parts[0], secondAndThird, tagWord)
	}
	return nil
}

// tryOWithAdjTokens ports tryOWithAdj(word, leftWord, rightAnalyzedTokens).
// Builds synthetic right-side tokens already tagged; left is o-adj stem.
func tryOWithAdjTokens(word, leftWord string, rightTokens []*languagetool.AnalyzedToken, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if leftWord == "" || len(rightTokens) == 0 {
		return nil
	}
	// Require left looks like o-adj stem (ends with о/е) — Java O_ADJ_PATTERN .+(о|[чшщ]е)
	if !isOAdjLeft(leftWord) {
		return nil
	}
	// oAdjMatch: right must be adj readings
	var adjRight []*languagetool.AnalyzedToken
	for _, r := range rightTokens {
		if r != nil && r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), "adj") {
			adjRight = append(adjRight, r)
		}
	}
	if len(adjRight) == 0 {
		return nil
	}
	// left validation (LEFT_O_ADJ / dict) — reuse leftOAdjDictOK
	if !leftOAdjDictOK(leftWord, tagWord) {
		// LEFT_O_ADJ_INVALID still produces :bad in oAdjMatch for hyphen forms
		leftLow := strings.ToLower(leftWord)
		if _, inv := leftOAdjInvalid[leftLow]; !inv {
			return nil
		}
	}

	extraTag := ""
	leftLow := strings.ToLower(leftWord)
	if _, inv := leftOAdjInvalid[leftLow]; inv {
		// capitalized dual prop exception handled elsewhere; here solid invalid → :bad
		extraTag = ":bad"
	}

	var out []*languagetool.AnalyzedToken
	for _, r := range adjRight {
		pos := *r.GetPOSTag()
		if strings.Contains(pos, "v_kly") {
			continue
		}
		// strip :comp*
		pos = AdjCompRegex.ReplaceAllString(pos, "")
		if strings.Contains(extraTag, ":bad") {
			pos = strings.ReplaceAll(pos, ":arch", "")
		}
		if extraTag != "" {
			pos = AddIfNotContains(pos, extraTag)
		}
		// Java oAdjMatch lemma: leftWord.toLowerCase() + "-" + right lemma
		lemma := leftLow + "-" + lemmaOfToken(r)
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isOAdjLeft(leftWord string) bool {
	rs := []rune(leftWord)
	if len(rs) < 2 {
		return false
	}
	last := unicode.ToLower(rs[len(rs)-1])
	if last == 'о' {
		return true
	}
	// [чшщ]е
	if last == 'е' && len(rs) >= 2 {
		prev := unicode.ToLower(rs[len(rs)-2])
		return prev == 'ч' || prev == 'ш' || prev == 'щ'
	}
	return false
}

