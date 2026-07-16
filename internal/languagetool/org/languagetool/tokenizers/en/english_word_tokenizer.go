package en

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// EnglishWordTokenizer ports org.languagetool.tokenizers.en.EnglishWordTokenizer.
type EnglishWordTokenizer struct{}

func NewEnglishWordTokenizer() *EnglishWordTokenizer {
	return &EnglishWordTokenizer{}
}

// wordCharacters from Java (simplified as character class for matching runs)
const wordCharactersClass = `±§©@€£¥\$\p{L}\d\-\x{0300}-\x{036F}\x{00A8}°%‰‱&\x{FFFD}\x{00AD}\x{00AC}\x{FF0C}\x{FF1F}`

var (
	tokenizerPattern = regexp.MustCompile(`[` + wordCharactersClass + `]+|[^` + wordCharactersClass + `]`)
	singleQuote      = regexp.MustCompile(`'`)
	curlyQuote       = regexp.MustCompile(`’`)
	apostypew        = regexp.MustCompile(`xxAPOSTYPEWxx`)
	apostypog        = regexp.MustCompile(`xxAPOSTYPOGxx`)
	softHyphen       = regexp.MustCompile(`\x{00AD}`)
	patternList      = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(fo['’]c['’]sle|rec['’][ds]|OK['’]d|cc['’][ds]|DJ['’][d]|[pd]m['’]d|rsvp['’]d)$`),
		regexp.MustCompile(`(?i)^(['’]?)(are|is|were|was|do|does|did|have|has|had|wo|would|ca|could|sha|should|must|ai|ought|might|need|may|am|dare|das|dass|hai|used|use)(n['’]t)$`),
		regexp.MustCompile(`(?i)^(.+)(['’]m|['’]re|['’]ll|['’]ve|['’]d|['’]s)(['’-]?)$`),
		regexp.MustCompile(`(?i)^(['’]t)(was|were|is)$`),
	}
	// known hyphenated forms kept whole (from Java wordsToAdd)
	hyphenExceptions = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true, "anti-ivg": true, "anti-uv": true,
		"anti-vih": true, "al-qaida": true,
	}
)

func (w *EnglishWordTokenizer) Tokenize(text string) []string {
	auxText := singleQuote.ReplaceAllString(text, "xxAPOSTYPEWxx")
	auxText = curlyQuote.ReplaceAllString(auxText, "xxAPOSTYPOGxx")
	var l []string
	for _, loc := range tokenizerPattern.FindAllStringIndex(auxText, -1) {
		s := auxText[loc[0]:loc[1]]
		// variation selectors FE00-FE0F attach to previous
		if len(l) > 0 {
			r, _ := utf8.DecodeRuneInString(s)
			if len(s) == utf8.RuneLen(r) && r >= 0xFE00 && r <= 0xFE0F {
				l[len(l)-1] = l[len(l)-1] + s
				continue
			}
		}
		s = apostypew.ReplaceAllString(s, "'")
		s = apostypog.ReplaceAllString(s, "’")
		matchFound := false
		var groups []string
		if strings.Contains(s, "'") || strings.Contains(s, "’") {
			for _, pattern := range patternList {
				if m := pattern.FindStringSubmatch(s); m != nil {
					matchFound = true
					groups = m[1:] // capturing groups
					break
				}
			}
		}
		if matchFound {
			for _, g := range groups {
				if g == "" {
					continue
				}
				l = append(l, wordsToAdd(g)...)
			}
		} else {
			l = append(l, wordsToAdd(s)...)
		}
	}
	return tokenizers.JoinEMailsAndUrls(l)
}

var keepApostropheForm = regexp.MustCompile(`(?i)^(fo['’]c['’]sle|rec['’][ds]|ok['’]d|cc['’][ds]|dj['’][d]|[pd]m['’]d|rsvp['’]d|n['’]t|['’](m|re|ll|ve|d|s)|doin['’]|ne['’]er|e['’]er|o['’]er|jack-o['’]-lantern)$`)

func wordsToAdd(s string) []string {
	var l []string
	hyphensAtEnd := 0
	if s == "" {
		return l
	}
	for strings.HasPrefix(s, "-") {
		l = append(l, "-")
		s = s[1:]
	}
	for strings.HasSuffix(s, "-") {
		s = s[:len(s)-1]
		hyphensAtEnd++
	}
	if s != "" {
		if !strings.Contains(s, "-") && !strings.Contains(s, "'") && !strings.Contains(s, "’") {
			l = append(l, s)
		} else {
			normalized := softHyphen.ReplaceAllString(s, "")
			normalized = curlyQuote.ReplaceAllString(normalized, "'")
			if isTaggedEnglish(normalized) || keepApostropheForm.MatchString(s) || hyphenExceptions[strings.ToLower(s)] {
				l = append(l, s)
			} else {
				// split on ’ and ' only (not hyphen)
				l = append(l, splitKeepDelim(s, "’'")...)
			}
		}
	}
	for hyphensAtEnd > 0 {
		l = append(l, "-")
		hyphensAtEnd--
	}
	return l
}

func isTaggedEnglish(s string) bool {
	// Without full EnglishTagger: keep emails/handles whole.
	return strings.Contains(s, "@")
}

func splitKeepDelim(s, delims string) []string {
	var out []string
	var cur strings.Builder
	for _, r := range s {
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
	_ = unicode.MaxRune
	return out
}
