package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java NUMR_ADJ_PATTERN: left ends with одно|дво|ох|и (numeral stems for adj compounds).
var reNumrAdjLeft = regexp.MustCompile(`(?i).+?(одно|дво|ох|и)$`)

// DynamicNumrAdjReadings ports CompoundTagger.numrAdjMatch (дво-триметровий…).
// Left must tag as numr in wordTagger; right must be adj. Fail-closed without dict.
func DynamicNumrAdjReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
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
	if tagging.UTF16Len(leftWord) < 2 {
		return nil
	}
	if !reNumrAdjLeft.MatchString(leftWord) {
		return nil
	}

	// Java: wordTagger.tag(leftWord) must have numr
	leftTags := tagWord(leftWord)
	if len(leftTags) == 0 {
		low := strings.ToLower(leftWord)
		if low != leftWord {
			leftTags = tagWord(low)
		}
	}
	hasNumr := false
	for _, tw := range leftTags {
		if strings.HasPrefix(tw.PosTag, "numr") {
			hasNumr = true
			break
		}
	}
	if !hasNumr {
		return nil
	}

	// right adj from dict
	rightTags := tagWord(rightWord)
	if len(rightTags) == 0 {
		low := strings.ToLower(rightWord)
		if low != rightWord {
			rightTags = tagWord(low)
		}
	}
	if len(rightTags) == 0 {
		return nil
	}

	leftLow := strings.ToLower(leftWord)
	extraBad := false
	// двох-трьохметровий - bad
	if regexp.MustCompile(`(?i).*(двох|трьох|чотирьох)$`).MatchString(leftLow) {
		extraBad = true
	} else if len(rightTags) > 0 {
		// три-метровий - bad when right does not start with (дво|три|…)
		rt := strings.ToLower(rightWord)
		if !regexp.MustCompile(`(?i)^(дво|три|чотири|п'яти|шести|семи|вісьми|двох|трьох|чотирьох).+`).MatchString(rt) {
			extraBad = true
		}
	}

	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range rightTags {
		pos := tw.PosTag
		if pos == "" || !strings.HasPrefix(pos, "adj") {
			continue
		}
		if i := strings.Index(pos, ":comp"); i >= 0 {
			end := i + len(":comp")
			for end < len(pos) && pos[end] != ':' {
				end++
			}
			pos = pos[:i] + pos[end:]
		}
		if extraBad && !strings.Contains(pos, ":bad") {
			pos = pos + ":bad"
		}
		lemRight := tw.Lemma
		if lemRight == "" {
			lemRight = strings.ToLower(rightWord)
		}
		lemma := leftLow + "-" + lemRight
		key := lemma + "|" + pos
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}
