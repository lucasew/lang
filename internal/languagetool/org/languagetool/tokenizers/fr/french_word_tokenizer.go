package fr

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// FrenchWordTokenizer ports org.languagetool.tokenizers.fr.FrenchWordTokenizer.
type FrenchWordTokenizer struct{}

func NewFrenchWordTokenizer() *FrenchWordTokenizer { return &FrenchWordTokenizer{} }

const wordCharacters = `§©@€£\$_\p{L}\d\-\x{0300}-\x{036F}\x{00A8}\x{2070}-\x{209F}°%‰‱&\x{FFFD}\x{00AD}\x{00AC}`

var (
	tokenizerPattern = regexp.MustCompile(`[` + wordCharacters + `]+|[^` + wordCharacters + `]`)
	typewriterApos   = regexp.MustCompile(`(?i)([\p{L}])'([\p{L}1"‘“«])`)
	typographicApos  = regexp.MustCompile(`(?i)([\p{L}])’([\p{L}1"‘“«])`)
	nearbyHyphens    = regexp.MustCompile(`(?i)([\p{L}])-([\p{L}])-([\p{L}])`)
	hyphens          = regexp.MustCompile(`(?i)([\p{L}])-([\p{L}\d])`)
	decimalPoint     = regexp.MustCompile(`(?i)([\d])\.([\d])`)
	decimalComma     = regexp.MustCompile(`(?i)([\d]),([\d])`)
	spaceDigits0     = regexp.MustCompile(`(?i)([\d]{4}) `)
	spaceDigits      = regexp.MustCompile(`(?i)([\d]) ([\d][\d][\d])\b`)
	spaceDigits2     = regexp.MustCompile(`(?i)([\d]) ([\d][\d][\d]) ([\d][\d][\d])\b`)

	doNotSplit = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true, "anti-ivg": true, "anti-uv": true,
		"anti-vih": true, "al-qaïda": true, "c'est-à-dire": true, "add-on": true, "add-ons": true,
		"rendez-vous": true, "garde-à-vous": true, "chez-eux": true, "chez-moi": true,
		"chez-nous": true, "chez-soi": true, "chez-toi": true, "chez-vous": true, "m'as-tu-vu": true,
		// Soft stand-in for FrenchTagger dictionary hits used by Java wordsToAdd.
		"strauss-kahn": true, "petit-déjeunes": true,
	}

	frPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(c['’]te?|m['’]as-tu-vu|c['’]est-à-dire|add-on|add-ons|rendez-vous|garde-à-vous|chez-eux|chez-moi|chez-nous|chez-soi|chez-toi|chez-vous)$`),
		regexp.MustCompile(`(?i)^([cç]['’]|j['’]|n['’]|m['’]|t['’]|s['’]|l['’]|d['’]|qu['’]|jusqu['’]|lorsqu['’]|puisqu['’]|quoiqu['’])([^\-]*)(-ce|-elle|-t-elle|-elles|-t-elles|-en|-il|-t-il|-ils|-t-ils|-je|-la|-le|-les|-leur|-lui|-moi|-nous|-on|-t-on|-toi|-tu|-vous|-vs|-y)$`),
		regexp.MustCompile(`(?i)^([cç]['’]|j['’]|n['’]|m['’]|t['’]|s['’]|l['’]|d['’]|qu['’]|jusqu['’]|lorsqu['’]|puisqu['’]|quoiqu['’])([^'’\-].*)$`),
		regexp.MustCompile(`(?i)^([^\-\d]+)(-ce|-t-elle|-t-elles|-elle|-elles|-en|-il|-t-il|-ils|-t-ils|-je|-la|-le|-les|-leur|-lui|-moi|-nous|-on|-t-on|-toi|-tu|-vous|-vs|-y)(-ce|-elle|-t-elle|-elles|-t-elles|-en|-il|-t-il|-ils|-t-ils|-je|-la|-le|-les|-leur|-lui|-moi|-nous|-on|-t-on|-toi|-tu|-vous|-vs|-y)$`),
		regexp.MustCompile(`(?i)^([^\-]*)(-t|-m)(['’]en|['’]y)$`),
		regexp.MustCompile(`(?i)^(.*)(-t-elle|-t-elles|-t-il|-t-ils|-t-on)$`),
		regexp.MustCompile(`(?i)^(.*)(-ce|-elle|-t-elle|-elles|-t-elles|-en|-il|-t-il|-ils|-t-ils|-je|-la|-le|-les|-leur|-lui|-moi|-nous|-on|-t-on|-toi|-tu|-vous|-vs|-y)$`),
	}
)

func (w *FrenchWordTokenizer) Tokenize(text string) []string {
	auxText := typewriterApos.ReplaceAllString(text, "${1}xxFR_APOS_TYPEWxx${2}")
	auxText = typographicApos.ReplaceAllString(auxText, "${1}xxFR_APOS_TYPOGxx${2}")
	auxText = nearbyHyphens.ReplaceAllString(auxText, "${1}xxFR_HYPHENxx${2}xxFR_HYPHENxx${3}")
	auxText = hyphens.ReplaceAllString(auxText, "${1}xxFR_HYPHENxx${2}")
	auxText = decimalPoint.ReplaceAllString(auxText, "${1}xxFR_DECIMALPOINTxx${2}")
	auxText = decimalComma.ReplaceAllString(auxText, "${1}xxFR_DECIMALCOMMAxx${2}")
	auxText = spaceDigits2.ReplaceAllString(auxText, "${1}xxFR_SPACExx${2}xxFR_SPACExx${3}")
	auxText = spaceDigits0.ReplaceAllString(auxText, "${1}xxFR_SPACE0xx")
	auxText = spaceDigits.ReplaceAllString(auxText, "${1}xxFR_SPACExx${2}")
	auxText = strings.ReplaceAll(auxText, "xxFR_SPACE0xx", " ")

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
		s = strings.ReplaceAll(s, "xxFR_APOS_TYPEWxx", "'")
		s = strings.ReplaceAll(s, "xxFR_APOS_TYPOGxx", "’")
		s = strings.ReplaceAll(s, "xxFR_HYPHENxx", "-")
		s = strings.ReplaceAll(s, "xxFR_DECIMALPOINTxx", ".")
		s = strings.ReplaceAll(s, "xxFR_DECIMALCOMMAxx", ",")
		s = strings.ReplaceAll(s, "xxFR_SPACExx", " ")

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
		for _, p := range frPatterns {
			if m := p.FindStringSubmatch(s); m != nil {
				matchFound = true
				groups = m[1:]
				break
			}
		}
		if matchFound {
			for _, g := range groups {
				if g == "" {
					continue
				}
				l = append(l, wordsToAddFR(g)...)
			}
		} else {
			l = append(l, wordsToAddFR(s)...)
		}
		for hyphensAtEnd > 0 {
			l = append(l, "-")
			hyphensAtEnd--
		}
	}
	return tokenizers.JoinEMailsAndUrls(l)
}

func wordsToAddFR(s string) []string {
	var l []string
	if s == "" {
		return l
	}
	if !strings.Contains(s, "-") {
		l = append(l, s)
		return l
	}
	// Clitic suffixes from pattern groups (e.g. -tu, -t-elle) stay whole.
	if strings.HasPrefix(s, "-") {
		l = append(l, s)
		return l
	}
	// Soft hyphen compounds stay whole (check before stripping U+00AD).
	if strings.Contains(s, "\u00AD") || doNotSplit[strings.ToLower(s)] || isTaggedFR(s) {
		l = append(l, s)
	} else {
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
	}
	return l
}

func isTaggedFR(s string) bool {
	// Without FrenchTagger POS data, treat dictionary lookup as a miss so
	// wordsToAdd splits hyphens (Java splits untagged forms). Soft-hyphen
	// compounds and doNotSplit entries are handled by wordsToAdd callers.
	_ = s
	return false
}
