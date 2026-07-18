package tokenizers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// SoftCJKLexicon is a set of multi-rune surfaces taken from soft grammar packs
// (ja-upstream-soft.xml / zh-upstream-soft.xml). Used for longest-match
// segmentation approximating Java Sen/HanLP dictionaries when those are not embedded.
type SoftCJKLexicon struct {
	Words  map[string]struct{}
	MaxLen int // max rune length of a lexicon word
}

var (
	softCJKMu    sync.Mutex
	softCJKCache = map[string]*SoftCJKLexicon{}
)

// Capture token open tag + body so we can skip regexp="yes" pure patterns carefully.
var softTokenRE = regexp.MustCompile(`<token(\s[^>]*)?>([^<]+)</token>`)

// SoftCJKLexiconForLang loads or returns a cached lexicon for "ja" or "zh".
// Discovery: LANG_{JA,ZH}_SOFT_GRAMMAR, then walk-up testdata/grammar/{lang}-upstream-soft.xml.
func SoftCJKLexiconForLang(lang string) *SoftCJKLexicon {
	base := strings.ToLower(lang)
	if i := strings.IndexByte(base, '-'); i > 0 {
		base = base[:i]
	}
	if base != "ja" && base != "zh" {
		return nil
	}
	softCJKMu.Lock()
	defer softCJKMu.Unlock()
	if lex, ok := softCJKCache[base]; ok {
		return lex
	}
	path := discoverSoftGrammar(base)
	lex := &SoftCJKLexicon{Words: map[string]struct{}{}, MaxLen: 1}
	add := func(part string) {
		part = strings.TrimSpace(part)
		if part == "" {
			return
		}
		n := utf8.RuneCountInString(part)
		if n < 2 || n > 12 {
			return
		}
		if isAllASCIIWord(part) {
			return
		}
		// Require at least one CJK rune.
		hasCJK := false
		for _, r := range part {
			if isHan(r) || isHiragana(r) || isKatakana(r) {
				hasCJK = true
				break
			}
		}
		if !hasCJK {
			return
		}
		lex.Words[part] = struct{}{}
		if n > lex.MaxLen {
			lex.MaxLen = n
		}
	}
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			for _, m := range softTokenRE.FindAllSubmatch(data, -1) {
				attrs := string(m[1])
				w := strings.TrimSpace(string(m[2]))
				for _, part := range softLexiconParts(w, attrs) {
					add(part)
				}
			}
		}
	}
	softCJKCache[base] = lex
	return lex
}

// ClearSoftCJKLexiconCache is for tests that change soft grammar paths.
func ClearSoftCJKLexiconCache() {
	softCJKMu.Lock()
	softCJKCache = map[string]*SoftCJKLexicon{}
	softCJKMu.Unlock()
}

func discoverSoftGrammar(lang string) string {
	return walkUpSoftFile(lang, []string{
		filepath.Join("testdata", "grammar", lang+"-upstream-soft.xml"),
		filepath.Join("testdata", "upstream", lang, lang+"-from-upstream-soft.xml"),
	}, "LANG_"+strings.ToUpper(lang)+"_SOFT_GRAMMAR")
}

func walkUpSoftFile(lang string, rels []string, envKey string) string {
	if p := os.Getenv(envKey); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 12; i++ {
		for _, rel := range rels {
			p := filepath.Join(dir, rel)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func isAllASCIIWord(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '\'') {
			return false
		}
	}
	return true
}

// softLexiconParts extracts literal CJK multi-char surfaces from a soft <token> body.
func softLexiconParts(w, attrs string) []string {
	w = strings.TrimSpace(w)
	if w == "" {
		return nil
	}
	// (有|任何)? → 有, 任何
	if strings.HasPrefix(w, "(") && (strings.HasSuffix(w, ")?") || strings.HasSuffix(w, ")*") || strings.HasSuffix(w, ")+")) {
		inner := w
		switch {
		case strings.HasSuffix(inner, ")?"):
			inner = strings.TrimSuffix(inner, ")?")
		case strings.HasSuffix(inner, ")*"):
			inner = strings.TrimSuffix(inner, ")*")
		default:
			inner = strings.TrimSuffix(inner, ")+")
		}
		inner = strings.TrimPrefix(inner, "(")
		if !strings.ContainsAny(inner, ".*+?[]{}\\") {
			return splitBar(inner)
		}
	}
	if strings.Contains(attrs, `regexp="yes"`) {
		if strings.HasPrefix(w, ".*") && !strings.ContainsAny(w[2:], ".*+?[](){}\\|") {
			return []string{w[2:]}
		}
		if strings.Contains(w, "|") && !strings.ContainsAny(w, ".*+?[](){}\\") {
			return splitBar(w)
		}
		return nil
	}
	if strings.Contains(w, "|") && !strings.ContainsAny(w, ".*+?[](){}\\") {
		return splitBar(w)
	}
	return []string{w}
}

func splitBar(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "|") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// SegmentCJKLongestMatch splits text using forward maximum matching against
// lex for CJK (Han/kana), with a light look-ahead so e.g. し+いっそう wins over
// しい when いっそう is a lexicon word. Latin/digit runs exclude CJK.
func SegmentCJKLongestMatch(text string, lex *SoftCJKLexicon) []string {
	if text == "" {
		return nil
	}
	words := map[string]struct{}{}
	maxLen := 1
	if lex != nil {
		words = lex.Words
		maxLen = lex.MaxLen
		if maxLen < 1 {
			maxLen = 1
		}
	}
	runes := []rune(text)
	var out []string
	i := 0
	for i < len(runes) {
		r := runes[i]
		if unicode.IsSpace(r) {
			out = append(out, string(r))
			i++
			continue
		}
		if isHan(r) || isHiragana(r) || isKatakana(r) {
			limit := maxLen
			if rem := len(runes) - i; rem < limit {
				limit = rem
			}
			matched := false
			for L := limit; L >= 2; L-- {
				cand := string(runes[i : i+L])
				if _, ok := words[cand]; ok {
					out = append(out, cand)
					i += L
					matched = true
					break
				}
			}
			if !matched {
				out = append(out, string(r))
				i++
			}
			continue
		}
		if isASCIILetterOrDigit(r) {
			j := i + 1
			for j < len(runes) {
				if isASCIILetterOrDigit(runes[j]) {
					j++
					continue
				}
				break
			}
			out = append(out, string(runes[i:j]))
			i = j
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			j := i + 1
			for j < len(runes) {
				rj := runes[j]
				if (unicode.IsLetter(rj) || unicode.IsDigit(rj)) && !isHan(rj) && !isHiragana(rj) && !isKatakana(rj) {
					j++
					continue
				}
				break
			}
			out = append(out, string(runes[i:j]))
			i = j
			continue
		}
		out = append(out, string(r))
		i++
	}
	return out
}

func isHan(r rune) bool      { return unicode.Is(unicode.Han, r) }
func isHiragana(r rune) bool { return unicode.In(r, unicode.Hiragana) }
func isKatakana(r rune) bool { return unicode.In(r, unicode.Katakana) }

func isASCIILetterOrDigit(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
