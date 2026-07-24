package hunspell

// HunspellDictionary ports org.languagetool.rules.spelling.hunspell.HunspellDictionary.
type HunspellDictionary interface {
	Spell(word string) bool
	Add(word string)
	Suggest(word string) []string
	IsClosed() bool
	Close() error
}

// MapHunspellDictionary is an in-memory dictionary for tests.
type MapHunspellDictionary struct {
	words    map[string]struct{}
	suggests map[string][]string
	closed   bool
}

func NewMapHunspellDictionary(words []string) *MapHunspellDictionary {
	m := &MapHunspellDictionary{
		words:    map[string]struct{}{},
		suggests: map[string][]string{},
	}
	for _, w := range words {
		m.words[w] = struct{}{}
	}
	return m
}

func (m *MapHunspellDictionary) Spell(word string) bool {
	if m == nil || m.closed {
		return false
	}
	_, ok := m.words[word]
	return ok
}

func (m *MapHunspellDictionary) Add(word string) {
	if m == nil || m.closed {
		return
	}
	m.words[word] = struct{}{}
}

func (m *MapHunspellDictionary) Suggest(word string) []string {
	if m == nil || m.closed {
		return nil
	}
	if s, ok := m.suggests[word]; ok {
		return append([]string(nil), s...)
	}
	return nil
}

func (m *MapHunspellDictionary) SetSuggestions(word string, suggestions []string) {
	m.suggests[word] = append([]string(nil), suggestions...)
}

func (m *MapHunspellDictionary) IsClosed() bool { return m != nil && m.closed }

func (m *MapHunspellDictionary) Close() error {
	if m != nil {
		m.closed = true
	}
	return nil
}
