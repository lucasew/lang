package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// tagsToRemove ports CompoundTagger.TAGS_TO_REMOVE (:comp.|:predic|:insert).
var reTagsToRemove = regexp.MustCompile(`:comp[^:]*|:predic|:insert`)

// reNapivDual ports напів(.+?)-напів(.+)
var reNapivDual = regexp.MustCompile(`(?i)^напів(.+?)-напів(.+)$`)

// DynamicPreRedupReadings ports гірко-прегірко / гіркий-прегіркий
// (right starts with "пре" and remainder equals left). Dict-gated on left POS.
func DynamicPreRedupReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	rightLow := strings.ToLower(rightWord)
	if !strings.HasPrefix(rightLow, "пре") {
		return nil
	}
	// right after "пре" (3 runes) must equal left (case-insensitive)
	rs := []rune(rightWord)
	if len(rs) < 4 {
		return nil
	}
	rest := string(rs[3:])
	if !strings.EqualFold(leftWord, rest) {
		return nil
	}

	leftTags := lookupBothCases(leftWord, tagWord)
	if len(leftTags) == 0 {
		return nil
	}

	hasAdv, hasAdj := false, false
	for _, tw := range leftTags {
		if strings.HasPrefix(tw.PosTag, "adv") {
			hasAdv = true
		}
		if strings.HasPrefix(tw.PosTag, "adj") {
			hasAdj = true
		}
	}

	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	if hasAdv {
		// Java: lemma = word (full surface), POS stripped of comp/predic/insert
		for _, tw := range leftTags {
			if !strings.HasPrefix(tw.PosTag, "adv") {
				continue
			}
			pos := reTagsToRemove.ReplaceAllString(tw.PosTag, "")
			key := token + "|" + pos
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, struct{ Lemma, POS string }{Lemma: token, POS: pos})
		}
		if len(out) > 0 {
			return out
		}
	}
	if hasAdj {
		// Java: lemma = lemma + "-пре" + lemma
		for _, tw := range leftTags {
			if !strings.HasPrefix(tw.PosTag, "adj") {
				continue
			}
			pos := reTagsToRemove.ReplaceAllString(tw.PosTag, "")
			lem := tw.Lemma
			if lem == "" {
				lem = leftWord
			}
			lemma := lem + "-пре" + lem
			key := lemma + "|" + pos
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
		}
	}
	return out
}

// DynamicNapivDualReadings ports напівпольської-напіванглійської.
// Tags bases via wordTagger, prefixes lemma with "напів", then FullTagMatch-style merge.
// Fail-closed without both-side dict hits.
func DynamicNapivDualReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" {
		return nil
	}
	m := reNapivDual.FindStringSubmatch(token)
	if m == nil {
		return nil
	}
	baseL, baseR := m[1], m[2]
	if baseL == "" || baseR == "" {
		return nil
	}

	// Java: adjust(tag(base), "напів", null) — lemma prefix
	leftTags := lookupBothCases(baseL, tagWord)
	rightTags := lookupBothCases(baseR, tagWord)
	if len(leftTags) == 0 || len(rightTags) == 0 {
		return nil
	}

	// Reuse FullTagMatch by building a synthetic surface from bases isn't right;
	// merge families like FullTagMatchReadings with lemma = напів+L - напів+R
	var out []struct{ Lemma, POS string }
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
			caseL, caseR := caseMarker(lt.PosTag), caseMarker(rt.PosTag)
			if caseL != "" && caseR != "" && caseL != caseR {
				continue
			}
			pos := mergePOS(lt.PosTag, rt.PosTag)
			llem := lt.Lemma
			if llem == "" {
				llem = baseL
			}
			rlem := rt.Lemma
			if rlem == "" {
				rlem = baseR
			}
			// Java adjust prefix "напів" on each side then tagMatch combines
			lemma := "напів" + llem + "-напів" + rlem
			key := pos + "|" + lemma
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			if !strings.Contains(pos, "prop") {
				lemma = strings.ToLower(lemma)
			}
			out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
		}
	}
	return out
}
