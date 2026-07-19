package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Dynamic numeric compounds: 100-й, 50-х, 100-річному, 100-відсотково, 100-річчя
// (Java CompoundTagger.matchDigitCompound + LetterEndingForNumericHelper).
// reNumLeft ports ADJ_PREFIX_NUMBER / digit left of matchDigitCompound (simplified; no Roman yet).
var reNumLeft = regexp.MustCompile(`^[0-9]+([,][0-9]+)?([-–—][0-9]+([,][0-9]+)?)?%?$`)

// Java CompoundTagger REQ_NUM_*_PATTERN + getTryPrefix.
var (
	reReqNumSto    = regexp.MustCompile(`(?i)^(річч|літт|метрів|грамов|тисячник).{0,3}$`)
	reReqNumDesyat = regexp.MustCompile(`(?i)^(класни[кц]|бальни[кц]|раундов|томн|томов|хвилин|десятиріч|кілометрів|річ).{0,4}$`)
	reReqNumDva    = regexp.MustCompile(`(?i)^(місн|томник|поверхів).{0,4}$`)
)

// getTryPrefix ports CompoundTagger.getTryPrefix — prefix a dict-known base onto the right
// half so "річчя" tags via "сторіччя", lemma strip back to "річчя".
func getTryPrefix(rightWord string) string {
	low := strings.ToLower(rightWord)
	switch {
	case reReqNumSto.MatchString(low):
		return "сто"
	case reReqNumDesyat.MatchString(low):
		return "десяти"
	case reReqNumDva.MatchString(low):
		return "дво"
	default:
		return ""
	}
}

// DynamicNumericReadings tags digit-hyphen ordinal/adj endings.
// Short letter endings use official LetterEndingForNumericHelper maps only (no invent cases).
// Longer right halves require tagWord hits (Java wordTagger) — fail-closed without dict.
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

	// Java: right "мм" → full adj paradigm (all genders × cases except v_kly); lemma = full word.
	if strings.EqualFold(rightWord, "мм") {
		return mmDigitAdjParadigm(token)
	}

	if tagWord == nil {
		return nil
	}
	rightLow := strings.ToLower(rightWord)

	// Java: 100-річчя via getTryPrefix + wordTagger.tag(prefix+right)
	if pref := getTryPrefix(rightLow); pref != "" {
		if out := numericTryPrefixReadings(leftWord, rightLow, pref, tagWord); len(out) > 0 {
			return out
		}
	}

	// Java: e.g. 100-річному, 100-відсотково — right from wordTagger;
	// only adj POS or lemma "відсотково" (not bare noun invent).
	tws := tagWord(rightWord)
	if len(tws) == 0 && rightWord != rightLow {
		tws = tagWord(rightLow)
	}
	if len(tws) == 0 {
		return nil
	}
	// Java: empty or has pron → null
	for _, tw := range tws {
		if strings.Contains(tw.PosTag, "pron") {
			return nil
		}
	}
	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		lemmaRight := tw.Lemma
		if lemmaRight == "" {
			lemmaRight = rightLow
		}
		// Java: adj* OR lemma "відсотково"
		if !strings.HasPrefix(pos, "adj") && lemmaRight != "відсотково" {
			continue
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

// mmDigitAdjParadigm ports matchDigitCompound "мм" branch (PosTagHelper genders × vidminky).
func mmDigitAdjParadigm(word string) []struct{ Lemma, POS string } {
	genders := []string{"m", "f", "n"}
	var out []struct{ Lemma, POS string }
	for _, g := range genders {
		for _, vid := range nvCases {
			out = append(out, struct{ Lemma, POS string }{
				Lemma: word,
				POS:   "adj:" + g + ":" + vid,
			})
		}
	}
	return out
}

// numericTryPrefixReadings ports getTryPrefix branch of matchDigitCompound.
// Java: wordTagger.tag(prefix+right); lemma = left + "-" + dictLemma[len(prefix):]
func numericTryPrefixReadings(leftWord, rightLow, pref string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	tws := tagWord(pref + rightLow)
	if len(tws) == 0 {
		return nil
	}
	prefRunes := []rune(pref)
	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		lemmaRight := rightLow
		if tw.Lemma != "" {
			rs := []rune(tw.Lemma)
			if len(rs) >= len(prefRunes) && strings.EqualFold(string(rs[:len(prefRunes)]), pref) {
				lemmaRight = string(rs[len(prefRunes):])
			} else {
				lemmaRight = tw.Lemma
			}
		}
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

