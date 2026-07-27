package pl

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PolishWordTokenizer ports org.languagetool.tokenizers.pl.PolishWordTokenizer.
// Without SetTagger, hyphen compounds (including number ranges) stay whole —
// same as Java when tagger is null. Call SetTagger for hybrid hyphen splitting
// and digit-range split (1-23).
type PolishWordTokenizer struct {
	plTokenizing string
	// tagger optional; nil matches Java before setTagger()
	tagger PolishHyphenTagger
}

// PolishHyphenTagger is the subset of org.languagetool.tagging.Tagger used for
// hybrid hyphen tokenization (Java field type is Tagger).
// Implementations typically wrap PolishTagger.Tag.
type PolishHyphenTagger interface {
	// Tag returns one readings object per input token (same contract as Tagger.tag).
	Tag(tokens []string) []PolishHyphenReadings
}

// PolishHyphenReadings is the subset of AnalyzedTokenReadings consulted by
// PolishWordTokenizer. *languagetool.AnalyzedTokenReadings implements this.
type PolishHyphenReadings interface {
	IsTagged() bool
	HasPosTag(posTag string) bool
	HasPartialPosTag(posTag string) bool
}

// ATRTagFunc adapts a batch-tag function to PolishHyphenTagger.
// Use to wrap PolishTagger.Tag (or any ATR producer) without invent POS lists.
type ATRTagFunc func(tokens []string) []PolishHyphenReadings

// Tag implements PolishHyphenTagger.
func (f ATRTagFunc) Tag(tokens []string) []PolishHyphenReadings {
	if f == nil {
		return nil
	}
	return f(tokens)
}

// WrapATRTagger converts a batch-tag function for SetTagger.
func WrapATRTagger(tag func(tokens []string) []PolishHyphenReadings) PolishHyphenTagger {
	return ATRTagFunc(tag)
}

func NewPolishWordTokenizer() *PolishWordTokenizer {
	return &PolishWordTokenizer{
		// Java: super.getTokenizingCharacters() + "–‚" (n-dash + low-9 quote)
		plTokenizing: tokenizers.TokenizingCharacters() + "–‚",
	}
}

// SetTagger ports PolishWordTokenizer.setTagger — enables hybrid hyphen splitting
// and digit-range splits using POS from the real Polish tagger.
func (w *PolishWordTokenizer) SetTagger(t PolishHyphenTagger) {
	w.tagger = t
}

// Polish prefixes that should never be used to split parts of words (Java static set).
var polishPrefixes = map[string]bool{
	"arcy": true, "neo": true, "pre": true, "anty": true, "eks": true, "bez": true,
	"beze": true, "ekstra": true, "hiper": true, "infra": true, "kontr": true,
	"maksi": true, "midi": true, "między": true, "mini": true, "nad": true,
	"nade": true, "około": true, "ponad": true, "post": true, "pro": true,
	"przeciw": true, "pseudo": true, "super": true, "śród": true, "ultra": true,
	"wice": true, "wokół": true, "wokoło": true,
}

// Tokenize ports PolishWordTokenizer.tokenize.
func (w *PolishWordTokenizer) Tokenize(text string) []string {
	// Java: StringTokenizer(text, plTokenizing, true)
	raw := splitKeepDelims(text, w.plTokenizing)
	var l []string
	for _, token := range raw {
		// Java: token.length() > 1 (UTF-16 code units)
		if tokenizers.UTF16Len(token) > 1 {
			if strings.HasSuffix(token, "-") {
				// hyphen at end of word → detach
				l = append(l, token[:len(token)-1], "-")
			} else if strings.HasPrefix(token, "-") {
				// Java: token.charAt(0) == '-'
				l = append(l, "-")
				l = append(l, w.Tokenize(token[1:])...)
			} else if strings.Contains(token, "-") {
				// Java: token.split("-")
				tokenParts := strings.Split(token, "-")
				// Java: prefixes.contains(tokenParts[0]) || tagger == null → keep whole
				if polishPrefixes[tokenParts[0]] || w.tagger == nil {
					l = append(l, token)
				} else if lastPartStartsWithDigit(tokenParts) {
					// split numbers at dash or minus sign, 1-10
					// Java: Character.isDigit(tokenParts[tokenParts.length - 1].charAt(0))
					for i, p := range tokenParts {
						l = append(l, p)
						if i != len(tokenParts)-1 {
							l = append(l, "-")
						}
					}
				} else {
					l = append(l, w.splitCompoundWithTagger(token, tokenParts)...)
				}
			} else {
				l = append(l, token)
			}
		} else {
			l = append(l, token)
		}
	}
	return tokenizers.JoinEMailsAndUrls(l)
}

// lastPartStartsWithDigit ports Java's digit-range guard:
// Character.isDigit(tokenParts[tokenParts.length - 1].charAt(0)).
// Empty last part cannot reach here when endsWith("-") is handled first;
// still guard to avoid panics on pathological splits.
func lastPartStartsWithDigit(tokenParts []string) bool {
	if len(tokenParts) == 0 {
		return false
	}
	last := tokenParts[len(tokenParts)-1]
	if last == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(last)
	return unicode.IsDigit(r)
}

func (w *PolishWordTokenizer) splitCompoundWithTagger(token string, tokenParts []string) []string {
	if w.tagger == nil {
		return []string{token}
	}
	// Java: testedTokens = tokenParts + whole token; tagger.tag(testedTokens)
	tested := append(append([]string{}, tokenParts...), token)
	taggedToks := w.tagger.Tag(tested)
	// Java: taggedToks.size() == tokenParts.length + 1
	//       && !taggedToks.get(tokenParts.length).isTagged()
	if len(taggedToks) == len(tokenParts)+1 &&
		taggedToks[len(tokenParts)] != nil &&
		!taggedToks[len(tokenParts)].IsTagged() {
		isCompound := false
		switch len(tokenParts) {
		case 2:
			// "niemiecko-indonezyjski" / "kobieta-wojownik" / "osiemnaście-dwadzieścia"
			if taggedToks[0] != nil && taggedToks[1] != nil &&
				((taggedToks[0].HasPosTag("adja") && taggedToks[1].HasPartialPosTag("adj:")) ||
					(taggedToks[0].HasPartialPosTag("subst:") && taggedToks[1].HasPartialPosTag("subst:")) ||
					(taggedToks[0].HasPartialPosTag("num:") && taggedToks[1].HasPartialPosTag("num:"))) {
				isCompound = true
			}
		case 3:
			// "polsko-niemiecko-indonezyjski"
			if taggedToks[0] != nil && taggedToks[1] != nil && taggedToks[2] != nil &&
				taggedToks[0].HasPosTag("adja") &&
				taggedToks[1].HasPosTag("adja") &&
				taggedToks[2].HasPartialPosTag("adj:") {
				isCompound = true
			}
		}
		if isCompound {
			var out []string
			for i, p := range tokenParts {
				out = append(out, p)
				if i != len(tokenParts)-1 {
					out = append(out, "-")
				}
			}
			return out
		}
	}
	return []string{token}
}

// splitKeepDelims ports StringTokenizer(text, delims, true).
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
