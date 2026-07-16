package spelling

// Constants from org.languagetool.rules.spelling.SpellingCheckRule.
const (
	HighConfidence     = float32(0.99)
	LanguageTool       = "LanguageTool"
	LanguageTooler     = "LanguageTooler"
	MaxTokenLength     = 200
	SpellingIgnoreFile = "/hunspell/ignore.txt"
	SpellingFile       = "/hunspell/spelling.txt"
	CustomSpellingFile = "/hunspell/spelling_custom.txt"
	GlobalSpellingFile = "spelling_global.txt"
)

// SpellingCheckRule is a surface for spellcheck rules (full match deferred).
type SpellingCheckRule struct {
	ID           string
	Description  string
	LanguageCode string
	// IsMisspelled returns true if word is not in the dictionary.
	IsMisspelled func(word string) bool
	// IgnoreWords is a set of words to accept.
	IgnoreWords map[string]struct{}
}

func NewSpellingCheckRule(id, description, languageCode string) *SpellingCheckRule {
	return &SpellingCheckRule{
		ID:           id,
		Description:  description,
		LanguageCode: languageCode,
		IgnoreWords:  map[string]struct{}{},
	}
}

func (r *SpellingCheckRule) GetID() string          { return r.ID }
func (r *SpellingCheckRule) GetDescription() string { return r.Description }

// AcceptWord reports whether word should not be flagged (ignore list or not misspelled).
func (r *SpellingCheckRule) AcceptWord(word string) bool {
	if r == nil {
		return false
	}
	if len(word) > MaxTokenLength {
		return true
	}
	if _, ok := r.IgnoreWords[word]; ok {
		return true
	}
	if r.IsMisspelled == nil {
		return true
	}
	return !r.IsMisspelled(word)
}

// AddIgnoreWords adds words to the ignore set.
func (r *SpellingCheckRule) AddIgnoreWords(words ...string) {
	if r.IgnoreWords == nil {
		r.IgnoreWords = map[string]struct{}{}
	}
	for _, w := range words {
		r.IgnoreWords[w] = struct{}{}
	}
}
