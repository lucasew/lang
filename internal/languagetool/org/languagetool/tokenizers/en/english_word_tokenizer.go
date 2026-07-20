package en

import (
	"regexp"
	"strings"
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
	// Java EnglishWordTokenizer.wordsToAdd camel-case hyphen exceptions only.
	javaHyphenExceptions = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true, "anti-ivg": true, "anti-uv": true,
		"anti-vih": true, "al-qaida": true,
	}
)

// IsTaggedEN optional EnglishTagger.INSTANCE.tag(...).isTagged() hook.
// Java keeps hyphen/apostrophe compounds only when EnglishTagger tags them.
// Without a tagger, miss — do not invent soft keep lists for doin'/ne'er/etc.
var IsTaggedEN func(s string) bool

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

// wordsToAdd ports EnglishWordTokenizer.wordsToAdd.
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
			// Java: EnglishTagger.INSTANCE.tag(...).isTagged() OR equalsIgnoreCase exceptions.
			if isTaggedEN(normalized) || javaHyphenExceptions[strings.ToLower(s)] {
				l = append(l, s)
			} else {
				// Java: split on ’ and ' only (not hyphen) — StringTokenizer(s, "’'", true)
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

func isTaggedEN(s string) bool {
	// Java: EnglishTagger.INSTANCE.tag(...).isTagged(). Without a tagger, miss —
	// do not invent soft keep for @, doin', ne'er, etc. (emails via JoinEMailsAndUrls).
	if IsTaggedEN != nil {
		return IsTaggedEN(s)
	}
	return false
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
	return out
}
