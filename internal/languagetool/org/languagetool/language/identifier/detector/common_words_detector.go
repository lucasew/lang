package detector

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"sync"
)

// CommonWordsDetector ports org.languagetool.language.identifier.detector.CommonWordsDetector
// with a pluggable word→language-code dictionary (no Languages registry).
type CommonWordsDetector struct {
	mu        sync.RWMutex
	word2Lang map[string][]string // lowercase word → lang short codes
}

var (
	numberPattern  = regexp.MustCompile(`^[0-9.,%-]+$`)
	spanishPattern = regexp.MustCompile(`^[a-zñ]+(ón|cion|aban|ábamos|ábais|íamos|íais|[úí]a[sn]?|úe[ns]?)$`)
	notSpanishPat  = regexp.MustCompile(`^[lmndts]['’].*$|^.*(ns|[áéó].i[oa]s?)$|^.*(ss|[çàèòïâêôãõìù]|l·l).*$`)
	notCatalanPat  = regexp.MustCompile(`^.*([áéó].i[oa]s?|d[oa]s)$|^.*[áâêôãõìùñ].*$`)
	portuguesePat  = regexp.MustCompile(`^.*([áó]ri[oa]|ério)s?$`)
	// Match Java PUNCT_PATTERN: hyphen is NOT a member (the `-` between `}` and `«`
	// is a character range in Java, not a literal hyphen), so "Autohaus-Wirklichkeit"
	// stays hyphenated until SPACE_OR_HYPHEN_PATTERN splits it.
	punctPattern = regexp.MustCompile(`[(),.:;!?„“"¡¿\s\[\]{}«»”]`)
	charsPattern   = regexp.MustCompile(`\p{L}+$`)
	spaceOrHyphen  = regexp.MustCompile(`[ -]`)
)

func NewCommonWordsDetector() *CommonWordsDetector {
	return &CommonWordsDetector{word2Lang: map[string][]string{}}
}

// LoadWords adds common words for a language short code from a line-oriented reader.
func (d *CommonWordsDetector) LoadWords(langShortCode string, r io.Reader) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key := strings.ToLower(line)
		if len(key) == 1 && key[0] <= ' ' {
			continue
		}
		langs := d.word2Lang[key]
		found := false
		for _, l := range langs {
			if l == langShortCode {
				found = true
				break
			}
		}
		if !found {
			d.word2Lang[key] = append(langs, langShortCode)
		}
	}
	return sc.Err()
}

// GetKnownWordsPerLanguage returns counts of common words per language code.
func (d *CommonWordsDetector) GetKnownWordsPerLanguage(text string) map[string]int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := map[string]int{}
	aux := punctPattern.ReplaceAllString(text, " ")
	if !strings.HasSuffix(aux, " ") && strings.Count(aux, " ") > 0 {
		aux = charsPattern.ReplaceAllString(aux, "")
	}
	words := spaceOrHyphen.Split(aux, -1)
	for _, word := range words {
		if numberPattern.MatchString(word) {
			continue
		}
		lc := strings.ToLower(word)
		langs := d.word2Lang[lc]
		if langs != nil {
			for _, lang := range langs {
				result[lang]++
			}
		}
		// Portuguese heuristic
		if (langs == nil || !contains(langs, "pt")) && portuguesePat.MatchString(lc) {
			result["pt"]++
		}
		// Spanish heuristic
		if (langs == nil || !contains(langs, "es")) && spanishPattern.MatchString(lc) {
			result["es"]++
		}
		if (langs == nil || !contains(langs, "es")) && notSpanishPat.MatchString(lc) {
			result["es"]--
		}
		// Catalan heuristic (negative only in Java for notCatalan)
		if (langs == nil || !contains(langs, "ca")) && notCatalanPat.MatchString(lc) {
			result["ca"]--
		}
	}
	return result
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
