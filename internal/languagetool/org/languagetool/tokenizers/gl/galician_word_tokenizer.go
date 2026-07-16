package gl

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GalicianWordTokenizer ports org.languagetool.tokenizers.gl.GalicianWordTokenizer.
type GalicianWordTokenizer struct{}

func NewGalicianWordTokenizer() *GalicianWordTokenizer { return &GalicianWordTokenizer{} }

const splitChars = " \u002d\u00A0" +
	"\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008" +
	"\u2009\u2013\u2014\u2015\u200A\u200B\u200c\u200d\u200e" +
	"\u200f\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	"\u002A\u002B×∗·÷:=≠≂≃≄≅≆≇≈≉≤≥≪≫∧∨∩∪∈∉∊∋∌∍" +
	",.;<>()[]{}¿¡!?:\"«»`'’‘„“”…\\/\t\r\n"

const (
	decimalCommaSubst     = '\uE001'
	nonBreakingSpaceSubst = '\uE002'
	nonBreakingDotSubst   = '\uE003'
	nonBreakingColonSubst = '\uE004'
)

var (
	decimalCommaPattern = regexp.MustCompile(`(?i)([\d]),([\d])`)
	dottedNumbers       = regexp.MustCompile(`(?i)([\d])\.([\d])`)
	colonNumbers        = regexp.MustCompile(`(?i)([\d]):([\d])`)
	datePattern         = regexp.MustCompile(`(?i)([\d]{2})\.([\d]{2})\.([\d]{4})|([\d]{4})\.([\d]{2})\.([\d]{2})|([\d]{4})-([\d]{2})-([\d]{2})`)
	dottedOrdinals      = regexp.MustCompile(`(?i)([\d])\.([aoªº][sˢ]?)`)
	spacedNumberPat     = regexp.MustCompile(`(^|[\s(])(\d{1,3}(?: \d{3})+)([\s(]|$)`)
)

func (w *GalicianWordTokenizer) Tokenize(text string) []string {
	if strings.Contains(text, ",") {
		text = decimalCommaPattern.ReplaceAllString(text, "$1"+string(decimalCommaSubst)+"$2")
	}
	dotIndex := strings.IndexByte(text, '.')
	if dotIndex >= 0 && dotIndex < len(text)-1 {
		text = datePattern.ReplaceAllStringFunc(text, func(m string) string {
			return strings.ReplaceAll(strings.ReplaceAll(m, ".", string(nonBreakingDotSubst)), "-", string(nonBreakingDotSubst))
		})
		text = dottedNumbers.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+"$2")
		text = dottedOrdinals.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+"$2")
	}
	text = spacedNumberPat.ReplaceAllStringFunc(text, func(m string) string {
		idx := 0
		for idx < len(m) && (m[idx] < '0' || m[idx] > '9') {
			idx++
		}
		prefix := m[:idx]
		rest := m[idx:]
		end := len(rest)
		if end > 0 {
			last := rest[end-1]
			if last == ' ' || last == '(' {
				end--
			}
		}
		num := rest[:end]
		suffix := rest[end:]
		num = strings.ReplaceAll(num, " ", string(nonBreakingSpaceSubst))
		return prefix + num + suffix
	})
	if strings.Contains(text, ":") {
		text = colonNumbers.ReplaceAllString(text, "$1"+string(nonBreakingColonSubst)+"$2")
	}
	raw := splitKeepDelims(text, splitChars)
	var tokenList []string
	for _, token := range raw {
		token = strings.ReplaceAll(token, string(decimalCommaSubst), ",")
		token = strings.ReplaceAll(token, string(nonBreakingColonSubst), ":")
		token = strings.ReplaceAll(token, string(nonBreakingSpaceSubst), " ")
		token = strings.ReplaceAll(token, string(nonBreakingDotSubst), ".")
		tokenList = append(tokenList, token)
	}
	return tokenizers.JoinEMailsAndUrls(tokenList)
}

func splitKeepDelims(text, delims string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var cur strings.Builder
	for _, r := range text {
		if strings.ContainsRune(delims, r) {
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
			out = append(out, string(r))
		} else {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
