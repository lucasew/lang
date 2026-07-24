package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java CompoundTagger.WORDS_WITH_YEAR / WORDS_WITH_NUM (official lists).
var wordsWithYear = map[string]struct{}{
	"бюджет": {}, "вибори": {}, "гра": {}, "держбюджет": {}, "кошторис": {}, "кампанія": {},
	"єврокубок": {}, "єврокваліфікація": {}, "євровідбір": {}, "єврофорум": {},
	"конкурс": {}, "кінофестиваль": {}, "кубок": {}, "мундіаль": {}, "м'яч": {},
	"олімпіада": {}, "оцінювання": {}, "оскар": {},
	"пектораль": {}, "перегони": {}, "першість": {}, "політреформа": {}, "премія": {},
	"рейтинг": {}, "реформа": {}, "сезон": {},
	"турнір": {}, "універсіада": {}, "фестиваль": {}, "форум": {},
	"чемпіонат": {}, "чемпіон": {}, "чемпіонка": {}, "ярмарок": {}, "ЧУ": {}, "ЧЄ": {},
}

var wordsWithNum = map[string]struct{}{
	"Формула": {}, "Карпати": {}, "Динамо": {}, "Шахтар": {}, "Фукусіма": {},
	"Квартал": {}, "Золоте": {}, "Мінськ": {}, "Нюренберг": {},
	"омега": {}, "плутоній": {}, "полоній": {}, "стронцій": {}, "уран": {}, "потік": {},
}

var (
	reYearNumber    = regexp.MustCompile(`^[12][0-9]{3}$`)
	reNounPrefixNum = regexp.MustCompile(`^[0-9]+$`)
)

// DynamicBadSuffixReadings ports BAD_SUFFIX (був-би, … but not м-б).
// Left from dict; lemma gets "-" + right; POS gets :bad.
func DynamicBadSuffixReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if tagging.UTF16Len(leftWord) <= 1 {
		return nil
	}
	// Java BAD_SUFFIX.contains(rightWord) — list entries are lowercase
	if _, ok := map[string]struct{}{"б": {}, "би": {}, "ж": {}, "же": {}}[rightWord]; !ok {
		return nil
	}

	leftTags := lookupBothCases(leftWord, tagWord)
	if len(leftTags) == 0 {
		return nil
	}
	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range leftTags {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		if !strings.Contains(pos, ":bad") {
			pos = pos + ":bad"
		}
		lem := tw.Lemma
		if lem == "" {
			lem = leftWord
		}
		// Java adjust(left, null, "-"+right) — then addIfNotContains :bad on left list
		// (second step overwrites lemma suffix in Java; we keep both suffix + :bad)
		lemma := lem + "-" + rightWord
		key := lemma + "|" + pos
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}

// DynamicAlPrefixReadings ports left "аль" → tag "Аль-"+right with :bad.
func DynamicAlPrefixReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if leftWord != "аль" {
		return nil
	}
	// Java: wordTagger.tag("Аль-" + rightWord)
	lookup := "Аль-" + rightWord
	tws := tagWord(lookup)
	if len(tws) == 0 {
		tws = tagWord(strings.ToLower(lookup))
	}
	if len(tws) == 0 {
		return nil
	}
	var out []struct{ Lemma, POS string }
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		if !strings.Contains(pos, ":bad") {
			pos = pos + ":bad"
		}
		lem := tw.Lemma
		if lem == "" {
			lem = lookup
		}
		out = append(out, struct{ Lemma, POS string }{Lemma: lem, POS: pos})
	}
	return out
}

// DynamicYearCompoundReadings ports Вибори-2014 / budget-year compounds.
func DynamicYearCompoundReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if !reYearNumber.MatchString(rightWord) {
		return nil
	}

	leftTags := lookupBothCases(leftWord, tagWord)
	if len(leftTags) == 0 {
		return nil
	}
	isUpper := false
	if rs := []rune(leftWord); len(rs) > 0 {
		isUpper = unicode.IsUpper(rs[0])
	}

	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range leftTags {
		pos := tw.PosTag
		lem := tw.Lemma
		if lem == "" {
			lem = leftWord
		}
		// !:prop && !WORDS_WITH_YEAR → skip
		_, inYear := wordsWithYear[lem]
		if !strings.Contains(pos, ":prop") && !inYear {
			continue
		}
		if pos == "" || !strings.HasPrefix(pos, "noun:inanim") {
			continue
		}
		if strings.Contains(pos, "v_kly") {
			continue
		}
		// plural only for гра/бюджет or :ns
		if strings.Contains(pos, ":p:") && lem != "гра" && lem != "бюджет" && !strings.Contains(pos, ":ns") {
			continue
		}
		pos = strings.ReplaceAll(pos, ":geo", "")
		if !strings.Contains(pos, ":prop") {
			if isUpper {
				pos = pos + ":prop"
				lem = capitalizeFirst(lem)
			}
		}
		lemma := lem + "-" + rightWord
		key := lemma + "|" + pos
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}

// DynamicNumSuffixCompoundReadings ports Формула-1 / омега-3 (WORDS_WITH_NUM).
func DynamicNumSuffixCompoundReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if !reNounPrefixNum.MatchString(rightWord) {
		return nil
	}

	leftTags := lookupBothCases(leftWord, tagWord)
	if len(leftTags) == 0 {
		return nil
	}
	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range leftTags {
		pos := tw.PosTag
		lem := tw.Lemma
		if lem == "" {
			lem = leftWord
		}
		if pos == "" || !strings.HasPrefix(pos, "noun:inanim") {
			continue
		}
		if strings.Contains(pos, "v_kly") {
			continue
		}
		// Java: if not :prop and lemma not in WORDS_WITH_NUM → add :prop + capitalize
		if !strings.Contains(pos, ":prop") {
			if _, ok := wordsWithNum[lem]; !ok {
				pos = pos + ":prop"
				lem = capitalizeFirst(lem)
			}
		}
		// Java: final gate — lemma must be in WORDS_WITH_NUM
		if _, ok := wordsWithNum[lem]; !ok {
			continue
		}
		lemma := lem + "-" + rightWord
		key := lemma + "|" + pos
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: lemma, POS: pos})
	}
	return out
}

func capitalizeFirst(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}
