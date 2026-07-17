package uk

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// FullTagMatchReadings tags A-B when both sides share a POS family (soft full-tag match).
// tagWord is the dictionary lookup (inject MapWordTagger ok).
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

	lowL, lowR := strings.ToLower(left), strings.ToLower(right)

	leftTags := lookupBothCases(left, tagWord)
	rightTags := lookupBothCases(right, tagWord)
	if len(leftTags) == 0 || len(rightTags) == 0 {
		// soft redup / adv-adv without dict
		if lowL == lowR && len([]rune(lowL)) <= 5 && isCyrillicWord(lowL) {
			lemma := lowL + "-" + lowR
			if isElongatedVowelRun(lowL) || len([]rune(lowL)) <= 2 {
				p, l := "intj", lemma
				return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, &p, &l)}
			}
			if strings.HasSuffix(lowL, "о") || strings.HasSuffix(lowL, "е") || strings.HasSuffix(lowL, "и") {
				p, l := "adv", lemma
				return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, &p, &l)}
			}
			p, l := "noninfl:predic", lemma
			return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, &p, &l)}
		}
		if strings.HasSuffix(lowL, "о") && strings.HasSuffix(lowR, "о") &&
			len([]rune(lowL)) > 3 && len([]rune(lowR)) > 2 &&
			unicode.IsLower([]rune(left)[0]) && unicode.IsLower([]rune(right)[0]) {
			p, l := "adv", lowL+"-"+lowR
			return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, &p, &l)}
		}
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

// NebudSoftReadings tags missing-hyphen -небудь forms as adj:…:pron:…:bad.
func NebudSoftReadings(surface, hyphenated string) []*languagetool.AnalyzedToken {
	// lemma: try map common bases
	// якого-небудь → який-небудь
	low := strings.ToLower(hyphenated)
	parts := strings.SplitN(low, "-небудь", 2)
	if len(parts) == 0 || parts[0] == "" {
		return nil
	}
	base := parts[0]
	// crude lemma normalize endings to -ий when looks like adj form
	lemma := base + "-небудь"
	if strings.HasSuffix(base, "ого") {
		lemma = strings.TrimSuffix(base, "ого") + "ий-небудь"
	} else if strings.HasSuffix(base, "ому") {
		lemma = strings.TrimSuffix(base, "ому") + "ий-небудь"
	} else if strings.HasSuffix(base, "им") {
		lemma = strings.TrimSuffix(base, "им") + "ий-небудь"
	}
	// case from ending
	cases := []string{":m:v_rod", ":m:v_zna:ranim", ":n:v_rod"}
	switch {
	case strings.HasSuffix(base, "ого"):
		cases = []string{":m:v_rod", ":m:v_zna:ranim", ":n:v_rod"}
	case strings.HasSuffix(base, "ому"):
		cases = []string{":m:v_dav", ":m:v_mis", ":n:v_dav", ":n:v_mis"}
	case strings.HasSuffix(base, "ий") || strings.HasSuffix(base, "ій"):
		cases = []string{":m:v_naz", ":m:v_zna:rinanim"}
	case strings.HasSuffix(base, "а") || strings.HasSuffix(base, "я"):
		cases = []string{":f:v_naz"}
	}
	var out []*languagetool.AnalyzedToken
	for _, c := range cases {
		p := "adj" + c + ":pron:int:rel:def:bad"
		l := lemma
		out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
	}
	return out
}
