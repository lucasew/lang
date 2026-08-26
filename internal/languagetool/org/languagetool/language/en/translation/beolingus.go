package translation

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"regexp"
	"strings"
)

var (
	reBrackets   = regexp.MustCompile(`\[.*?\]`)
	reBraces     = regexp.MustCompile(`\{.*?\}`)
	reParens     = regexp.MustCompile(`\(.*?\)`)
	reAbbrev     = regexp.MustCompile(`/[A-Z]+/`)
	reAbbrevWord = regexp.MustCompile(` /[A-Z][a-z]+\.?/`)
	reAngle      = regexp.MustCompile(`<(.*)>`)
	reSpaces     = regexp.MustCompile(`\s+`)
)

// verbsWithTo — keep "to " after these prev words (Java BeoLingusTranslator.verbsWithTo subset).
var verbsWithTo = map[string]struct{}{
	"want": {}, "need": {}, "have": {}, "like": {}, "try": {},
}

// SplitAtSemicolon ports BeoLingusTranslator.splitAtSemicolon — split on "; " unless inside {...}.
func SplitAtSemicolon(s string) []string {
	parts := regexp.MustCompile(`;\s+`).Split(s, -1)
	var merged []string
	merging := false
	for _, item := range parts {
		item = tools.JavaStringTrim(item)
		openPos := strings.IndexByte(item, '{')
		closePos := strings.IndexByte(item, '}')
		if merging {
			if len(merged) == 0 {
				merged = append(merged, item)
			} else {
				merged[len(merged)-1] = merged[len(merged)-1] + "; " + item
			}
			if closePos >= 0 {
				merging = false
			}
			continue
		}
		if openPos > closePos {
			// ";" inside "{...}" — start merge
			merged = append(merged, item)
			merging = true
			continue
		}
		merged = append(merged, item)
	}
	return merged
}

// Split ports BeoLingusTranslator.split (currently just splitAtSemicolon).
func Split(s string) []string {
	return SplitAtSemicolon(s)
}

// CleanTranslationForReplace ports BeoLingusTranslator.cleanTranslationForReplace.
func CleanTranslationForReplace(s, prevWord string) string {
	clean := reBrackets.ReplaceAllString(s, "")
	clean = reBraces.ReplaceAllString(clean, "")
	clean = reParens.ReplaceAllString(clean, "")
	clean = strings.ReplaceAll(clean, "sth./sb.", "")
	clean = strings.ReplaceAll(clean, "sb./sth.", "")
	clean = strings.ReplaceAll(clean, "sth.", "")
	clean = strings.ReplaceAll(clean, "sb.", "")
	clean = reAbbrev.ReplaceAllString(clean, "")
	clean = reAbbrevWord.ReplaceAllString(clean, "")
	clean = strings.ReplaceAll(clean, "<> ", "")
	clean = reAngle.ReplaceAllString(clean, "")
	clean = reSpaces.ReplaceAllString(clean, " ")
	clean = tools.JavaStringTrim(clean)
	if prevWord == "to" && strings.HasPrefix(clean, "to ") {
		return clean[3:]
	}
	if prevWord != "to" && strings.HasPrefix(clean, "to ") {
		if _, ok := verbsWithTo[prevWord]; !ok {
			return clean[3:]
		}
	}
	return clean
}

// BeoLingusTranslator is an inject-friendly dictionary surface
// (full Beolingus file deferred).
type BeoLingusTranslator struct {
	// Dict maps German source lemma → English translation entries (raw strings).
	Dict map[string][]string
	// Inflected maps surface form → base lemma used for lookup.
	Inflected map[string]string
}

// NewBeoLingusTranslator builds an empty inject translator.
func NewBeoLingusTranslator() *BeoLingusTranslator {
	return &BeoLingusTranslator{
		Dict:      map[string][]string{},
		Inflected: map[string]string{},
	}
}

// Translate looks up German word (or lemma) and returns cleaned English variants.
func (t *BeoLingusTranslator) Translate(german, prevWord string) []string {
	if t == nil || german == "" {
		return nil
	}
	key := german
	if t.Inflected != nil {
		if base, ok := t.Inflected[german]; ok {
			key = base
		}
	}
	raw, ok := t.Dict[key]
	if !ok {
		// case fold soft
		for k, v := range t.Dict {
			if strings.EqualFold(k, key) {
				raw = v
				ok = true
				break
			}
		}
	}
	if !ok {
		return nil
	}
	var out []string
	for _, r := range raw {
		for _, part := range Split(r) {
			c := CleanTranslationForReplace(part, prevWord)
			if c != "" {
				out = append(out, c)
			}
		}
	}
	return out
}

// TranslateInflectedForm ports translate of an inflected DE form via Inflected map.
func (t *BeoLingusTranslator) TranslateInflectedForm(surface, prevWord string) []string {
	return t.Translate(surface, prevWord)
}

// AmericanToBritish soft maps common AE→BE spelling in translations.
func AmericanToBritish(s string) string {
	repl := map[string]string{
		"color": "colour", "favor": "favour", "center": "centre",
		"theater": "theatre", "organize": "organise",
	}
	low := strings.ToLower(s)
	if b, ok := repl[low]; ok {
		// preserve simple casing of first letter
		if s != "" && s[0] >= 'A' && s[0] <= 'Z' {
			return strings.ToUpper(b[:1]) + b[1:]
		}
		return b
	}
	return s
}

// GetTranslationSuffix ports BeoLingusTranslator.getTranslationSuffix.
func GetTranslationSuffix(s string) string {
	var sb strings.Builder
	var lookingFor []byte
	contains := func(c byte) bool {
		for _, x := range lookingFor {
			if x == c {
				return true
			}
		}
		return false
	}
	remove := func(c byte) {
		out := lookingFor[:0]
		for _, x := range lookingFor {
			if x != c {
				out = append(out, x)
			}
		}
		lookingFor = out
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '[':
			lookingFor = append(lookingFor, ']')
		case ']':
			if contains(']') {
				sb.WriteByte(c)
				sb.WriteByte(' ')
				remove(']')
			}
		case '<':
			lookingFor = append(lookingFor, '>')
		case '>':
			if contains('>') {
				sb.WriteByte(c)
				sb.WriteByte(' ')
				remove('>')
			}
		case '(':
			lookingFor = append(lookingFor, ')')
		case ')':
			if contains(')') {
				sb.WriteByte(c)
				sb.WriteByte(' ')
				remove(')')
			}
		case '{':
			lookingFor = append(lookingFor, '}')
		case '}':
			if contains('}') {
				sb.WriteByte(c)
				sb.WriteByte(' ')
				remove('}')
			}
		}
		if len(lookingFor) > 0 {
			sb.WriteByte(c)
		}
	}
	return tools.JavaStringTrim(sb.String())
}
