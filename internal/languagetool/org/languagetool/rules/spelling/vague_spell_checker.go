package spelling

import "sync"

// WordValidator is a language-scoped dictionary probe.
type WordValidator func(word string) bool

// VagueSpellChecker ports org.languagetool.rules.spelling.VagueSpellChecker.
// Fast validity check without suggestions; inject per-language validators.
type VagueSpellChecker struct {
	mu         sync.Mutex
	validators map[string]WordValidator // language code → validator
}

func NewVagueSpellChecker() *VagueSpellChecker {
	return &VagueSpellChecker{validators: map[string]WordValidator{}}
}

// Register sets a validity probe for languageCode (short code with optional variant).
func (v *VagueSpellChecker) Register(languageCode string, isValid WordValidator) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.validators[languageCode] = isValid
}

// IsValidWord ports VagueSpellChecker.isValidWord.
// Returns false if no validator is registered for the language.
func (v *VagueSpellChecker) IsValidWord(word, languageCode string) bool {
	if v == nil || word == "" {
		return false
	}
	v.mu.Lock()
	fn := v.validators[languageCode]
	if fn == nil {
		// try short code
		if i := indexDash(languageCode); i >= 0 {
			fn = v.validators[languageCode[:i]]
		}
	}
	v.mu.Unlock()
	if fn == nil {
		return false
	}
	return fn(word)
}

func indexDash(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '-' || s[i] == '_' {
			return i
		}
	}
	return -1
}
