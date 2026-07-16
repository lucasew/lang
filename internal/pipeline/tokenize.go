package pipeline

import (
	"unicode"
	"unicode/utf8"
)

// Token is one analyzed token (subset of LT AnalyzedTokenReadings).
// Start/End are Unicode code-point offsets (Java String indices for BMP text),
// matching LanguageTool RuleMatch fromPos/toPos for typical rule data.
type Token struct {
	Text       string
	Start      int // rune offset in original text
	End        int // exclusive rune offset
	Whitespace bool
	Linebreak  bool
	// Readings holds POS/lemma analyses (empty = untagged / UNKNOWN).
	Readings []Reading
}

// Reading is one morphological analysis (lemma + POS tag).
type Reading struct {
	Lemma string
	POS   string
}

// Lemmas returns all lemmas (surface form if untagged).
func (t Token) Lemmas() []string {
	if len(t.Readings) == 0 {
		return []string{t.Text}
	}
	out := make([]string, 0, len(t.Readings))
	for _, r := range t.Readings {
		if r.Lemma != "" {
			out = append(out, r.Lemma)
		}
	}
	if len(out) == 0 {
		return []string{t.Text}
	}
	return out
}

// TokenizeWhitespaceAware splits text into non-whitespace runs and single-character
// whitespace tokens, approximating LanguageTool's whitespace tokenization enough
// for MultipleWhitespaceRule and similar text-level rules.
//
// Full LT WordTokenizer per language is a later stage; this is not a stub for the
// pipeline architecture — it is the first real tokenizer implementation, scoped
// to rules that only inspect whitespace flags.
func TokenizeWhitespaceAware(text string) []Token {
	if text == "" {
		return nil
	}
	var tokens []Token
	runeIdx := 0
	byteIdx := 0
	for byteIdx < len(text) {
		r, size := utf8.DecodeRuneInString(text[byteIdx:])
		if r == utf8.RuneError && size == 1 {
			tokens = append(tokens, Token{Text: text[byteIdx : byteIdx+1], Start: runeIdx, End: runeIdx + 1})
			byteIdx++
			runeIdx++
			continue
		}
		if isLTWhitespaceRune(r) {
			tokens = append(tokens, Token{
				Text:       text[byteIdx : byteIdx+size],
				Start:      runeIdx,
				End:        runeIdx + 1,
				Whitespace: true,
				Linebreak:  r == '\n' || r == '\r',
			})
			byteIdx += size
			runeIdx++
			continue
		}
		// Non-whitespace run.
		startRune := runeIdx
		startByte := byteIdx
		byteIdx += size
		runeIdx++
		for byteIdx < len(text) {
			r2, sz := utf8.DecodeRuneInString(text[byteIdx:])
			if isLTWhitespaceRune(r2) {
				break
			}
			byteIdx += sz
			runeIdx++
		}
		tokens = append(tokens, Token{
			Text:  text[startByte:byteIdx],
			Start: startRune,
			End:   runeIdx,
		})
	}
	return tokens
}

func isLTWhitespaceRune(r rune) bool {
	switch r {
	case '\u00A0', '\u202F', '\uFEFF', '\u200B', '\u2060':
		return true
	}
	return unicode.IsSpace(r)
}

// IsNonBreakingWhitespace reports LT StringTools.isNonBreakingWhitespace.
func IsNonBreakingWhitespace(s string) bool {
	return s == "\u00A0"
}

// IsFirstWhite ports MultipleWhitespaceRule.isFirstWhite.
func IsFirstWhite(t Token) bool {
	if !(t.Whitespace || IsNonBreakingWhitespace(t.Text)) {
		return false
	}
	if t.Linebreak {
		return false
	}
	if containsInvisible(t.Text) {
		return false
	}
	return true
}

// IsRemovableWhite ports MultipleWhitespaceRule.isRemovableWhite.
func IsRemovableWhite(t Token) bool {
	if !(t.Whitespace || IsNonBreakingWhitespace(t.Text)) {
		return false
	}
	if t.Linebreak || t.Text == "\t" {
		return false
	}
	if containsInvisible(t.Text) {
		return false
	}
	return true
}

func containsInvisible(s string) bool {
	for _, sub := range []string{"\u200B", "\uFEFF", "\u2060"} {
		if containsStr(s, sub) {
			return true
		}
	}
	return false
}

func containsStr(s, sub string) bool {
	return len(sub) == 0 || indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
