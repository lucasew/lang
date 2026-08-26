package patterns

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Language synthesizers for MatchState.toFinalString (Java Language.getSynthesizer).
// Nil means surface-only path (no invent forms).
var (
	synthMu       sync.RWMutex
	langSynths    = map[string]synthesis.Synthesizer{}
	defaultSynth  synthesis.Synthesizer
)

// RegisterLanguageSynthesizer ports Language module synthesizer registration.
// lang may be "en" or "en-US" (base code also registered).
func RegisterLanguageSynthesizer(lang string, s synthesis.Synthesizer) {
	if lang == "" || s == nil {
		return
	}
	synthMu.Lock()
	defer synthMu.Unlock()
	lang = strings.ToLower(lang)
	langSynths[lang] = s
	if i := strings.IndexByte(lang, '-'); i > 0 {
		langSynths[lang[:i]] = s
	}
}

// SetDefaultSynthesizer sets a fallback synthesizer when language is unknown.
func SetDefaultSynthesizer(s synthesis.Synthesizer) {
	synthMu.Lock()
	defer synthMu.Unlock()
	defaultSynth = s
}

// LanguageSynthesizer returns the synthesizer for lang, or nil.
func LanguageSynthesizer(lang string) synthesis.Synthesizer {
	synthMu.RLock()
	defer synthMu.RUnlock()
	if lang == "" {
		return defaultSynth
	}
	lang = strings.ToLower(lang)
	if s, ok := langSynths[lang]; ok {
		return s
	}
	if i := strings.IndexByte(lang, '-'); i > 0 {
		if s, ok := langSynths[lang[:i]]; ok {
			return s
		}
	}
	return defaultSynth
}
