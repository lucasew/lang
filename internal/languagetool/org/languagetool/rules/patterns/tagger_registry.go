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

// IsUnknownToTagger ports MatchState.toFinalString suppress_misspelled check:
//
//	AnalyzedToken t0 = tagger.tag(...).get(i).getAnalyzedToken(0);
//	if (t0.getLemma() == null && t0.hasNoTag()) → MISTAKE
//
// AnalyzedToken.hasNoTag: posTag==null || SENTENCE_END || PARAGRAPH_END
// (SENTENCE_START is a real tag — not hasNoTag).
// Only the first reading is consulted (Java getAnalyzedToken(0)).
// No registered tagger → false (leave form; incomplete, no invent misspell).
// Empty tagger result → true (unknown / no readings).
func IsUnknownToTagger(lang, word string) bool {
	if word == "" {
		return true
	}
	fn := languageTaggerFn(lang)
	if fn == nil {
		return false
	}
	// Java tags each formatted string as one unit (including multiword +DT forms).
	tags := fn(word)
	if len(tags) == 0 {
		return true
	}
	// First reading only (getAnalyzedToken(0)).
	t0 := tags[0]
	// hasNoTag: null POS or sentence/paragraph end markers (not SENTENCE_START).
	hasNoTag := t0.POS == "" ||
		t0.POS == languagetool.SentenceEndTagName ||
		t0.POS == languagetool.ParagraphEndTagName
	// lemma == null (empty string is Go's stand-in for Java null lemma)
	lemmaNull := t0.Lemma == ""
	return lemmaNull && hasNoTag
}
