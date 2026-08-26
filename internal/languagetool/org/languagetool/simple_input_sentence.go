package languagetool

import "hash/fnv"

// SimpleInputSentence ports org.languagetool.SimpleInputSentence — cache key for analysis.
// Java package-private; exposed in Go for ResultCache wiring.
type SimpleInputSentence struct {
	Text         string
	LanguageCode string // Language twin not fully wired; code stands in for Language.equals
}

// NewSimpleInputSentence ports SimpleInputSentence(String text, Language lang).
// text and languageCode must be non-null (Go empty string for text is allowed; empty code panics).
func NewSimpleInputSentence(text, languageCode string) SimpleInputSentence {
	// Java: Objects.requireNonNull(text); Objects.requireNonNull(lang)
	if languageCode == "" {
		panic("language required")
	}
	return SimpleInputSentence{Text: text, LanguageCode: languageCode}
}

func (s SimpleInputSentence) GetText() string { return s.Text }

// Equal ports equals (text + lang).
func (s SimpleInputSentence) Equal(o SimpleInputSentence) bool {
	return s.Text == o.Text && s.LanguageCode == o.LanguageCode
}

// Hash ports hashCode (Objects.hash(text, lang)).
func (s SimpleInputSentence) Hash() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.Text))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(s.LanguageCode))
	return h.Sum64()
}

func (s SimpleInputSentence) String() string { return s.Text }
