package pipeline

import (
	"regexp"
	"strings"
)

// English word-character class from EnglishWordTokenizer.
// Apostrophes are protected via placeholders before matching (as in LT).
const enWordChars = `±§©@€£¥\$\p{L}\d\x{0300}-\x{036F}\x{00A8}°%‰‱&\x{FFFD}\x{00AD}\x{00AC}\x{FF0C}\x{FF1F}-`

var (
	enTokenizerPattern = regexp.MustCompile(`[` + enWordChars + `]+|[^` + enWordChars + `]`)
	enContractions     = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(fo['’]c['’]sle|rec['’][ds]|OK['’]d|cc['’][ds]|DJ['’][d]|[pd]m['’]d|rsvp['’]d)$`),
		regexp.MustCompile(`(?i)^(['’]?)(are|is|were|was|do|does|did|have|has|had|wo|would|ca|could|sha|should|must|ai|ought|might|need|may|am|dare|das|dass|hai|used|use)(n['’]t)$`),
		regexp.MustCompile(`(?i)^(.+)(['’]m|['’]re|['’]ll|['’]ve|['’]d|['’]s)(['’-]?)$`),
		regexp.MustCompile(`(?i)^(['’]t)(was|were|is)$`),
	}
)

const (
	phTypeW = "xxAPOSTYPEWxx"
	phTypeG = "xxAPOSTYPOGxx"
)

// EnglishWordTokenize ports EnglishWordTokenizer.tokenize (contraction splits).
// Offsets are rune indices into the original text.
func EnglishWordTokenize(text string) []Token {
	if text == "" {
		return nil
	}
	aux := strings.ReplaceAll(text, "'", phTypeW)
	aux = strings.ReplaceAll(aux, "’", phTypeG)

	var pieces []string
	for _, loc := range enTokenizerPattern.FindAllStringIndex(aux, -1) {
		s := aux[loc[0]:loc[1]]
		s = strings.ReplaceAll(s, phTypeW, "'")
		s = strings.ReplaceAll(s, phTypeG, "’")
		if strings.ContainsAny(s, "'’") {
			split := false
			for _, pat := range enContractions {
				if m := pat.FindStringSubmatch(s); m != nil {
					for i := 1; i < len(m); i++ {
						if m[i] != "" {
							pieces = append(pieces, m[i])
						}
					}
					split = true
					break
				}
			}
			if split {
				continue
			}
			pieces = append(pieces, splitKeep(s, "'’")...)
			continue
		}
		pieces = append(pieces, s)
	}
	return assignOffsets(text, pieces)
}

func splitKeep(s, seps string) []string {
	var out []string
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(seps, r) {
			if b.Len() > 0 {
				out = append(out, b.String())
				b.Reset()
			}
			out = append(out, string(r))
		} else {
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

func assignOffsets(text string, pieces []string) []Token {
	runes := []rune(text)
	var tokens []Token
	pos := 0
	for _, p := range pieces {
		pr := []rune(p)
		found := -1
		for i := pos; i+len(pr) <= len(runes); i++ {
			match := true
			for j := range pr {
				if runes[i+j] != pr[j] {
					match = false
					break
				}
			}
			if match {
				found = i
				break
			}
		}
		if found < 0 {
			continue
		}
		if found > pos {
			tokens = append(tokens, tokenFromRunes(runes, pos, found)...)
		}
		start, end := found, found+len(pr)
		tok := Token{Text: p, Start: start, End: end}
		if isOnlySpaceRunes(pr) {
			tok.Whitespace = true
			if p == "\n" || p == "\r" {
				tok.Linebreak = true
			}
		}
		tokens = append(tokens, tok)
		pos = end
	}
	if pos < len(runes) {
		tokens = append(tokens, tokenFromRunes(runes, pos, len(runes))...)
	}
	return tokens
}

func tokenFromRunes(runes []rune, start, end int) []Token {
	var out []Token
	i := start
	for i < end {
		r := runes[i]
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\u00A0' {
			out = append(out, Token{
				Text: string(r), Start: i, End: i + 1,
				Whitespace: true, Linebreak: r == '\n' || r == '\r',
			})
			i++
			continue
		}
		j := i + 1
		for j < end {
			rj := runes[j]
			if rj == ' ' || rj == '\t' || rj == '\n' || rj == '\r' || rj == '\u00A0' {
				break
			}
			j++
		}
		out = append(out, Token{Text: string(runes[i:j]), Start: i, End: j})
		i = j
	}
	return out
}

func isOnlySpaceRunes(pr []rune) bool {
	if len(pr) == 0 {
		return false
	}
	for _, r := range pr {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' && r != '\u00A0' {
			return false
		}
	}
	return true
}

// WordTokenizeForLang chooses English tokenizer for en, else generic WordTokenize.
func WordTokenizeForLang(langFamily, text string) []Token {
	if langFamily == "en" {
		return EnglishWordTokenize(text)
	}
	return WordTokenize(text)
}
