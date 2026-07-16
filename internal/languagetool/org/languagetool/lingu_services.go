package languagetool

// LinguServices ports org.languagetool.LinguServices — LO extension hooks.
// Default implementations are no-ops / empty (same as Java base class).
type LinguServices struct {
	// ThesaurusRelevantRule stores the last rule set via SetThesaurusRelevantRule.
	ThesaurusRelevantRule any
	// SynonymsFn optional override.
	SynonymsFn func(word, langCode string) []string
	// SpellFn optional override; default false.
	SpellFn func(word, langCode string) bool
	// SyllablesFn optional; default 0.
	SyllablesFn func(word, langCode string) int
}

func NewLinguServices() *LinguServices { return &LinguServices{} }

func (l *LinguServices) GetSynonyms(word, langCode string) []string {
	if l != nil && l.SynonymsFn != nil {
		return l.SynonymsFn(word, langCode)
	}
	return nil
}

func (l *LinguServices) IsCorrectSpell(word, langCode string) bool {
	if l != nil && l.SpellFn != nil {
		return l.SpellFn(word, langCode)
	}
	return false
}

func (l *LinguServices) GetNumberOfSyllables(word, langCode string) int {
	if l != nil && l.SyllablesFn != nil {
		return l.SyllablesFn(word, langCode)
	}
	return 0
}

func (l *LinguServices) SetThesaurusRelevantRule(rule any) {
	if l != nil {
		l.ThesaurusRelevantRule = rule
	}
}
