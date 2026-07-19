package uk

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// FullTagMatchReadings tags A-B when both sides share a POS family (Java CompoundTagger
// tagMatch path). tagWord is the dictionary lookup — fail closed when either side is
// untagged (no soft invent of intj/adv/redup without dict).
func FullTagMatchReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")
	t = strings.ReplaceAll(t, "\u2011", "-") // non-breaking hyphen
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
		// Java tags both parts via wordTagger — incomplete without dictionary, not invent.
		return nil
	}

	// find shared POS family and matching case when possible
	var out []*languagetool.AnalyzedToken
	seen := map[string]struct{}{}
	for _, lt := range leftTags {
		fam := posFamily(lt.PosTag)
		if fam == "" {
			continue
		}
		for _, rt := range rightTags {
			if posFamily(rt.PosTag) != fam {
				continue
			}
			// prefer same case marker if present
			caseL, caseR := caseMarker(lt.PosTag), caseMarker(rt.PosTag)
			if caseL != "" && caseR != "" && caseL != caseR {
				continue
			}
			pos := mergePOS(lt.PosTag, rt.PosTag)
			lemma := combineLemma(lt.Lemma, rt.Lemma, left, right)
			key := pos + "|" + lemma
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			p, l := pos, lemma
			// lower lemma for non-proper
			if !strings.Contains(pos, "prop") {
				l = strings.ToLower(l)
			}
			out = append(out, languagetool.NewAnalyzedToken(token, &p, &l))
		}
	}
	return out
}

func lookupBothCases(word string, tagWord func(string) []tagging.TaggedWord) []tagging.TaggedWord {
	if tagWord == nil {
		return nil
	}
	tws := tagWord(word)
	if len(tws) > 0 {
		return tws
	}
	low := strings.ToLower(word)
	if low != word {
		return tagWord(low)
	}
	// title-case try
	rs := []rune(word)
	if len(rs) > 0 {
		rs[0] = unicode.ToUpper(rs[0])
		return tagWord(string(rs))
	}
	return nil
}

func posFamily(pos string) string {
	switch {
	case strings.HasPrefix(pos, "verb"):
		return "verb"
	case strings.HasPrefix(pos, "noun"):
		return "noun"
	case strings.HasPrefix(pos, "adj"):
		return "adj"
	case strings.HasPrefix(pos, "adv"):
		return "adv"
	case strings.HasPrefix(pos, "numr"):
		return "numr"
	case strings.HasPrefix(pos, "intj"):
		// Java CompoundTagger multi-hyphen intj match
		return "intj"
	case strings.HasPrefix(pos, "noninfl"):
		return "noninfl"
	default:
		return ""
	}
}

func caseMarker(pos string) string {
	for _, c := range []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly", "past", "pres", "futr", "impr"} {
		if strings.Contains(pos, c) {
			return c
		}
	}
	return ""
}

func mergePOS(left, right string) string {
	// keep left POS as base (Java often merges lemmas, keeps shared structure)
	return left
}

func combineLemma(l1, l2, surfL, surfR string) string {
	a, b := l1, l2
	if a == "" {
		a = surfL
	}
	if b == "" {
		b = surfR
	}
	return a + "-" + b
}

// NebudMissingHyphenReadings ports UkrainianTagger MISSING_HYPHEN:
// tag base (group 1) via dictionary; require pronoun POS; adjust lemma with -небудь + :bad.
// Without tagWord or without pron readings, fail closed (no soft invent paradigms).
func NebudMissingHyphenReadings(surface, hyphenated string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil {
		return nil
	}
	low := strings.ToLower(hyphenated)
	parts := strings.SplitN(low, "-небудь", 2)
	if len(parts) == 0 || parts[0] == "" {
		return nil
	}
	base := parts[0]
	wdList := lookupBothCases(base, tagWord)
	if len(wdList) == 0 {
		return nil
	}
	// Java: PosTagHelper.hasPosTagPart2(wdList, "pron")
	hasPron := false
	for _, tw := range wdList {
		if strings.Contains(tw.PosTag, "pron") {
			hasPron = true
			break
		}
	}
	if !hasPron {
		return nil
	}
	// Java PosTagHelper.adjust(wdList, null, "-"+group2, ":bad")
	suffix := "-небудь"
	var out []*languagetool.AnalyzedToken
	for _, tw := range wdList {
		pos := tw.PosTag
		if pos != "" && !strings.Contains(pos, ":bad") {
			pos = pos + ":bad"
		}
		lemma := tw.Lemma
		if lemma == "" {
			lemma = base
		}
		if !strings.HasSuffix(lemma, suffix) {
			lemma = lemma + suffix
		}
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
	}
	return out
}

// NebudSoftReadings is a deprecated name for NebudMissingHyphenReadings without dict
// (returns nil — fail closed). Prefer NebudMissingHyphenReadings with TagWord.
func NebudSoftReadings(surface, hyphenated string) []*languagetool.AnalyzedToken {
	return NebudMissingHyphenReadings(surface, hyphenated, nil)
}
