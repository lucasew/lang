package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// reVesLemma ports LemmaHelper pattern [ув]?весь|[ву]с[еі] for equal redup compounds
// (Усе-усе, всього-всього, весь-весь).
var reVesLemma = regexp.MustCompile(`(?i)^([ув]?весь|[ву]с[еі])$`)

// equalParts ports CompoundTagger.equalParts: lemma has one hyphen and both sides equal.
func equalParts(lemma string) bool {
	if !strings.Contains(lemma, "-") {
		return false
	}
	parts := strings.SplitN(lemma, "-", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	return parts[0] == parts[1]
}

// DynamicEqualRedupReadings ports left.equalsIgnoreCase(right) + ves/us lemma redup
// filtered by equalParts after tagMatch. Fail-closed without dict on both sides.
func DynamicEqualRedupReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if !strings.EqualFold(leftWord, rightWord) {
		return nil
	}

	leftTags := lookupBothCases(leftWord, tagWord)
	rightTags := lookupBothCases(rightWord, tagWord)
	if len(leftTags) == 0 || len(rightTags) == 0 {
		return nil
	}

	// Java: LemmaHelper.hasLemma(left, [ув]?весь|[ву]с[еі])
	hasVes := false
	for _, tw := range leftTags {
		lem := tw.Lemma
		if lem == "" {
			lem = leftWord
		}
		if reVesLemma.MatchString(lem) || reVesLemma.MatchString(strings.ToLower(lem)) {
			hasVes = true
			break
		}
		// also allow surface form match (усе tagged with lemma увесь)
		if reVesLemma.MatchString(strings.ToLower(leftWord)) {
			hasVes = true
			break
		}
	}
	if !hasVes {
		return nil
	}

	// tagMatch-style merge, keep only equalParts lemmas
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
				llem = leftWord
			}
			rlem := rt.Lemma
			if rlem == "" {
				rlem = rightWord
			}
			lemma := llem + "-" + rlem
			if !equalParts(lemma) {
				// Java may still form lower surface redup adv/noun lemmas
				// equalParts is strict on lemma sides; also accept lower surface redup
				if !equalParts(strings.ToLower(leftWord) + "-" + strings.ToLower(rightWord)) {
					continue
				}
				// if lemma sides unequal but surface equal, use lower surface redup lemma when families match
				// Prefer lemma when equalParts; else skip (faithful to filter)
				continue
			}
			key := lemma + "|" + pos
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
