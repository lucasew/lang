package languagetool

// LinguServices ports org.languagetool.LinguServices — LO extension hooks.
// Default implementations are no-ops / empty (same as Java base class).
type LinguServices struct {
	// ThesaurusRelevantRule stores the last rule set via SetThesaurusRelevantRule
	// (Java base method is empty; field retained for Go overrides / tests).
	ThesaurusRelevantRule any
	// SynonymsFn optional override.
	SynonymsFn func(word, langCode string) []string
	// SpellFn optional override; default false.
	SpellFn func(word, langCode string) bool
	// SyllablesFn optional; default 0.
	SyllablesFn func(word, langCode string) int
}

func NewLinguServices() *LinguServices { return &LinguServices{} }

// GetSynonyms ports getSynonyms — Java returns new empty ArrayList (never null).
func (l *LinguServices) GetSynonyms(word, langCode string) []string {
	if l != nil && l.SynonymsFn != nil {
		if out := l.SynonymsFn(word, langCode); out != nil {
			return out
		}
	}
	return []string{}
}

// IsCorrectSpell ports isCorrectSpell — default false.
func (l *LinguServices) IsCorrectSpell(word, langCode string) bool {
	if l != nil && l.SpellFn != nil {
		return l.SpellFn(word, langCode)
	}
	return false
}

// GetNumberOfSyllables ports getNumberOfSyllables — default 0
// (Java doc says -1 if not found; base impl returns 0).
func (l *LinguServices) GetNumberOfSyllables(word, langCode string) int {
	if l != nil && l.SyllablesFn != nil {
		return l.SyllablesFn(word, langCode)
	}
	return 0
}

// SetThesaurusRelevantRule ports setThesaurusRelevantRule (Java base: empty body).
func (l *LinguServices) SetThesaurusRelevantRule(rule any) {
	if l != nil {
		l.ThesaurusRelevantRule = rule
	}
}
