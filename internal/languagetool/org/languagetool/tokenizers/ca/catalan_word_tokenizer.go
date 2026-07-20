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

	// Java CatalanWordTokenizer.wordsToAdd camel-case hyphen exceptions only.
	javaHyphenExceptions = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true,
	}
)

// IsTaggedCA optional CatalanTagger.INSTANCE_CAT.tag(...).isTagged() hook.
// Java keeps hyphen compounds only when CatalanTagger tags them.
// Without a tagger, miss (split hyphens) — do not invent a soft compound lexicon.
var IsTaggedCA func(s string) bool

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

// wordsToAddCA ports CatalanWordTokenizer.wordsToAdd.
func wordsToAddCA(s string) []string {
	var l []string
	if s == "" {
		return l
	}
	if !strings.Contains(s, "-") && !strings.HasSuffix(s, "'") && !strings.HasSuffix(s, "’") {
		l = append(l, s)
		return l
	}
	// Java: CatalanTagger.INSTANCE_CAT.tag(...).isTagged()
	normalized := strings.ReplaceAll(s, "\u00AD", "")
	normalized = strings.ReplaceAll(normalized, "’", "'")
	if isTaggedCA(normalized) {
		l = append(l, s)
		return l
	}
	// Java camel-case hyphen exceptions
	if javaHyphenExceptions[strings.ToLower(s)] {
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
	// if not found, the word is split on hyphens (keep separators)
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
	// Java: CatalanTagger.INSTANCE_CAT.tag(...).isTagged(). Without a tagger, miss
	// (split hyphens) — do not invent a soft compound lexicon.
	if IsTaggedCA != nil {
		return IsTaggedCA(s)
	}
	return false
}
