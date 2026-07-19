package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Dynamic numeric compounds: 100-й, 50-х, 100-річному, 100-відсотково
// (Java CompoundTagger.matchDigitCompound + LetterEndingForNumericHelper).
// reNumLeft ports ADJ_PREFIX_NUMBER / digit left of matchDigitCompound (simplified).
var reNumLeft = regexp.MustCompile(`^[0-9]+([,][0-9]+)?([-–—][0-9]+([,][0-9]+)?)?%?$`)

// DynamicNumericReadings tags digit-hyphen ordinal/adj endings.
// Short letter endings use official LetterEndingForNumericHelper maps only (no invent cases).
// Longer right halves (100-річному) require tagWord hits (Java wordTagger) — fail-closed without dict.
func DynamicNumericReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")

	dash := strings.LastIndex(t, "-")
	if dash <= 0 || dash == len(t)-1 {
		return nil
	}
	leftWord := t[:dash]
	rightWord := t[dash+1:]
	if !reNumLeft.MatchString(leftWord) {
		return nil
	}

	// Java: LetterEndingForNumericHelper.findTagsAdj
	if tags := FindTagsAdj(leftWord, strings.ToLower(rightWord)); len(tags) > 0 {
		lemma := leftWord + "-й"
		var out []struct{ Lemma, POS string }
		for _, tag := range tags {
			posTag := tag
			if strings.Contains(posTag, ":bad") {
				posTag = strings.Replace(posTag, ":bad", ":numr:bad", 1)
			} else {
				posTag = posTag + ":numr"
			}
			out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: "adj" + posTag})
		}
		return out
	}

	// Java: findTagsNoun → numr + tag (bad forms)
	if tags := FindTagsNoun(leftWord, strings.ToLower(rightWord)); len(tags) > 0 {
		var out []struct{ Lemma, POS string }
		for _, tag := range tags {
			out = append(out, struct{ Lemma, POS string }{Lemma: leftWord, POS: "numr" + tag})
		}
		return out
	}

	// Java: e.g. 100-річному, 100-відсотково — right from wordTagger only
	if tagWord == nil {
		return nil
	}
	rightLow := strings.ToLower(rightWord)
	tws := tagWord(rightWord)
	if len(tws) == 0 && rightWord != rightLow {
		tws = tagWord(rightLow)
	}
	if len(tws) == 0 {
		return nil
	}
	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		// prefer adj / adv (Java matchDigitCompound filters)
		if !strings.HasPrefix(pos, "adj") && !strings.HasPrefix(pos, "adv") && !strings.HasPrefix(pos, "noun") {
			continue
		}
		lemmaRight := tw.Lemma
		if lemmaRight == "" {
			lemmaRight = rightLow
		}
		lemma := leftWord + "-" + lemmaRight
		key := lemma + "|" + pos
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}

// MissingApostropheCandidates inserts ' before ї/є/ю/я after a consonant.
func MissingApostropheCandidates(token string) []string {
	if strings.Contains(token, "'") || strings.Contains(token, "’") {
		return nil
	}
	rs := []rune(token)
	var out []string
	needApo := map[rune]bool{'ї': true, 'є': true, 'ю': true, 'я': true, 'Ї': true, 'Є': true, 'Ю': true, 'Я': true}
	consonants := "бвгґджзклмнпрстфхцчшщБВГҐДЖЗКЛМНПРСТФХЦЧШЩ"
	for i := 1; i < len(rs); i++ {
		if !needApo[rs[i]] {
			continue
		}
		if !strings.ContainsRune(consonants, rs[i-1]) {
			continue
		}
		out = append(out, string(rs[:i])+"'"+string(rs[i:]))
	}
	return out
}

