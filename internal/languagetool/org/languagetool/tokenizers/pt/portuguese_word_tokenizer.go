package pt

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PortugueseWordTokenizer ports org.languagetool.tokenizers.pt.PortugueseWordTokenizer.
type PortugueseWordTokenizer struct{}

func NewPortugueseWordTokenizer() *PortugueseWordTokenizer {
	return &PortugueseWordTokenizer{}
}

const (
	decimalCommaSubst     = '\uE001'
	nonBreakingSpaceSubst = '\uE002'
	nonBreakingDotSubst   = '\uE003'
	nonBreakingColonSubst = '\uE004'
	hyphenSubstText       = "\u0001\u0001PT_HYPHEN\u0001\u0001"
)

var (
	decimalCommaPattern  = regexp.MustCompile(`(?i)([\d]),([\d])`)
	dottedNumbersPattern = regexp.MustCompile(`(?i)([\d])\.([\d])`)
	colonNumbersPattern  = regexp.MustCompile(`(?i)([\d]):([\d])`)
	datePattern          = regexp.MustCompile(`(?i)([\d]{2})\.([\d]{2})\.([\d]{4})|([\d]{4})\.([\d]{2})\.([\d]{2})|([\d]{4})-([\d]{2})-([\d]{2})`)
	dottedOrdinals       = regexp.MustCompile(`(?i)([\d])\.([aoªºᵃᵒ][sˢ]?)`)
	hyphenPattern        = regexp.MustCompile(`(?i)([\p{L}])-([\p{L}\d])`)
	nearbyHyphens        = regexp.MustCompile(`(?i)([\p{L}])-([\p{L}])-([\p{L}])`)

	wordChars          = "°\\^\\-\\p{L}\\d\\x{0300}-\\x{036F}\\x{00A8}\\x{2070}-\\x{209F}" + string(decimalCommaSubst) + string(nonBreakingDotSubst) + string(nonBreakingColonSubst) + string(nonBreakingSpaceSubst)
	wordCharsLeftEdge  = `−@€£\$¢¥¤`
	wordCharsRightEdge = `€£\$%‰‱ºªᵃᵒˢ`
	wordPattern        = regexp.MustCompile(`(?i)[` + wordCharsLeftEdge + `]?[` + wordChars + hyphenSubstText + `]+[` + wordCharsRightEdge + `]?|[^` + wordChars + `]`)

	// Java wordsToAdd camel-case hyphen exceptions only (not invent soft compounds).
	javaHyphenExceptions = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true, "anti-ivg": true, "anti-uv": true,
		"anti-vih": true, "al-qaïda": true,
	}
)

// IsTaggedPT optional PortugueseTagger.tag(...).isTagged() hook.
// Java PortugueseWordTokenizer uses PortugueseTagger for hyphen compounds.
// Without a tagger, miss (split hyphens) — do not invent a soft compound lexicon.
var IsTaggedPT func(s string) bool

func (w *PortugueseWordTokenizer) Tokenize(text string) []string {
	tokenisedText := text
	if strings.Contains(tokenisedText, ",") {
		tokenisedText = decimalCommaPattern.ReplaceAllString(tokenisedText, "$1"+string(decimalCommaSubst)+"$2")
	}
	dotIndex := strings.IndexByte(tokenisedText, '.')
	if dotIndex >= 0 && dotIndex < len(tokenisedText)-1 {
		tokenisedText = datePattern.ReplaceAllStringFunc(tokenisedText, func(m string) string {
			return strings.ReplaceAll(strings.ReplaceAll(m, ".", string(nonBreakingDotSubst)), "-", string(nonBreakingDotSubst))
		})
		tokenisedText = dottedNumbersPattern.ReplaceAllString(tokenisedText, "$1"+string(nonBreakingDotSubst)+"$2")
		tokenisedText = dottedOrdinals.ReplaceAllString(tokenisedText, "$1"+string(nonBreakingDotSubst)+"$2")
	}
	tokenisedText = protectPTSpacedNumbers(tokenisedText)
	if strings.Contains(tokenisedText, ":") {
		tokenisedText = colonNumbersPattern.ReplaceAllString(tokenisedText, "$1"+string(nonBreakingColonSubst)+"$2")
	}
	if strings.Contains(tokenisedText, "-") {
		tokenisedText = nearbyHyphens.ReplaceAllString(tokenisedText, "$1"+hyphenSubstText+"$2"+hyphenSubstText+"$3")
		tokenisedText = hyphenPattern.ReplaceAllString(tokenisedText, "$1"+hyphenSubstText+"$2")
	}

	var tokenList []string
	for _, loc := range wordPattern.FindAllStringIndex(tokenisedText, -1) {
		token := tokenisedText[loc[0]:loc[1]]
		if len(tokenList) > 0 {
			r, size := utf8.DecodeRuneInString(token)
			if size == len(token) && r >= 0xFE00 && r <= 0xFE0F {
				tokenList[len(tokenList)-1] = tokenList[len(tokenList)-1] + token
				continue
			}
		}
		token = strings.ReplaceAll(token, string(decimalCommaSubst), ",")
		token = strings.ReplaceAll(token, string(nonBreakingColonSubst), ":")
		token = strings.ReplaceAll(token, string(nonBreakingSpaceSubst), " ")
		token = strings.ReplaceAll(token, string(nonBreakingDotSubst), ".")
		token = strings.ReplaceAll(token, hyphenSubstText, "-")
		tokenList = append(tokenList, wordsToAddPT(token)...)
	}
	return tokenizers.JoinEMailsAndUrls(tokenList)
}

func protectPTSpacedNumbers(text string) string {
	// Java: (?<=^|[\s(])\d{1,3}( \d{3})+(?:[.,…]\d+)?(?=\D|$)
	re := regexp.MustCompile(`(?:^|[\s(])(\d{1,3}(?: \d{3})+(?:[` + string(decimalCommaSubst) + string(nonBreakingDotSubst) + `.,]\d+)?)`)
	return re.ReplaceAllStringFunc(text, func(m string) string {
		i := 0
		for i < len(m) && (m[i] < '0' || m[i] > '9') {
			i++
		}
		prefix := m[:i]
		num := m[i:]
		num = strings.ReplaceAll(num, " ", string(nonBreakingSpaceSubst))
		num = strings.ReplaceAll(num, "\u00A0", string(nonBreakingSpaceSubst))
		return prefix + num
	})
}

// wordsToAddPT ports PortugueseWordTokenizer.wordsToAdd.
func wordsToAddPT(s string) []string {
	var l []string
	if s == "" {
		return l
	}
	if tokenizers.IsCurrencyExpression(s) {
		return tokenizers.SplitCurrencyExpression(s)
	}
	if !strings.Contains(s, "-") {
		l = append(l, s)
		return l
	}
	normalized := strings.ReplaceAll(s, "’", "'")
	// Java: tagger.tag(...).isTagged() OR equalsIgnoreCase exception list.
	if isTaggedPT(normalized) || javaHyphenException(s) {
		l = append(l, s)
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

func javaHyphenException(s string) bool {
	return javaHyphenExceptions[strings.ToLower(s)]
}

func isTaggedPT(s string) bool {
	// Java: PortugueseTagger.tag(...).isTagged(). Without a tagger, miss
	// (split hyphens) — do not invent a soft compound lexicon.
	if IsTaggedPT != nil {
		return IsTaggedPT(s)
	}
	return false
}
