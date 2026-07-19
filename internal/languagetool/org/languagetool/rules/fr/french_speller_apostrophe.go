package fr

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Patterns from MorfologikFrenchSpellerRule (CASE_INSENSITIVE | UNICODE_CASE).
var (
	// Go RE2: (?i) only (no (?u)); Unicode letter class via \p{L}.
	apostrofIniciVerbs     = regexp.MustCompile(`(?i)^([lnts])(h?[aeiouàéèíòóú].*[^è])$`)
	apostrofIniciVerbsM    = regexp.MustCompile(`(?i)^(m)(h?[aeiouàéèíòóú].*[^è])$`)
	apostrofIniciVerbsC    = regexp.MustCompile(`(?i)^(c)([eiéèê].*)$`)
	apostrofIniciVerbsInf  = regexp.MustCompile(`(?i)^([lntsmd]|nous|vous)(h?[aeiouàéèíòóú].*[^è])$`)
	apostrofIniciNomSing   = regexp.MustCompile(`(?i)^([ld])(h?[aeiouàéèíòóú]...+)$`)
	apostrofIniciNomPlural = regexp.MustCompile(`(?i)^(d)(h?[aeiouàéèíòóú].+)$`)
	imperativeHyphen       = regexp.MustCompile(`(?i)^([\p{L}]+)['’]?(moi|toi|le|la|lui|nous|vous|les|leur|y|en)$`)
	hyphenOn               = regexp.MustCompile(`(?i)^([\p{L}]+[^aeiou])['’]?(il|elle|ce|on)$`)
	hyphenJe               = regexp.MustCompile(`(?i)^([\p{L}]+[^e])['’]?(je)$`)
	hyphenTu               = regexp.MustCompile(`(?i)^([\p{L}]+)['’]?(tu)$`)
	hyphenNous             = regexp.MustCompile(`(?i)^([\p{L}]+)['’]?(nous)$`)
	hyphenVous             = regexp.MustCompile(`(?i)^([\p{L}]+)['’]?(vous)$`)
	hyphenIls              = regexp.MustCompile(`(?i)^([\p{L}]+)['’]?(ils|elles)$`)

	// POS patterns (Java: "V .*(ind|sub).*" etc. — space and dots are literal-ish in LT tags)
	verbIndSubj  = regexp.MustCompile(`(?i)^V .*(ind|sub).*`)
	verbIndSubjM = regexp.MustCompile(`(?i)^V .* [123] s$|^V .* [23] p$`)
	verbIndSubjC = regexp.MustCompile(`(?i)^V .* 3 s$`)
	verbInf      = regexp.MustCompile(`(?i)^V.* inf`)
	verbImp      = regexp.MustCompile(`(?i)^V.* imp .*`)
	nomSing      = regexp.MustCompile(`(?i)^[NJZ] .* (s|sp)$|^V .inf$|^V .*ppa.* s$`)
	nomPlural    = regexp.MustCompile(`(?i)^[NJZ] .* (p|sp)$|^V .*ppa.* p$`)
	verb1S       = regexp.MustCompile(`(?i)^V .*(ind).* 1 s$`)
	verb2S       = regexp.MustCompile(`(?i)^V .*(ind).* 2 s$`)
	verb3S       = regexp.MustCompile(`(?i)^V .*(ind).* 3 s$`)
	verb1P       = regexp.MustCompile(`(?i)^V .*(ind).* 1 p$`)
	verb2P       = regexp.MustCompile(`(?i)^V .*(ind).* 2 p$`)
	verb3P       = regexp.MustCompile(`(?i)^V .*(ind).* 3 p$`)
)

var frenchSplitDigitsAtEnd = map[string]struct{}{
	"et": {}, "ou": {}, "de": {}, "en": {}, "à": {}, "aux": {}, "des": {},
}

// matchPostagRegexp ports matchPostagRegexp: any reading POS matches pattern (null → "UNKNOWN").
func matchPostagRegexp(tags []string, re *regexp.Regexp) bool {
	if re == nil {
		return false
	}
	if len(tags) == 0 {
		return re.MatchString("UNKNOWN")
	}
	for _, pos := range tags {
		if pos == "" {
			pos = "UNKNOWN"
		}
		if re.MatchString(pos) {
			return true
		}
	}
	return false
}

// findSuggestion ports findSuggestion (recursive spell-suggest on the target group).
func (r *MorfologikFrenchSpellerRule) findSuggestion(
	word string,
	wordPattern, postagPattern *regexp.Regexp,
	suggestionPosition int,
	separator string,
	recursive bool,
) []string {
	if r == nil || wordPattern == nil || postagPattern == nil {
		return nil
	}
	m := wordPattern.FindStringSubmatch(word)
	if m == nil || len(m) <= suggestionPosition {
		return nil
	}
	newSuggestion := m[suggestionPosition]
	tags := r.tagPOS(newSuggestion)
	if matchPostagRegexp(tags, postagPattern) {
		// matcher.group(1) + separator + matcher.group(2)
		if len(m) >= 3 {
			return []string{m[1] + separator + m[2]}
		}
		return nil
	}
	if !recursive {
		return nil
	}
	more := r.wordSuggestions(newSuggestion)
	var out []string
	for i, ms := range more {
		if i > 5 {
			break
		}
		var newWord string
		if suggestionPosition == 1 {
			newWord = ms + m[2]
		} else {
			newWord = m[1] + ms
		}
		moreSugg := r.findSuggestion(newWord, wordPattern, postagPattern, suggestionPosition, separator, false)
		out = append(out, moreSugg...)
	}
	return out
}

func (r *MorfologikFrenchSpellerRule) tagPOS(word string) []string {
	if r == nil || r.TagPOS == nil {
		return nil
	}
	return r.TagPOS(word)
}

func (r *MorfologikFrenchSpellerRule) wordSuggestions(word string) []string {
	if r == nil || word == "" {
		return nil
	}
	if r.Speller != nil {
		if s := r.Speller.FindReplacements(word); len(s) > 0 {
			return s
		}
	}
	if FilterDictAvailable() {
		return FilterDictSuggest(word)
	}
	return nil
}

// apostropheHyphenTopSuggestions ports findSuggestion chain in getAdditionalTopSuggestionsString.
// Requires TagPOS; returns nil when unset (fail-closed).
func (r *MorfologikFrenchSpellerRule) apostropheHyphenTopSuggestions(word string) []string {
	if r == nil || r.TagPOS == nil || word == "" {
		return nil
	}
	var out []string
	type arm struct {
		wp, pp *regexp.Regexp
		pos    int
		sep    string
	}
	arms := []arm{
		{apostrofIniciVerbs, verbIndSubj, 2, "'"},
		{apostrofIniciVerbsM, verbIndSubjM, 2, "'"},
		{apostrofIniciVerbsC, verbIndSubjC, 2, "'"},
		{apostrofIniciVerbsInf, verbInf, 2, "'"},
		{apostrofIniciNomSing, nomSing, 2, "'"},
		{apostrofIniciNomPlural, nomPlural, 2, "'"},
		{imperativeHyphen, verbImp, 1, "-"},
		{hyphenJe, verb1S, 1, "-"},
		{hyphenTu, verb2S, 1, "-"},
		{hyphenOn, verb3S, 1, "-"},
		{hyphenNous, verb1P, 1, "-"},
		{hyphenVous, verb2P, 1, "-"},
		{hyphenIls, verb3P, 1, "-"},
	}
	for _, a := range arms {
		out = append(out, r.findSuggestion(word, a.wp, a.pp, a.pos, a.sep, true)...)
	}
	return out
}

// splitDigitsAtEnd ports StringTools.splitDigitsAtEnd.
func splitDigitsAtEnd(input string) []string {
	if input == "" {
		return nil
	}
	runes := []rune(input)
	last := len(runes) - 1
	for last >= 0 && unicode.IsDigit(runes[last]) {
		last--
	}
	nonDigit := string(runes[:last+1])
	digit := string(runes[last+1:])
	if nonDigit != "" && digit != "" {
		return []string{nonDigit, digit}
	}
	return []string{input}
}

// isTagged reports whether TagPOS returned any non-empty POS (Java isTagged).
func (r *MorfologikFrenchSpellerRule) isTagged(word string) bool {
	if r == nil || r.TagPOS == nil {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if t != "" {
			return true
		}
	}
	return false
}

// digitSplitTopSuggestion ports splitDigitsAtEnd arm.
func (r *MorfologikFrenchSpellerRule) digitSplitTopSuggestion(word string) string {
	parts := splitDigitsAtEnd(word)
	if len(parts) != 2 {
		return ""
	}
	p0 := parts[0]
	if !r.isTagged(p0) {
		return ""
	}
	if utf8.RuneCountInString(p0) > 2 {
		return strings.Join(parts, " ")
	}
	if _, ok := frenchSplitDigitsAtEnd[strings.ToLower(p0)]; ok {
		return strings.Join(parts, " ")
	}
	return ""
}
