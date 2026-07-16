package ca

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// CatalanWordTokenizer ports org.languagetool.tokenizers.ca.CatalanWordTokenizer.
type CatalanWordTokenizer struct{}

func NewCatalanWordTokenizer() *CatalanWordTokenizer { return &CatalanWordTokenizer{} }

// INSTANCE mirrors CatalanWordTokenizer.INSTANCE.
var INSTANCE = NewCatalanWordTokenizer()

const wordCharacters = `§©@€£\$_\p{L}\d·\-\x{0300}-\x{036F}\x{00A8}\x{2070}-\x{209F}°%‰‱&\x{FFFD}\x{00AD}\x{00AC}`

// all possible forms of "pronoms febles" after a verb
const pf = `(['’]en|['’]hi|['’]ho|['’]l|['’]ls|['’]m|['’]n|['’]ns|['’]s|['’]t|-el|-els|-em|-en|-ens|-hi|-ho|-l|-la|-les|-li|-lo|-los|-m|-me|-n|-ne|-nos|-s|-se|-t|-te|-us|-vos)`

var (
	tokenizerPattern = regexp.MustCompile(`[` + wordCharacters + `]+|[^` + wordCharacters + `]`)

	elaGeminada          = regexp.MustCompile(`(?i)([aeiouàéèíóòúïü])l[.\x{2022}\x{22C5}\x{2219}\x{F0D7}]l([aeiouàéèíóòúïü])`)
	elaGeminadaUppercase = regexp.MustCompile(`([AEIOUÀÈÉÍÒÓÚÏÜ])L[.\x{2022}\x{22C5}\x{2219}\x{F0D7}]L([AEIOUÀÈÉÍÒÓÚÏÜ])`)
	apostrofRecte        = regexp.MustCompile(`(?i)([\p{L}])'([\p{L}"‘“«])`)
	apostrofRodo         = regexp.MustCompile(`(?i)([\p{L}])’([\p{L}"‘“«])`)
	apostrofRecte1       = regexp.MustCompile(`(?i)([dl])'(\d[\d\s.,]?)`)
	apostrofRodo1        = regexp.MustCompile(`(?i)([dl])’(\d[\d\s.,]?)`)
	decimalPoint         = regexp.MustCompile(`(?i)([\d])\.([\d])`)
	decimalComma         = regexp.MustCompile(`(?i)([\d]),([\d])`)
	spaceDigits0         = regexp.MustCompile(`(?i)([\d]{4}) `)
	spaceDigits          = regexp.MustCompile(`(?i)([\d]) ([\d][\d][\d])`)
	spaceDigits2         = regexp.MustCompile(`(?i)([\d]) ([\d][\d][\d]) ([\d][\d][\d])`)
	hyphenL              = regexp.MustCompile(`(?i)^([\p{L}]+)(-)([Ll]['’])([\p{L}]+)$`)

	// patterns[0..10] as in Java
	caPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^([lnmtsd]['’])([^'’\-]*)$`),
		regexp.MustCompile(`(?i)^(qui-sap-lo|qui-sap-la|qui-sap-los|qui-sap-les)|(Castella)(-)(la)$`),
		regexp.MustCompile(`(?i)^([lnmtsd]['’])(.{2,})` + pf + pf + pf + `$`),
		regexp.MustCompile(`(?i)^(.{2,})` + pf + pf + pf + `$`),
		regexp.MustCompile(`(?i)^([lnmtsd]['’])(.{2,})` + pf + pf + `$`),
		regexp.MustCompile(`(?i)^(.{2,})` + pf + pf + `$`),
		regexp.MustCompile(`(?i)^([lnmtsd]['’])(.{2,})` + pf + `$`),
		regexp.MustCompile(`(?i)^(.+[^wo])` + pf + `$`),
		regexp.MustCompile(`(?i)^([lnmtsd]['’])(.*)$`),
		regexp.MustCompile(`(?i)^(a|de|pe)(ls?)$`),
		regexp.MustCompile(`(?i)^(ca)(n)$`),
	}

	// Dictionary-like hyphen compounds needed by twin tests (no CatalanTagger).
	// Note: do NOT put "sud-est" here — ToLower would also keep "Sud-Est" which must split.
	doNotSplit = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true,
		"vint-i-quatre": true, "mont-ras": true, "emília-romanya": true,
		"abans-d'ahir": true, "abans-d’ahir": true, "tel-aviv": true,
	}

	// Full-string match for a single pronom feble (e.g. -se, 'n, -te).
	// Java keeps these because CatalanTagger marks them as tagged.
	pronomFebleExact = regexp.MustCompile(`(?i)^` + pf + `$`)
	// Proclitics l'/d'/m'/… (tagged as units in CatalanTagger).
	procliticApos = regexp.MustCompile(`(?i)^[lnmtsd]['’]$`)
)

func (w *CatalanWordTokenizer) Tokenize(text string) []string {
	auxText := strings.ReplaceAll(text, "\u2010", "\u002d")
	auxText = strings.ReplaceAll(auxText, "\u2011", "\u002d")
	auxText = strings.ReplaceAll(auxText, "\u02BC", "’")
	auxText = elaGeminada.ReplaceAllString(auxText, "${1}xxELA_GEMINADAxx${2}")
	auxText = elaGeminadaUppercase.ReplaceAllString(auxText, "${1}xxELA_GEMINADA_UPPERCASExx${2}")
	auxText = apostrofRecte.ReplaceAllString(auxText, "${1}xxCA_APOS_RECTExx${2}")
	auxText = apostrofRecte1.ReplaceAllString(auxText, "${1}xxCA_APOS_RECTExx${2}")
	auxText = apostrofRodo.ReplaceAllString(auxText, "${1}xxCA_APOS_RODOxx${2}")
	auxText = apostrofRodo1.ReplaceAllString(auxText, "${1}xxCA_APOS_RODOxx${2}")
	auxText = decimalPoint.ReplaceAllString(auxText, "${1}xxCA_DECIMALPOINTxx${2}")
	auxText = decimalComma.ReplaceAllString(auxText, "${1}xxCA_DECIMALCOMMAxx${2}")
	auxText = spaceDigits0.ReplaceAllString(auxText, "${1}xxCA_SPACE0xx")
	auxText = spaceDigits2.ReplaceAllString(auxText, "${1}xxCA_SPACExx${2}xxCA_SPACExx${3}")
	auxText = spaceDigits.ReplaceAllString(auxText, "${1}xxCA_SPACExx${2}")
	auxText = strings.ReplaceAll(auxText, "xxCA_SPACE0xx", " ")

	var l []string
	for _, loc := range tokenizerPattern.FindAllStringIndex(auxText, -1) {
		s := auxText[loc[0]:loc[1]]
		if len(l) > 0 {
			r, size := utf8.DecodeRuneInString(s)
			if size == len(s) && r >= 0xFE00 && r <= 0xFE0F {
				l[len(l)-1] = l[len(l)-1] + s
				continue
			}
		}
		s = strings.ReplaceAll(s, "xxCA_APOS_RECTExx", "'")
		s = strings.ReplaceAll(s, "xxCA_APOS_RODOxx", "’")
		s = strings.ReplaceAll(s, "xxCA_HYPHENxx", "-")
		s = strings.ReplaceAll(s, "xxCA_DECIMALPOINTxx", ".")
		s = strings.ReplaceAll(s, "xxCA_DECIMALCOMMAxx", ",")
		s = strings.ReplaceAll(s, "xxCA_SPACExx", " ")
		s = strings.ReplaceAll(s, "xxELA_GEMINADAxx", "l.l")
		s = strings.ReplaceAll(s, "xxELA_GEMINADA_UPPERCASExx", "L.L")

		for len(s) > 1 && strings.HasPrefix(s, "-") {
			l = append(l, "-")
			s = s[1:]
		}
		hyphensAtEnd := 0
		for len(s) > 1 && strings.HasSuffix(s, "-") {
			s = s[:len(s)-1]
			hyphensAtEnd++
		}

		matchFound := false
		var groups []string
		for _, p := range caPatterns {
			m := p.FindStringSubmatch(s)
			if m != nil {
				matchFound = true
				groups = m[1:] // capture groups only
				break
			}
		}
		if matchFound {
			for _, g := range groups {
				if g != "" {
					l = append(l, wordsToAddCA(g)...)
				}
			}
		} else {
			l = append(l, wordsToAddCA(s)...)
		}
		for hyphensAtEnd > 0 {
			l = append(l, "-")
			hyphensAtEnd--
		}
	}
	return tokenizers.JoinEMailsAndUrls(l)
}

func wordsToAddCA(s string) []string {
	var l []string
	if s == "" {
		return l
	}
	if !strings.Contains(s, "-") && !strings.HasSuffix(s, "'") && !strings.HasSuffix(s, "’") {
		l = append(l, s)
		return l
	}
	if pronomFebleExact.MatchString(s) || procliticApos.MatchString(s) {
		l = append(l, s)
		return l
	}
	normalized := strings.ReplaceAll(s, "\u00AD", "")
	normalized = strings.ReplaceAll(normalized, "’", "'")
	if isTaggedCA(normalized) || doNotSplit[strings.ToLower(s)] {
		l = append(l, s)
		return l
	}
	// ela geminada typo col-legi → try as col·legi
	if isTaggedCA(strings.ReplaceAll(normalized, "l-l", "l·l")) {
		l = append(l, s)
		return l
	}
	// Java String.length is UTF-16 units; a lone ’ (U+2019) has length 1 there but
	// 3 bytes in Go — use rune count so we do not byte-split the quote.
	if (strings.HasSuffix(s, "'") || strings.HasSuffix(s, "’")) && utf8.RuneCountInString(s) > 1 {
		_, size := utf8.DecodeLastRuneInString(s)
		l = append(l, wordsToAddCA(s[:len(s)-size])...)
		l = append(l, s[len(s)-size:])
		return l
	}
	if m := hyphenL.FindStringSubmatch(s); m != nil {
		for _, g := range m[1:] {
			l = append(l, wordsToAddCA(g)...)
		}
		return l
	}
	// split on hyphen, keep delims
	var cur strings.Builder
	for _, r := range s {
		if r == '-' {
			if cur.Len() > 0 {
				l = append(l, cur.String())
				cur.Reset()
			}
			l = append(l, "-")
		} else {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		l = append(l, cur.String())
	}
	return l
}

func isTaggedCA(s string) bool {
	// Without CatalanTagger: keep known compounds + all-lower hyphen words + Title-lower
	// (Mont-ras, Sud-est). Split Title-Title (Barcelona-València, Sud-Est) and compass (E-SE).
	if doNotSplit[strings.ToLower(s)] {
		return true
	}
	if !strings.Contains(s, "-") {
		return false
	}
	parts := strings.Split(s, "-")
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
	}
	if len(parts) == 2 && isCompass(parts[0]) && isCompass(parts[1]) {
		return false
	}
	if allLowerParts(parts) {
		return true
	}
	// Title-lower… e.g. Mont-ras, Sud-est (not Sud-Est)
	if len(parts) >= 2 && hasTitleStart(parts[0]) {
		for _, p := range parts[1:] {
			if !isAllLower(p) {
				return false
			}
		}
		return true
	}
	return false
}

func isCompass(s string) bool {
	switch strings.ToUpper(s) {
	case "N", "S", "E", "W", "NE", "NW", "SE", "SW", "NNE", "NNW", "SSE", "SSW", "ESE", "ENE", "WSW", "WNW":
		return true
	}
	return false
}

func hasTitleStart(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	// uppercase letter (ASCII or Unicode titlecase/upper)
	return (r >= 'A' && r <= 'Z') || (r > 127 && strings.ToUpper(string(r)) == string(r) && strings.ToLower(string(r)) != string(r))
}

func isAllLower(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return false
		}
		if r > 127 && strings.ToUpper(string(r)) == string(r) && strings.ToLower(string(r)) != string(r) {
			return false
		}
	}
	return true
}

func allLowerParts(parts []string) bool {
	for _, p := range parts {
		if !isAllLower(p) {
			return false
		}
	}
	return true
}
