package pipeline

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// LT WordTokenizer TOKENIZING_CHARACTERS (+ English underscore).
// Source: org.languagetool.tokenizers.WordTokenizer
const tokenizingChars = "\u0020\u00A0\u115f\u1160\u1680" +
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
	"\t\n\r\u000B" +
	"_" // EnglishWordTokenizer adds underscore

var tokenizingSet map[rune]bool

func init() {
	tokenizingSet = make(map[rune]bool, utf8.RuneCountInString(tokenizingChars))
	for _, r := range tokenizingChars {
		tokenizingSet[r] = true
	}
}

// WordTokenize ports WordTokenizer.tokenize (character class split) without URL/email join yet.
// Offsets are rune indices (Java char for BMP).
func WordTokenize(text string) []Token {
	if text == "" {
		return nil
	}
	var tokens []Token
	runeIdx := 0
	i := 0
	for i < len(text) {
		r, size := utf8.DecodeRuneInString(text[i:])
		if tokenizingSet[r] {
			tok := text[i : i+size]
			tokens = append(tokens, Token{
				Text:       tok,
				Start:      runeIdx,
				End:        runeIdx + 1,
				Whitespace: isWSToken(r, tok),
				Linebreak:  r == '\n' || r == '\r',
			})
			i += size
			runeIdx++
			continue
		}
		// word run
		startRune := runeIdx
		startByte := i
		i += size
		runeIdx++
		for i < len(text) {
			r2, sz := utf8.DecodeRuneInString(text[i:])
			if tokenizingSet[r2] {
				break
			}
			i += sz
			runeIdx++
		}
		tokens = append(tokens, Token{
			Text:  text[startByte:i],
			Start: startRune,
			End:   runeIdx,
		})
	}
	return tokens
}

func isWSToken(r rune, s string) bool {
	if r == '\u00A0' || r == '\u202F' || r == '\uFEFF' || r == '\u200B' {
		return true
	}
	return unicode.IsSpace(r)
}

// TokensWithoutWhitespace drops whitespace tokens and prepends SENT_START (LT analysis).
func TokensWithoutWhitespace(tokens []Token) []Token {
	out := make([]Token, 0, len(tokens)+1)
	out = append(out, Token{Text: "SENT_START", Start: 0, End: 0})
	for _, t := range tokens {
		if t.Whitespace {
			continue
		}
		// pure space/tab/newline already whitespace; also skip empty
		if strings.TrimSpace(t.Text) == "" && t.Text != "" {
			// non-breaking etc. classified whitespace
			continue
		}
		if t.Text == " " || t.Text == "\t" || t.Text == "\n" || t.Text == "\r" {
			continue
		}
		// Tokenizing chars that are punctuation are kept (not whitespace).
		if t.Whitespace {
			continue
		}
		out = append(out, t)
	}
	return out
}

// SpaceBefore reports whether a non-whitespace token had whitespace immediately before it in the original stream.
func SpaceBefore(all []Token, nonWSIndex int, nonWS []Token) bool {
	if nonWSIndex <= 0 {
		return false
	}
	// Find this token in all by start offset; check previous all-token.
	target := nonWS[nonWSIndex]
	for i, t := range all {
		if t.Start == target.Start && t.End == target.End && t.Text == target.Text {
			if i == 0 {
				return false
			}
			return all[i-1].Whitespace || unicode.IsSpace([]rune(all[i-1].Text)[0])
		}
	}
	return false
}
