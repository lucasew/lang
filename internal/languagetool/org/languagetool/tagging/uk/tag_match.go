package uk

import (
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Ports CompoundTagger.tagMatch + agreement helpers (getAgreedPosTag, tryAnimInanim, …).

var (
	extraTagsRE           = regexp.MustCompile(`:bad`)
	extraTagsDropRE       = regexp.MustCompile(`:(?:comp.|np|ns|slang|xp[1-9]|predic|insert)`)
	extraTagsDropNoninflRE = regexp.MustCompile(`:(?:comp.|np|ns|slang|xp[1-9]|insert)`)
	stdNounTagRE          = regexp.MustCompile(`^noun:(?:in)?anim:(.):(v_...).*$`)
	singRegexF            = regexp.MustCompile(`:[mfn]:`)
	mnpNazRE              = regexp.MustCompile(`^.*?:[mnp]:v_naz.*$`)
	mnpZnaRE              = regexp.MustCompile(`^.*?:[mnp]:v_zna.*$`)
	mnpRodRE              = regexp.MustCompile(`^.*?:[mnp]:v_rod.*$`)
	stripPerfImperfRE     = regexp.MustCompile(`:(?:im)?perf|:adjp:(?:actv|pasv)`)
	adjpStripRE           = regexp.MustCompile(`:adjp:(?:actv|pasv):(?:im)?perf`)

	// local copies of LemmaHelper lists (tagging/uk cannot import rules/uk).
	tagMatchDaysOfWeek = []string{"понеділок", "вівторок", "середа", "четвер", "п'ятниця", "субота", "неділя"}
	tagMatchMonths     = []string{
		"січень", "лютий", "березень", "квітень", "травень", "червень", "липень",
		"серпень", "вересень", "жовтень", "листопад", "грудень",
	}

	dashMasterFollowerOnce sync.Once
	leftMasterSet          map[string]struct{}
	followerSet            map[string]struct{}
)

func loadMasterFollower() {
	dashMasterFollowerOnce.Do(func() {
		leftMasterSet = map[string]struct{}{}
		followerSet = map[string]struct{}{}
		if p := discoverUKResource("dash_left_master.txt"); p != "" {
			loadSetInto(p, leftMasterSet)
		}
		if p := discoverUKResource("dash_follower.txt"); p != "" {
			loadSetInto(p, followerSet)
		}
	})
}

func dropExtra(pos string) string {
	if strings.HasPrefix(pos, "noninfl") {
		return extraTagsDropNoninflRE.ReplaceAllString(pos, "")
	}
	return extraTagsDropRE.ReplaceAllString(pos, "")
}

func stripPerfImperf(pos string) string {
	return stripPerfImperfRE.ReplaceAllString(pos, "")
}

// TagMatch ports CompoundTagger.tagMatch for two sides already tagged.
// Returns nil when no agreement (Java null).
func TagMatch(word string, leftToks, rightToks []*languagetool.AnalyzedToken) []*languagetool.AnalyzedToken {
	loadMasterFollower()
	var newTokens []*languagetool.AnalyzedToken
	var animInanimTokens []*languagetool.AnalyzedToken
	// animInanimNotTagged unused except debug

	for _, left := range leftToks {
		if left == nil || left.GetPOSTag() == nil {
			continue
		}
		leftPosTag := *left.GetPOSTag()
		if strings.Contains(leftPosTag, "abbr") {
			continue
		}
		if strings.HasPrefix(leftPosTag, "noun:inanim") && strings.Contains(leftPosTag, "v_kly") {
			continue
		}

		leftPosTagExtra := ""
		leftNv := false
		if strings.Contains(leftPosTag, NoVidminokSubstr) {
			leftNv = true
			leftPosTag = strings.ReplaceAll(leftPosTag, NoVidminokSubstr, "")
		}
		leftPosTag = dropExtra(leftPosTag)
		if m := extraTagsRE.FindString(leftPosTag); m != "" {
			leftPosTagExtra += m
			leftPosTag = extraTagsRE.ReplaceAllString(leftPosTag, "")
		}

		for _, right := range rightToks {
			if right == nil || right.GetPOSTag() == nil {
				continue
			}
			rightPosTag := *right.GetPOSTag()
			if strings.Contains(rightPosTag, "abbr") || strings.Contains(rightPosTag, "v_zna:var") {
				continue
			}
			if strings.HasPrefix(rightPosTag, "noun:inanim") {
				if strings.Contains(rightPosTag, "v_kly") {
					continue
				}
				if strings.Contains(leftPosTag, ":geo") && !strings.Contains(rightPosTag, ":geo") {
					rl := lemmaOfToken(right)
					if !matchesGeoRightLemma(rl) {
						continue
					}
				}
			}
			if strings.HasPrefix(rightPosTag, "noun:anim:p:v_zna:rare") && strings.HasPrefix(leftPosTag, "noun:inanim") {
				continue
			}

			extraNvTag := ""
			rightNv := false
			if strings.Contains(rightPosTag, NoVidminokSubstr) {
				rightNv = true
				if leftNv {
					extraNvTag += NoVidminokSubstr
				}
			}
			rightPosTag = dropExtra(rightPosTag)
			rightPosTag = extraTagsRE.ReplaceAllString(rightPosTag, "")

			// equal POS family (adj/verb/adv/numr/intj redup)
			if stripPerfImperf(leftPosTag) == stripPerfImperf(rightPosTag) &&
				(startsWithAnyPOS(leftPosTag, "numr", "adv", "adj", "verb") ||
					(isIntjOrNoninfl(leftPosTag) && equalFoldLemma(left, right))) {
				newPosTag := leftPosTag + extraNvTag + leftPosTagExtra
				if (strings.Contains(leftPosTag, "adjp") && !strings.Contains(rightPosTag, "adjp")) ||
					(!strings.Contains(leftPosTag, "adjp") && strings.Contains(rightPosTag, "adjp")) {
					newPosTag = adjpStripRE.ReplaceAllString(newPosTag, "")
				}
				newTokens = append(newTokens, newCompoundToken(word, newPosTag, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
				continue
			}

			// noun-noun
			if strings.HasPrefix(leftPosTag, "noun") && strings.HasPrefix(rightPosTag, "noun") {
				agreed := getAgreedPosTag(leftPosTag, rightPosTag, leftNv, word)
				if agreed == "" && strings.HasPrefix(rightPosTag, "noun:inanim:m:v_naz") && isMinMax(right.GetToken()) {
					agreed = leftPosTag
				}
				if agreed == "" && !isSameAnimStatus(leftPosTag, rightPosTag) {
					agreed = tryAnimInanim(leftPosTag, rightPosTag, lemmaOfToken(left), lemmaOfToken(right), leftNv, rightNv, word)
					if agreed != "" {
						animInanimTokens = append(animInanimTokens, newCompoundToken(word, agreed+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
						continue
					}
				}
				if agreed != "" {
					newTokens = append(newTokens, newCompoundToken(word, agreed+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
				}
				continue
			}

			// numr-numr
			if strings.HasPrefix(leftPosTag, "numr") && strings.HasPrefix(rightPosTag, "numr") {
				agreed := getNumAgreedPosTag(leftPosTag, rightPosTag, leftNv)
				if agreed != "" {
					if strings.Contains(rightPosTag, ":p:") && !strings.Contains(agreed, ":p:") {
						agreed = singRegexF.ReplaceAllString(agreed, ":p:")
					}
					newTokens = append(newTokens, newCompoundToken(word, agreed+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
				}
				continue
			}

			// noun-numr
			if strings.HasPrefix(leftPosTag, "noun") && strings.HasPrefix(rightPosTag, "numr") {
				if lemmaOfToken(left) != "п'ята" {
					lgc := GetGenderConj(leftPosTag)
					if lgc != "" && lgc == GetGenderConj(rightPosTag) {
						newTokens = append(newTokens, newCompoundToken(word, leftPosTag+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
						if !strings.Contains(leftPosTag, ":p:") {
							pl := singRegexF.ReplaceAllString(leftPosTag, ":p:")
							newTokens = append(newTokens, newCompoundToken(word, pl+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
						}
					} else if agreed := getNumAgreedPosTag(leftPosTag, rightPosTag, leftNv); agreed != "" {
						newTokens = append(newTokens, newCompoundToken(word, agreed+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
						if !strings.Contains(agreed, ":p:") {
							pl := singRegexF.ReplaceAllString(agreed, ":p:")
							newTokens = append(newTokens, newCompoundToken(word, pl+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
						}
					}
				}
				continue
			}

			// noun-adj junior/senior OR noun-numr (Java elseif structure)
			if strings.HasPrefix(leftPosTag, "noun") &&
				(strings.HasPrefix(rightPosTag, "numr") || (strings.HasPrefix(rightPosTag, "adj") && isJuniorSenior(left, right))) {
				lgc := GetGenderConj(leftPosTag)
				if lgc != "" && lgc == GetGenderConj(rightPosTag) {
					newTokens = append(newTokens, newCompoundToken(word, leftPosTag+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+lemmaOfToken(right)))
				}
				continue
			}

			// чарка-друга
			if strings.HasPrefix(leftPosTag, "noun") && lemmaOfToken(right) == "другий" {
				lgc := GetGenderConj(leftPosTag)
				if lgc != "" && lgc == GetGenderConj(rightPosTag) {
					rightLemma := "друге"
					if strings.HasPrefix(lgc, "m") {
						rightLemma = "другий"
					} else if strings.HasPrefix(lgc, "f") {
						rightLemma = "друга"
					}
					newTokens = append(newTokens, newCompoundToken(word, leftPosTag+extraNvTag+leftPosTagExtra, lemmaOfToken(left)+"-"+rightLemma))
				}
			}
		}
	}

	// days/months → also plural if no :p: yet
	if len(newTokens) > 0 && !hasPosPartInList(newTokens, ":p:") {
		if (hasLemmaInList(leftToks, tagMatchDaysOfWeek) && hasLemmaInList(rightToks, tagMatchDaysOfWeek)) ||
			(hasLemmaInList(leftToks, tagMatchMonths) && hasLemmaInList(rightToks, tagMatchMonths)) {
			first := newTokens[0]
			pos := ""
			if first.GetPOSTag() != nil {
				pos = singRegexF.ReplaceAllString(*first.GetPOSTag(), ":p:")
			}
			newTokens = append(newTokens, newCompoundToken(word, pos, lemmaOfToken(first)))
		}
	}

	// dedupe by pos|lemma
	newTokens = dedupeTokens(newTokens)
	if len(newTokens) == 0 {
		newTokens = dedupeTokens(animInanimTokens)
	}
	if len(newTokens) == 0 {
		return nil
	}
	return newTokens
}

func matchesGeoRightLemma(lemma string) bool {
	low := strings.ToLower(lemma)
	switch low {
	case "ріка", "гора", "місто", "град", "поле", "море", "парк":
		return true
	}
	return false
}

func startsWithAnyPOS(pos string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(pos, p) {
			return true
		}
	}
	return false
}

func isIntjOrNoninfl(pos string) bool {
	return pos == "intj" || strings.HasPrefix(pos, "noninfl")
}

func equalFoldLemma(a, b *languagetool.AnalyzedToken) bool {
	return strings.EqualFold(lemmaOfToken(a), lemmaOfToken(b))
}

func lemmaOfToken(t *languagetool.AnalyzedToken) string {
	if t == nil {
		return ""
	}
	if t.GetLemma() != nil {
		return *t.GetLemma()
	}
	return t.GetToken()
}

func newCompoundToken(word, pos, lemma string) *languagetool.AnalyzedToken {
	p, l := pos, lemma
	return languagetool.NewAnalyzedToken(word, &p, &l)
}

func isMinMax(tok string) bool {
	return tok == "максимум" || tok == "мінімум"
}

func isSameAnimStatus(left, right string) bool {
	return strings.Contains(left, ":anim") == strings.Contains(right, ":anim")
}

func isPluralNoun(pos string) bool {
	return strings.HasPrefix(pos, "noun:") && strings.Contains(pos, ":p:")
}

func getNumAgreedPosTag(leftPosTag, rightPosTag string, leftNv bool) string {
	_ = leftNv
	leftP := strings.Contains(leftPosTag, ":p:")
	rightSing := singRegexF.MatchString(rightPosTag)
	rightP := strings.Contains(rightPosTag, ":p:")
	leftSing := singRegexF.MatchString(leftPosTag)
	if (leftP && rightSing) || (leftSing && rightP) {
		lc := GetConj(leftPosTag)
		if lc != "" && lc == GetConj(rightPosTag) {
			return leftPosTag
		}
	}
	return ""
}

func getAgreedPosTag(leftPosTag, rightPosTag string, leftNv bool, word string) string {
	if isPluralNoun(leftPosTag) != isPluralNoun(rightPosTag) {
		return ""
	}
	if !isSameAnimStatus(leftPosTag, rightPosTag) {
		return ""
	}
	ml := stdNounTagRE.FindStringSubmatch(leftPosTag)
	mr := stdNounTagRE.FindStringSubmatch(rightPosTag)
	if ml == nil || mr == nil {
		return ""
	}
	// groups: full, gender, case
	if ml[2] != mr[2] {
		return ""
	}
	if ml[1] != mr[1] {
		// gender mix: only for longer compounds
		if len([]rune(word)) < 10 {
			return ""
		}
	}
	if leftNv {
		return rightPosTag
	}
	return leftPosTag
}

func tryAnimInanim(leftPosTag, rightPosTag, leftLemma, rightLemma string, leftNv, rightNv bool, word string) string {
	if _, ok := leftMasterSet[leftLemma]; ok {
		if strings.Contains(leftPosTag, ":anim") {
			rightPosTag = strings.ReplaceAll(rightPosTag, ":inanim", ":anim")
		} else {
			rightPosTag = strings.ReplaceAll(rightPosTag, ":anim", ":inanim")
		}
		if a := getAgreedPosTag(leftPosTag, rightPosTag, leftNv, word); a != "" {
			return a
		}
		if !strings.Contains(leftPosTag, ":anim") {
			if posMatchesFull(mnpZnaRE, leftPosTag) && posMatchesFull(mnpNazRE, rightPosTag) && !leftNv && !rightNv {
				return leftPosTag
			}
		} else {
			if posMatchesFull(mnpZnaRE, leftPosTag) && posMatchesFull(mnpRodRE, rightPosTag) && !leftNv && !rightNv {
				return leftPosTag
			}
		}
		return ""
	}
	if _, ok := followerSet[rightLemma]; ok {
		rightPosTag = strings.ReplaceAll(rightPosTag, ":anim", ":inanim")
		if a := getAgreedPosTag(leftPosTag, rightPosTag, false, word); a != "" {
			return a
		}
		if strings.Contains(leftPosTag, ":inanim") {
			if posMatchesFull(mnpZnaRE, leftPosTag) && posMatchesFull(mnpNazRE, rightPosTag) &&
				GetNum(leftPosTag) == GetNum(rightPosTag) && !leftNv && !rightNv {
				return leftPosTag
			}
		}
		return ""
	}
	if _, ok := followerSet[leftLemma]; ok {
		leftPosTag = strings.ReplaceAll(leftPosTag, ":anim", ":inanim")
		if a := getAgreedPosTag(rightPosTag, leftPosTag, false, word); a != "" {
			return a
		}
		if strings.Contains(rightPosTag, ":inanim") {
			if posMatchesFull(mnpZnaRE, rightPosTag) && posMatchesFull(mnpNazRE, leftPosTag) &&
				GetNum(leftPosTag) == GetNum(rightPosTag) && !leftNv && !rightNv {
				return rightPosTag
			}
		}
	}
	return ""
}

var juniorSeniorNameRE = regexp.MustCompile(`^.*?:[flp]name.*$`)

func isJuniorSenior(left, right *languagetool.AnalyzedToken) bool {
	if left == nil || right == nil || left.GetPOSTag() == nil {
		return false
	}
	// Java: left POS .matches(".*?:[flp]name.*"); right lemma .matches(".*(молодший|старший)")
	if !posMatchesFull(juniorSeniorNameRE, *left.GetPOSTag()) {
		return false
	}
	rl := lemmaOfToken(right)
	return strings.Contains(rl, "молодший") || strings.Contains(rl, "старший")
}

func hasPosPartInList(toks []*languagetool.AnalyzedToken, part string) bool {
	for _, t := range toks {
		if t != nil && t.GetPOSTag() != nil && strings.Contains(*t.GetPOSTag(), part) {
			return true
		}
	}
	return false
}

func hasLemmaInList(toks []*languagetool.AnalyzedToken, lemmas []string) bool {
	set := map[string]struct{}{}
	for _, l := range lemmas {
		set[l] = struct{}{}
	}
	for _, t := range toks {
		if _, ok := set[lemmaOfToken(t)]; ok {
			return true
		}
	}
	return false
}

func dedupeTokens(toks []*languagetool.AnalyzedToken) []*languagetool.AnalyzedToken {
	seen := map[string]struct{}{}
	var out []*languagetool.AnalyzedToken
	for _, t := range toks {
		if t == nil {
			continue
		}
		pos, lem := "", ""
		if t.GetPOSTag() != nil {
			pos = *t.GetPOSTag()
		}
		if t.GetLemma() != nil {
			lem = *t.GetLemma()
		}
		key := pos + "|" + lem
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	return out
}

// taggedWordsToTokens converts dict hits to AnalyzedTokens with given surface.
func taggedWordsToTokens(surface string, words []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	var out []*languagetool.AnalyzedToken
	for _, tw := range words {
		p, l := tw.PosTag, tw.Lemma
		if p == "" {
			continue
		}
		out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
	}
	return out
}

// FullTagMatchViaTagMatch reimplements FullTagMatchReadings using TagMatch.
func FullTagMatchViaTagMatch(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")
	t = strings.ReplaceAll(t, "\u2011", "-")
	if strings.Count(t, "-") != 1 {
		return nil
	}
	parts := strings.SplitN(t, "-", 2)
	left, right := parts[0], parts[1]
	if left == "" || right == "" {
		return nil
	}
	leftTags := lookupBothCases(left, tagWord)
	rightTags := lookupBothCases(right, tagWord)
	if len(leftTags) == 0 || len(rightTags) == 0 {
		return nil
	}
	leftToks := taggedWordsToTokens(left, leftTags)
	rightToks := taggedWordsToTokens(right, rightTags)
	matched := TagMatch(token, leftToks, rightToks)
	if len(matched) == 0 {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, m := range matched {
		pos, lem := "", ""
		if m.GetPOSTag() != nil {
			pos = *m.GetPOSTag()
		}
		if m.GetLemma() != nil {
			lem = *m.GetLemma()
		}
		if pos != "" && !strings.Contains(pos, "prop") {
			parts := strings.SplitN(lem, "-", 2)
			for i := range parts {
				parts[i] = strings.ToLower(parts[i])
			}
			lem = strings.Join(parts, "-")
		}
		p, l := pos, lem
		out = append(out, languagetool.NewAnalyzedToken(token, &p, &l))
	}
	return out
}
