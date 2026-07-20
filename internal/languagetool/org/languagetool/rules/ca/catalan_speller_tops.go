package ca

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Patterns from MorfologikCatalanSpellerRule (CASE_INSENSITIVE | UNICODE_CASE).
// Java [90]? is optional apostrophe-substitute digits between clitic and stem (non-capturing).
var (
	// Go RE2: (?i) only (no (?iu)); \p{L} for letters.
	caApostrofIniciVerbs     = regexp.MustCompile(`(?i)^([lnts])[90]?(h?[aeiouàéèíòóú].*)$`)
	caApostrofIniciVerbsM    = regexp.MustCompile(`(?i)^(m)[90]?(h?[aeiouàéèíòóú].*)$`)
	caApostrofIniciNomSing   = regexp.MustCompile(`(?i)^([ld])[90]?(h?[aeiouàéèíòóú]...+)$`)
	caApostrofIniciNomPlural = regexp.MustCompile(`(?i)^(d)[90]?(h?[aeiouàéèíòóú].+)$`)
	caApostrofFinal          = regexp.MustCompile(`(?i)^(...+[aei])[90]?(l|ls|m|ns|n|t)$`)
	caApostrofFinalS         = regexp.MustCompile(`(?i)^(.+e)[90]?(s)$`)
	caGuionetFinal           = regexp.MustCompile(`(?i)^([\p{L}·]+)['’]?(hi|ho|la|les|li|lo|los|me|ne|nos|se|te|vos)$`)
	caGuionetFinalGerundi    = regexp.MustCompile(`(?i)^([\p{L}·]+n)(hi|ho|la|les|li|lo|los|me|ne|nos|se|te|vos)$`)

	// Catalan POS: Java '.' is any char (unescaped).
	caVerbIndSubj   = regexp.MustCompile(`(?i)^V.[SI].*`)
	caVerbIndSubjM  = regexp.MustCompile(`(?i)^V.[SI].[123]S.*|^V.[SI].[23]P.*`)
	caNomSing       = regexp.MustCompile(`(?i)^V.[NG].*|^V.P..S..|^N..[SN].*|^A...[SN].|^PX..S...|^DD..S.`)
	caNomPlural     = regexp.MustCompile(`(?i)^V.P..P..|^N..[PN].*|^A...[PN].|^PX..P...|^DD..P.`)
	caVerbInfGerImp = regexp.MustCompile(`(?i)^V.[NGM].*`)
	caVerbInf       = regexp.MustCompile(`(?i)^V.N.*`)
	caVerbGer       = regexp.MustCompile(`(?i)^V.G.*`)
)

// Java SPLIT_DIGITS_AT_END.
var caSplitDigitsAtEnd = map[string]struct{}{
	"en": {}, "de": {}, "del": {}, "al": {}, "dels": {}, "als": {},
	"a": {}, "i": {}, "o": {}, "amb": {},
}

// Java PronomsDarrere (longest-first for endsWith).
var caPronomsDarrere = []string{
	"losels", "losles", "nosels", "nosles", "vosels", "vosens", "vosles",
	"lesen", "leshi", "liles", "losel", "losen", "loshi", "losho", "losla",
	"lsels", "lsles", "meles", "nosel", "nosen", "noshi", "nosho", "nosla", "nosli",
	"nsels", "nsles", "seles", "sevos", "teles", "usels", "usens", "usles",
	"vosel", "vosem", "vosen", "voshi", "vosho", "vosla", "vosli",
	"lahi", "lihi", "liho", "lila", "lils", "lsel", "lsen", "lshi", "lsho", "lsla",
	"mela", "meli", "mels", "nsel", "nsen", "nshi", "nsho", "nsla", "nsli",
	"sela", "seli", "sels", "sens", "seus", "tela", "teli", "tels", "tens",
	"usel", "usem", "usen", "ushi", "usho", "usla", "usli",
	"lan", "len", "les", "lhi", "lil", "lin", "los", "mel", "men", "mhi", "mho", "nhi",
	"nos", "sel", "sem", "sen", "set", "shi", "sho", "tel", "tem", "ten", "thi", "tho", "vos",
	"hi", "ho", "la", "li", "lo", "ls", "me", "ne", "ns", "se", "te", "us",
	"mi", "nosi", "losi", "si", "lis",
}

// additionalTopCatalanSuggestions ports getAdditionalTopSuggestionsString.
// Apostrophe/hyphen/multi-pronoun arms need TagPOS (fail-closed when nil).
func (r *MorfologikCatalanSpellerRule) additionalTopCatalanSuggestions(word string) []string {
	if r == nil || word == "" {
		return nil
	}
	// camelCase split
	if parts := splitCamelCaseCA(word); len(parts) > 1 && utf8.RuneCountInString(parts[0]) > 1 {
		ok := true
		for _, p := range parts {
			if r.wordIsMisspelled(p) {
				ok = false
				break
			}
		}
		if ok {
			return []string{strings.Join(parts, " ")}
		}
	}
	// digit split
	if parts := splitDigitsAtEndCA(word); len(parts) == 2 {
		p0 := parts[0]
		_, shortOK := caSplitDigitsAtEnd[strings.ToLower(p0)]
		if r.isTaggedCA(p0) && (utf8.RuneCountInString(p0) > 2 || shortOK) {
			return []string{strings.Join(parts, " ")}
		}
	}
	// apostrophe / hyphen / multi-pronoun (TagPOS required)
	if s := r.apostropheHyphenTopSuggestion(word); s != "" {
		return []string{s}
	}
	return nil
}

// apostropheHyphenTopSuggestion ports the findSuggestion chain + findSuggestionMultiplePronouns.
func (r *MorfologikCatalanSpellerRule) apostropheHyphenTopSuggestion(word string) string {
	if r == nil || r.TagPOS == nil || word == "" {
		return ""
	}
	type arm struct {
		wp, pp *regexp.Regexp
		pos    int
		sep    string
		add    string
	}
	arms := []arm{
		{caApostrofIniciVerbs, caVerbIndSubj, 2, "'", ""},
		{caApostrofIniciVerbsM, caVerbIndSubjM, 2, "'", ""},
		{caApostrofIniciNomSing, caNomSing, 2, "'", ""},
		{caApostrofIniciNomPlural, caNomPlural, 2, "'", ""},
		{caApostrofFinal, caVerbInfGerImp, 1, "'", ""},
		{caApostrofFinalS, caVerbInf, 1, "'", ""},
		{caGuionetFinalGerundi, caVerbGer, 1, "-", "t"},
		{caGuionetFinal, caVerbInfGerImp, 1, "-", ""},
	}
	for _, a := range arms {
		if s := r.findSuggestionCA(word, a.wp, a.pp, a.pos, a.sep, a.add); s != "" {
			return s
		}
	}
	return r.findSuggestionMultiplePronouns(word)
}

// findSuggestionCA ports MorfologikCatalanSpellerRule.findSuggestion (non-recursive).
// Returns matcher.group(1)+addStr+separator+matcher.group(2) when POS matches.
func (r *MorfologikCatalanSpellerRule) findSuggestionCA(
	word string,
	wordPattern, postagPattern *regexp.Regexp,
	suggestionPosition int,
	separator, addStr string,
) string {
	if r == nil || wordPattern == nil || postagPattern == nil {
		return ""
	}
	m := wordPattern.FindStringSubmatch(word)
	if m == nil || len(m) <= suggestionPosition {
		return ""
	}
	newSuggestion := m[suggestionPosition] + addStr
	tags := r.tagPOSCA(newSuggestion)
	// Java: (!hasPosTag("VMIP1S0B") || fer|ajust|gran) && matchPostagRegexp
	if hasExactPOS(tags, "VMIP1S0B") &&
		!strings.EqualFold(newSuggestion, "fer") &&
		!strings.EqualFold(newSuggestion, "ajust") &&
		!strings.EqualFold(newSuggestion, "gran") {
		return ""
	}
	if !matchPostagRegexpCA(tags, postagPattern) {
		return ""
	}
	if len(m) < 3 {
		return ""
	}
	return m[1] + addStr + separator + m[2]
}

// findSuggestionMultiplePronouns ports anarsen → anar-se'n / danarsen → d'anar-se'n.
func (r *MorfologikCatalanSpellerRule) findSuggestionMultiplePronouns(word string) string {
	if r == nil || r.TagPOS == nil || word == "" {
		return ""
	}
	lcword := strings.ToLower(word)
	pronouns := endsWithPronounCA(lcword)
	if pronouns == "" {
		return ""
	}
	// Java: verb = lcword.substring(0, word.length() - pronouns.length())
	// uses original word length for cut (same as lcword when case differs by case only).
	if len(pronouns) >= len(lcword) {
		return ""
	}
	verb := lcword[:len(lcword)-len(pronouns)]
	if matchPostagRegexpCA(r.tagPOSCA(verb), caVerbInfGerImp) {
		return verb + TransformDarrere(pronouns, verb)
	}
	if len(verb) < 5 {
		return ""
	}
	if strings.HasPrefix(lcword, "d") || strings.HasPrefix(lcword, "l") {
		verb2 := verb[1:]
		if matchPostagRegexpCA(r.tagPOSCA(verb2), caVerbInf) {
			return string(lcword[0]) + "'" + verb2 + TransformDarrere(pronouns, verb2)
		}
	}
	return ""
}

func endsWithPronounCA(s string) string {
	for _, p := range caPronomsDarrere {
		if strings.HasSuffix(s, p) {
			return p
		}
	}
	return ""
}

func matchPostagRegexpCA(tags []string, re *regexp.Regexp) bool {
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

func hasExactPOS(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func (r *MorfologikCatalanSpellerRule) tagPOSCA(word string) []string {
	if r == nil || r.TagPOS == nil {
		return nil
	}
	return r.TagPOS(word)
}

func (r *MorfologikCatalanSpellerRule) wordIsMisspelled(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if r.IsMisspelled != nil {
		return r.IsMisspelled(word)
	}
	if r.Speller != nil {
		return r.Speller.IsMisspelled(word)
	}
	return false
}

func (r *MorfologikCatalanSpellerRule) isTaggedCA(word string) bool {
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

// splitCamelCaseCA delegates to tools.SplitCamelCase (StringTools).
func splitCamelCaseCA(word string) []string {
	if word == "" {
		return nil
	}
	return tools.SplitCamelCase(word)
}

// splitDigitsAtEndCA delegates to tools.SplitDigitsAtEnd (StringTools).
func splitDigitsAtEndCA(input string) []string {
	if input == "" {
		return nil
	}
	return tools.SplitDigitsAtEnd(input)
}
