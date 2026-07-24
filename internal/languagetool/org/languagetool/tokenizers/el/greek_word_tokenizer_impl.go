package el

import (
	"strings"
	"unicode/utf8"
)

// GreekWordTokenizerImpl ports the JFlex-generated scanner
// org.languagetool.tokenizers.el.GreekWordTokenizerImpl
// (spec: GreekWordTokenizerImpl.jflex).
//
// Rule-equivalent hand transcription of:
//
//	Delim = ( spaces + ",.;()[]{}!:\"'" + "·" + quotes + … + «» + \ / + \t \n )
//	Word  = ("ό,τι" | (!Delim)* | Delim)
//
// JFlex longest-match + DFA finality:
//   - "ό,τι" is recognized only at a token start (not mid non-Delim run)
//   - each Delim code point is its own token
//   - maximal non-Delim runs are word tokens (comma is Delim, so ends a run)
//
// Full packed DFA tables are not required when this semantics is proven
// against the jflex Word production / generated scanner behavior.
type GreekWordTokenizerImpl struct {
	text     string
	pos      int
	lastText string
}

// YYEOF ports GreekWordTokenizerImpl.YYEOF.
const YYEOF = -1

// greekDelimChars is every single-code-point Delim from GreekWordTokenizerImpl.jflex,
// character-for-character (73 code points). Order mirrors the jflex Delim production.
//
// Note: no '?', no '\r', no ASCII '-' (contrast core WordTokenizer / other langs).
const greekDelimChars = "\u0020\u00A0\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007" +
	"\u2008\u2009\u200A\u200B\u200c\u200d\u200e\u200f" +
	"\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	",.;()[]{}!:\"'" +
	"·" + // Greek Ano Teleia U+0387
	"’‘„“”…«»\\/\t\n"

// greekSpecialOti is the multi-char Word alternative "ό,τι" (contains Delim comma).
const greekSpecialOti = "ό,τι"

var greekDelim = buildGreekDelimSet()

func buildGreekDelimSet() map[rune]struct{} {
	m := make(map[rune]struct{}, len(greekDelimChars))
	for _, r := range greekDelimChars {
		m[r] = struct{}{}
	}
	return m
}

func isGreekDelim(r rune) bool {
	_, ok := greekDelim[r]
	return ok
}

// NewGreekWordTokenizerImpl ports new GreekWordTokenizerImpl(new StringReader("")).
func NewGreekWordTokenizerImpl() *GreekWordTokenizerImpl {
	return &GreekWordTokenizerImpl{}
}

// Yyreset ports yyreset(Reader) for a full input string (one-shot buffer).
func (s *GreekWordTokenizerImpl) Yyreset(text string) {
	s.text = text
	s.pos = 0
	s.lastText = ""
}

// GetNextToken ports getNextToken(): returns 0 on a Word match, YYEOF at end.
// Matched text is available via GetText().
func (s *GreekWordTokenizerImpl) GetNextToken() int {
	if s.pos >= len(s.text) {
		s.lastText = ""
		return YYEOF
	}
	// Special multi-char token only at token start (JFlex alternative before
	// non-Delim run; DFA reaches accepting final only from lexical state 0).
	if strings.HasPrefix(s.text[s.pos:], greekSpecialOti) {
		s.lastText = greekSpecialOti
		s.pos += len(greekSpecialOti)
		return 0
	}
	r, size := utf8.DecodeRuneInString(s.text[s.pos:])
	if r == utf8.RuneError && size == 1 {
		// Invalid UTF-8 byte: treat as a one-byte non-delim word fragment.
		s.lastText = s.text[s.pos : s.pos+1]
		s.pos++
		return 0
	}
	if isGreekDelim(r) {
		s.lastText = s.text[s.pos : s.pos+size]
		s.pos += size
		return 0
	}
	// Maximal non-Delim run. Do not re-check "ό,τι" mid-run: inside a word
	// the DFA treats ό/τ/ι as ordinary non-Delim; comma (Delim) ends the run.
	start := s.pos
	s.pos += size
	for s.pos < len(s.text) {
		r2, sz2 := utf8.DecodeRuneInString(s.text[s.pos:])
		if r2 == utf8.RuneError && sz2 == 1 {
			s.pos++
			continue
		}
		if isGreekDelim(r2) {
			break
		}
		s.pos += sz2
	}
	s.lastText = s.text[start:s.pos]
	return 0
}

// GetText ports getText() — text of the last successful GetNextToken match.
func (s *GreekWordTokenizerImpl) GetText() string {
	return s.lastText
}

// YylexTokenize runs the scanner to completion without joinEMailsAndUrls
// (raw JFlex token list). Convenience for tests / one-shot use.
func (s *GreekWordTokenizerImpl) YylexTokenize(text string) []string {
	s.Yyreset(text)
	var out []string
	for s.GetNextToken() != YYEOF {
		out = append(out, s.GetText())
	}
	return out
}
