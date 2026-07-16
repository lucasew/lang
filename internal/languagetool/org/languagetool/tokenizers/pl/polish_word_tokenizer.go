package pl

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PolishWordTokenizer ports org.languagetool.tokenizers.pl.PolishWordTokenizer
// without a tagger (hyphen compounds kept whole unless prefixes/digits apply).
// Call SetTagger later for hybrid compound splitting.
type PolishWordTokenizer struct {
	plTokenizing string
	// tagger optional; nil matches Java before setTagger()
	tagger PolishHyphenTagger
}

// PolishHyphenTagger is the subset of Tagger used for hyphen compounds.
type PolishHyphenTagger interface {
	// Tag returns readings for tokens; last entry is the full compound.
	// isTagged / hasPosTag / hasPartialPosTag inspect readings.
	Tag(tokens []string) []PolishTokenReadings
}

// PolishTokenReadings minimal readings for hyphen decisions.
type PolishTokenReadings struct {
	IsTagged       bool
	HasAdja        bool // pos tag "adja"
	HasAdjPartial  bool // partial "adj:"
	HasSubstPartial bool
	HasNumPartial  bool
}

func NewPolishWordTokenizer() *PolishWordTokenizer {
	return &PolishWordTokenizer{
		plTokenizing: tokenizers.TokenizingCharacters() + "–‚", // n-dash
	}
}

func (w *PolishWordTokenizer) SetTagger(t PolishHyphenTagger) {
	w.tagger = t
}

var polishPrefixes = map[string]bool{
	"arcy": true, "neo": true, "pre": true, "anty": true, "eks": true, "bez": true,
	"beze": true, "ekstra": true, "hiper": true, "infra": true, "kontr": true,
	"maksi": true, "midi": true, "między": true, "mini": true, "nad": true,
	"nade": true, "około": true, "ponad": true, "post": true, "pro": true,
	"przeciw": true, "pseudo": true, "super": true, "śród": true, "ultra": true,
	"wice": true, "wokół": true, "wokoło": true,
}

func (w *PolishWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, w.plTokenizing)
	var l []string
	for _, token := range raw {
		if utf8.RuneCountInString(token) > 1 {
			if strings.HasSuffix(token, "-") {
				l = append(l, token[:len(token)-1], "-")
			} else if token[0] == '-' {
				l = append(l, "-")
				l = append(l, w.Tokenize(token[1:])...)
			} else if strings.Contains(token, "-") {
				tokenParts := strings.Split(token, "-")
				// Number ranges (1-23) always split, even without a tagger.
				if len(tokenParts) > 0 && len(tokenParts[len(tokenParts)-1]) > 0 &&
					unicode.IsDigit(rune(tokenParts[len(tokenParts)-1][0])) {
					for i, p := range tokenParts {
						l = append(l, p)
						if i != len(tokenParts)-1 {
							l = append(l, "-")
						}
					}
				} else if polishPrefixes[tokenParts[0]] || w.tagger == nil {
					l = append(l, token)
				} else {
					// tagger-based compound split
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

func (w *PolishWordTokenizer) splitCompoundWithTagger(token string, tokenParts []string) []string {
	if w.tagger == nil {
		return []string{token}
	}
	tested := append(append([]string{}, tokenParts...), token)
	tagged := w.tagger.Tag(tested)
	if len(tagged) == len(tokenParts)+1 && !tagged[len(tokenParts)].IsTagged {
		isCompound := false
		switch len(tokenParts) {
		case 2:
			if (tagged[0].HasAdja && tagged[1].HasAdjPartial) ||
				(tagged[0].HasSubstPartial && tagged[1].HasSubstPartial) ||
				(tagged[0].HasNumPartial && tagged[1].HasNumPartial) {
				isCompound = true
			}
		case 3:
			if tagged[0].HasAdja && tagged[1].HasAdja && tagged[2].HasAdjPartial {
				isCompound = true
			}
		}
		if isCompound {
			var l []string
			for i, p := range tokenParts {
				l = append(l, p)
				if i != len(tokenParts)-1 {
					l = append(l, "-")
				}
			}
			return l
		}
	}
	return []string{token}
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
