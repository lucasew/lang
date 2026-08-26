package tokenizers

import (
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// TOKENIZING_CHARACTERS from WordTokenizer.java
const tokenizing = "\u0020\u00A0\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007" +
	"\u2008\u2009\u200A\u200B\u200c\u200d\u200e\u200f" +
	"\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	"¦‖∣|,.;()[]{}=*#∗+×·÷<>!?:~/\\\"'«»„”“‘’`´‛′›‹…¿¡‼⁇⁈⁉™®\u203d\u00B6\uFFEB\u2E2E" +
	"\u2012\u2013\u2014\u2015" +
	"\u2500\u3161\u2713" +
	"\u25CF\u25CB\u25C6\u27A2\u25A0\u25A1\u2605\u274F\u2794\u21B5\u2756\u25AA\u2751\u2022" +
	"\u2B9A\u2265\u2192\u21FE\u21C9\u21D2\u21E8\u21DB" +
	"\u00b9\u00b2\u00b3\u2070\u2071\u2074\u2075\u2076\u2077\u2078\u2079" +
	"\t\n\r\u000B"

// RemovedEmoji ports WordTokenizer.REMOVED_EMOJI.
const RemovedEmoji = "MyReMoVeDeMoJi"

// protocolList ports WordTokenizer.PROTOCOLS (including duplicate "magnet" as in Java).
var protocolList = []string{
	"http", "https", "ws", "wss", "ftp", "ftps", "sftp", "file", "mailto", "tel", "sms",
	"git", "ssh", "data", "magnet", "smb", "slack", "spotify", "magnet",
}

var (
	protocols = func() map[string]bool {
		m := make(map[string]bool, len(protocolList))
		for _, p := range protocolList {
			m[p] = true
		}
		return m
	}()
	// Java: [a-zA-ZÄÖÜäöü0-9/%$-_.+!*'(),?#~]+ where $-_ is a char range ($.._).
	urlChars = regexp.MustCompile(`^[a-zA-ZÄÖÜäöü0-9/%$` + `-` + `_.+!*'(),?#~]+$`)
	domainChars   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]+$`)
	noProtocolURL = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-]+\.)?([a-zA-Z0-9][a-zA-Z0-9-]+)\.([a-zA-Z0-9][a-zA-Z0-9-]+)/.*$`)
	// Java E_MAIL body without (?<!:) lookbehind (RE2 has no lookbehind; emulated in helpers).
	// No (?i) — Java pattern is case-sensitive.
	eMailBody = `@?\b[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))\b`
	eMailPattern = regexp.MustCompile(eMailBody)
	// Currency (same pattern Java uses for matches() and find()).
	currencySymbolsClass = `[A-Z]*[฿₿₵¢₡$₫֏€ƒ₲₴₭₾₺₼₦₱£៛₽₹₪৳₸₮₩¥¤]`
	currencyValue        = `\d+(?:[.,]\d+)*`
	currencyExpression   = regexp.MustCompile(
		`(?:(` + currencySymbolsClass + `)(` + currencyValue + `)|(` + currencyValue + `)(` + currencySymbolsClass + `))`,
	)
	currencyExpressionFull = regexp.MustCompile(
		`^(?:(` + currencySymbolsClass + `)(` + currencyValue + `)|(` + currencyValue + `)(` + currencySymbolsClass + `))$`,
	)
)

// WordTokenizer ports org.languagetool.tokenizers.WordTokenizer.
type WordTokenizer struct{}

func NewWordTokenizer() *WordTokenizer { return &WordTokenizer{} }

// GetProtocols ports WordTokenizer.getProtocols.
func GetProtocols() []string {
	out := make([]string, len(protocolList))
	copy(out, protocolList)
	return out
}

// GetTokenizingCharacters ports WordTokenizer.getTokenizingCharacters.
func (w *WordTokenizer) GetTokenizingCharacters() string {
	return tokenizing
}

// TokenizingCharacters is the shared character set for subclasses.
func TokenizingCharacters() string { return tokenizing }

func (w *WordTokenizer) Tokenize(text string) []string {
	// Java: StringTokenizer(text, getTokenizingCharacters(), true)
	delims := w.GetTokenizingCharacters()
	out := make([]string, 0)
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = nil
		}
	}
	for _, r := range text {
		if strings.ContainsRune(delims, r) {
			flush()
			out = append(out, string(r))
		} else {
			cur = append(cur, r)
		}
	}
	flush()
	return JoinEMailsAndUrls(out)
}

func JoinEMailsAndUrls(list []string) []string {
	return joinUrls(joinEMails(list))
}

func joinEMails(list []string) []string {
	var sb strings.Builder
	for _, item := range list {
		sb.WriteString(item)
	}
	text := sb.String()
	// Java: text.contains("@") && E_MAIL.matcher(text).find()
	if !strings.Contains(text, "@") || !emailFind(text) {
		return list
	}
	// Java walks tokens by UTF-16 length; matcher.start/end are UTF-16 indices.
	spans := emailFindSpans(text)
	if len(spans) == 0 {
		return list
	}
	var out []string
	idx := 0
	currentPosition := 0
	for _, sp := range spans {
		for currentPosition < sp.end && idx < len(list) {
			if currentPosition < sp.start {
				out = append(out, list[idx])
			} else if currentPosition == sp.start {
				// Java: l.add(matcher.group())
				out = append(out, sp.group)
			}
			currentPosition += UTF16Len(list[idx])
			idx++
		}
	}
	// Java: if (currentPosition < text.length()) { l.addAll(list.subList(idx, list.size())); }
	if currentPosition < UTF16Len(text) {
		out = append(out, list[idx:]...)
	}
	return out
}

type emailSpan struct {
	start, end int // UTF-16
	group      string
}

// emailFind reports whether E_MAIL with (?<!:) would find a match (Java find()).
func emailFind(text string) bool {
	return len(emailFindSpans(text)) > 0
}

// emailFindSpans ports E_MAIL.matcher(text).find() with (?<!:) lookbehind emulation.
func emailFindSpans(text string) []emailSpan {
	byteLocs := eMailPattern.FindAllStringIndex(text, -1)
	if len(byteLocs) == 0 {
		return nil
	}
	var spans []emailSpan
	for _, loc := range byteLocs {
		// (?<!:): reject when previous UTF-8/byte unit is ':'
		if loc[0] > 0 && text[loc[0]-1] == ':' {
			continue
		}
		spans = append(spans, emailSpan{
			start: UTF16Len(text[:loc[0]]),
			end:   UTF16Len(text[:loc[1]]),
			group: text[loc[0]:loc[1]],
		})
	}
	return spans
}

func joinUrls(l []string) []string {
	var newList []string
	inURL := false
	var url strings.Builder
	var urlQuote string
	for i := 0; i < len(l); i++ {
		if urlStartsAt(i, l) && !inURL {
			inURL = true
			if i-1 >= 0 {
				urlQuote = l[i-1]
			}
			url.WriteString(l[i])
		} else if inURL && urlEndsAt(i, l, urlQuote) {
			inURL = false
			urlQuote = ""
			newList = append(newList, url.String())
			url.Reset()
			newList = append(newList, l[i])
		} else if inURL {
			url.WriteString(l[i])
		} else {
			newList = append(newList, l[i])
		}
	}
	if url.Len() > 0 {
		newList = append(newList, url.String())
	}
	return newList
}

func urlStartsAt(i int, l []string) bool {
	token := l[i]
	if isProtocol(token) && len(l) > i+3 {
		if l[i+1] == ":" && l[i+2] == "/" && l[i+3] == "/" {
			return true
		}
	}
	if len(l) > i+1 {
		if l[i] == "www" && l[i+1] == "." {
			return true
		}
	}
	if len(l) > i+3 &&
		l[i+1] == "." &&
		l[i+3] == "/" &&
		domainChars.MatchString(token) &&
		domainChars.MatchString(l[i+2]) {
		return true
	}
	if len(l) > i+5 &&
		l[i+1] == "." &&
		l[i+3] == "." &&
		l[i+5] == "/" &&
		domainChars.MatchString(token) &&
		domainChars.MatchString(l[i+2]) &&
		domainChars.MatchString(l[i+4]) {
		return true
	}
	return false
}

func isProtocol(token string) bool { return protocols[token] }

func urlEndsAt(i int, l []string, urlQuote string) bool {
	token := l[i]
	if tools.IsWhitespace(token) || token == ")" || token == "]" {
		return true
	}
	if len(l) > i+1 {
		nextToken := l[i+1]
		if (tools.IsWhitespace(nextToken) || isAny(nextToken, "\"", "»", "«", "‘", "’", "“", "”", "'", ".")) &&
			(isAny(token, ".", ",", ";", ":", "!", "?") || token == urlQuote) {
			return true
		}
		if !urlChars.MatchString(token) {
			return true
		}
	} else {
		if !urlChars.MatchString(token) || token == "." || token == urlQuote {
			return true
		}
	}
	return false
}

func isAny(s string, opts ...string) bool {
	for _, o := range opts {
		if s == o {
			return true
		}
	}
	return false
}

// IsURL ports WordTokenizer.isUrl.
func IsURL(token string) bool {
	for _, p := range protocolList {
		if strings.HasPrefix(token, p+"://") || strings.HasPrefix(token, "www.") {
			return true
		}
	}
	return noProtocolURL.MatchString(token)
}

// IsEMail ports WordTokenizer.isEMail (Java Pattern.matcher(token).matches()).
func IsEMail(token string) bool {
	// Full-string match of E_MAIL; (?<!:) at index 0 always succeeds.
	loc := eMailPattern.FindStringIndex(token)
	return loc != nil && loc[0] == 0 && loc[1] == len(token)
}

// UTF16Len returns Java String.length() equivalent.
func UTF16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func utf16Len(s string) int { return UTF16Len(s) }

// BuildPositions returns start offset (UTF-16) for each token when concatenated.
func BuildPositions(tokens []string) []int {
	pos := make([]int, len(tokens))
	p := 0
	for i, t := range tokens {
		pos[i] = p
		p += UTF16Len(t)
	}
	return pos
}

// IsCurrencyExpression ports WordTokenizer.isCurrencyExpression.
func IsCurrencyExpression(token string) bool {
	return currencyExpressionFull.MatchString(token)
}

// SplitCurrencyExpression ports WordTokenizer.splitCurrencyExpression.
func SplitCurrencyExpression(token string) []string {
	var newList []string
	for _, m := range currencyExpression.FindAllStringSubmatch(token, -1) {
		if m[1] != "" && m[2] != "" {
			newList = append(newList, m[1], m[2])
		} else if m[3] != "" && m[4] != "" {
			newList = append(newList, m[3], m[4])
		}
	}
	if len(newList) == 0 {
		return []string{token}
	}
	return newList
}

// IsCurrencyExpression ports the instance method on WordTokenizer.
func (w *WordTokenizer) IsCurrencyExpression(token string) bool {
	return IsCurrencyExpression(token)
}

// SplitCurrencyExpression ports the instance method on WordTokenizer.
func (w *WordTokenizer) SplitCurrencyExpression(token string) []string {
	return SplitCurrencyExpression(token)
}

// ReplaceEmojis ports WordTokenizer.replaceEmojis.
// Output: cleaned string at [0], then removed emojis in order.
func (w *WordTokenizer) ReplaceEmojis(s string) []string {
	removedEmojis := make([]string, 0)
	// Java: s.length() > 1 && s.codePointCount(0, s.length()) != s.length()
	if UTF16Len(s) > 1 {
		cps := 0
		for range s {
			cps++
		}
		if cps != UTF16Len(s) {
			// Matcher is bound to the original string; replacements apply to s.
			original := s
			for _, found := range tools.CharsNotForSpelling.FindAllString(original, -1) {
				// emojis (😂) have a string length (UTF-16) larger than 1
				if UTF16Len(found) > 1 {
					s = strings.ReplaceAll(s, found, ","+RemovedEmoji+",")
					removedEmojis = append(removedEmojis, found)
				}
			}
		}
	}
	out := make([]string, 0, 1+len(removedEmojis))
	out = append(out, s)
	out = append(out, removedEmojis...)
	return out
}

// RestoreEmojis ports WordTokenizer.restoreEmojis.
func (w *WordTokenizer) RestoreEmojis(tokens []string, removedEmojis []string) []string {
	if len(removedEmojis) < 2 {
		return tokens
	}
	results := make([]string, 0, len(tokens))
	i := 0
	emojiCount := 1
	for i < len(tokens) {
		// Java: tokens.get(i).equals(",") && tokens.get(i+1).equals(REMOVED_EMOJI) && tokens.get(i+2).equals(",")
		if i+2 < len(tokens) && tokens[i] == "," &&
			tokens[i+1] == RemovedEmoji && tokens[i+2] == "," {
			results = append(results, removedEmojis[emojiCount])
			emojiCount++
			i += 3
		} else {
			results = append(results, tokens[i])
			i++
		}
	}
	return results
}
