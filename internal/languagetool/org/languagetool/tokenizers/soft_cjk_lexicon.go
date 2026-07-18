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
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			for _, m := range softTokenRE.FindAllSubmatch(data, -1) {
				attrs := string(m[1])
				w := strings.TrimSpace(string(m[2]))
				if w == "" {
					continue
				}
				// Expand simple | alternatives; strip trivial .* prefix REs
				// (e.g. .*琅琅 → 琅琅 for LANG1_LANG2 soft matching).
				parts := []string{w}
				if strings.Contains(attrs, `regexp="yes"`) {
					if strings.HasPrefix(w, ".*") && !strings.ContainsAny(w[2:], ".*+?[](){}\\|") {
						parts = []string{w[2:]}
					} else if strings.Contains(w, "|") && !strings.ContainsAny(w, ".*+?[](){}\\") {
						parts = strings.Split(w, "|")
					} else {
						continue
					}
				} else if strings.Contains(w, "|") && !strings.ContainsAny(w, ".*+?[](){}\\") {
					parts = strings.Split(w, "|")
				}
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part == "" {
						continue
					}
					n := utf8.RuneCountInString(part)
					if n < 2 {
						continue
					}
					if isAllASCIIWord(part) {
						continue
					}
					lex.Words[part] = struct{}{}
					if n > lex.MaxLen {
						lex.MaxLen = n
					}
				}
			}
		}
	}
	// Bust cache only on first load; tests may need clear — accept process lifetime.
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
	envKey := "LANG_" + strings.ToUpper(lang) + "_SOFT_GRAMMAR"
	if p := os.Getenv(envKey); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	name := lang + "-upstream-soft.xml"
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 12; i++ {
		for _, rel := range []string{
			filepath.Join("testdata", "grammar", name),
			filepath.Join("testdata", "upstream", lang, lang+"-from-upstream-soft.xml"),
		} {
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

// SegmentCJKLongestMatch splits text using longest-match against lex for CJK,
// kana runs when unknown, Latin/digit runs (excluding CJK letters), and
// punctuation as single tokens.
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
			// longest lexicon match first
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
			if matched {
				continue
			}
			// Unknown CJK: single character (soft rules often need ず+らい
			// and 呼+べ separately when not in the lexicon).
			out = append(out, string(r))
			i++
			continue
		}
		// Latin / digit runs — must NOT swallow CJK (Han is unicode.IsLetter).
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
		// Other letters (e.g. Arabic) as runs excluding CJK
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
