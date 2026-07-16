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

var (
	protocols = map[string]bool{
		"http": true, "https": true, "ws": true, "wss": true, "ftp": true, "ftps": true,
		"sftp": true, "file": true, "mailto": true, "tel": true, "sms": true, "git": true,
		"ssh": true, "data": true, "magnet": true, "smb": true, "slack": true, "spotify": true,
	}
	// Java: [a-zA-ZÄÖÜäöü0-9/%$-_.+!*'(),?#~]+ where $-_ is a char range ($.._).
	urlChars = regexp.MustCompile(`^[a-zA-ZÄÖÜäöü0-9/%$` + `-` + `_.+!*'(),?#~]+$`)
	domainChars   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]+$`)
	noProtocolURL = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-]+\.)?([a-zA-Z0-9][a-zA-Z0-9-]+)\.([a-zA-Z0-9][a-zA-Z0-9-]+)/.*$`)
	// simplified email pattern matching Java E_MAIL for tests
	eMailPattern = regexp.MustCompile(`(?i)@?\b[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))\b`)
)

// WordTokenizer ports org.languagetool.tokenizers.WordTokenizer.
type WordTokenizer struct{}

func NewWordTokenizer() *WordTokenizer { return &WordTokenizer{} }

func (w *WordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = nil
		}
	}
	for _, r := range text {
		if isTokenizing(r) {
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
	if !strings.Contains(text, "@") || !eMailPattern.MatchString(text) {
		return list
	}
	// Java walks tokens by UTF-16 length; emails are BMP so byte len == utf16 for ASCII emails.
	// Build cumulative UTF-16 positions.
	starts := make([]int, len(list))
	pos := 0
	for i, t := range list {
		starts[i] = pos
		pos += UTF16Len(t)
	}
	// Find email spans in flat text (UTF-8 byte → UTF-16 for BMP content)
	byteLocs := eMailPattern.FindAllStringIndex(text, -1)
	if len(byteLocs) == 0 {
		return list
	}
	type span struct{ start, end int }
	var spans []span
	for _, loc := range byteLocs {
		spans = append(spans, span{UTF16Len(text[:loc[0]]), UTF16Len(text[:loc[1]])})
	}
	var out []string
	idx := 0
	currentPosition := 0
	for _, sp := range spans {
		for currentPosition < sp.end && idx < len(list) {
			if currentPosition < sp.start {
				out = append(out, list[idx])
			} else if currentPosition == sp.start {
				// reconstruct email from covered tokens
				var em strings.Builder
				startIdx := idx
				for startIdx < len(list) && starts[startIdx] < sp.end {
					em.WriteString(list[startIdx])
					startIdx++
				}
				out = append(out, em.String())
			}
			currentPosition += UTF16Len(list[idx])
			idx++
		}
	}
	if idx < len(list) {
		out = append(out, list[idx:]...)
	}
	return out
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
	for p := range protocols {
		if strings.HasPrefix(token, p+"://") || strings.HasPrefix(token, "www.") {
			return true
		}
	}
	return noProtocolURL.MatchString(token)
}

// IsEMail ports WordTokenizer.isEMail.
func IsEMail(token string) bool {
	return eMailPattern.MatchString(token)
}

func isTokenizing(r rune) bool {
	return strings.ContainsRune(tokenizing, r)
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
