package patterns

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// TagWordFn ports Language.getTagger().tag for single-token spell checks in MatchState.
type TagWordFn func(token string) []languagetool.TokenTag

var (
	tagMu      sync.RWMutex
	langTagger = map[string]TagWordFn{}
)

// RegisterLanguageTagger registers a POS tagger used by suppress_misspelled checks
// (Java: lang.getTagger().tag(formattedStringElements)).
func RegisterLanguageTagger(lang string, fn TagWordFn) {
	if lang == "" || fn == nil {
		return
	}
	tagMu.Lock()
	defer tagMu.Unlock()
	lang = strings.ToLower(lang)
	langTagger[lang] = fn
	if i := strings.IndexByte(lang, '-'); i > 0 {
		langTagger[lang[:i]] = fn
	}
}

// languageTaggerFn returns the registered tagger, or nil.
func languageTaggerFn(lang string) TagWordFn {
	tagMu.RLock()
	defer tagMu.RUnlock()
	if lang == "" {
		return nil
	}
	lang = strings.ToLower(lang)
	if fn, ok := langTagger[lang]; ok {
		return fn
	}
	if i := strings.IndexByte(lang, '-'); i > 0 {
		if fn, ok := langTagger[lang[:i]]; ok {
			return fn
		}
	}
	return nil
}

// LanguageTagWord returns tagger readings for token, or nil if no tagger registered.
func LanguageTagWord(lang, token string) []languagetool.TokenTag {
	fn := languageTaggerFn(lang)
	if fn == nil {
		return nil
	}
	return fn(token)
}

// IsUnknownToTagger ports Java lemma==null && hasNoTag for a surface form.
// No registered tagger → false (leave form; no invent misspell).
// Registered tagger with empty readings → true (unknown).
func IsUnknownToTagger(lang, word string) bool {
	if word == "" {
		return true
	}
	fn := languageTaggerFn(lang)
	if fn == nil {
		return false
	}
	// Java tags the whole string as one unit (including multiword +DT forms).
	tags := fn(word)
	if len(tags) == 0 {
		return true
	}
	for _, t := range tags {
		if t.POS != "" && t.POS != languagetool.SentenceStartTagName &&
			t.POS != languagetool.SentenceEndTagName && t.POS != languagetool.ParagraphEndTagName {
			return false
		}
	}
	return true
}
