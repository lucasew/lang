package translation

import (
	"fmt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"strings"
)

// DataSource ports org.languagetool.rules.translation.DataSource.
type DataSource struct {
	LicenseURL string
	SourceName string
	SourceURL  string
}

func NewDataSource(licenseURL, sourceName, sourceURL string) DataSource {
	return DataSource{LicenseURL: licenseURL, SourceName: sourceName, SourceURL: sourceURL}
}

// TranslationEntry ports org.languagetool.rules.translation.TranslationEntry.
type TranslationEntry struct {
	L1        []string
	L2        []string
	ItemCount int
}

func NewTranslationEntry(l1, l2 []string, itemCount int) TranslationEntry {
	return TranslationEntry{
		L1:        append([]string(nil), l1...),
		L2:        append([]string(nil), l2...),
		ItemCount: itemCount,
	}
}

func (e TranslationEntry) GetL1() []string   { return append([]string(nil), e.L1...) }
func (e TranslationEntry) GetL2() []string   { return append([]string(nil), e.L2...) }
func (e TranslationEntry) GetItemCount() int { return e.ItemCount }

func (e TranslationEntry) String() string {
	return fmt.Sprintf("%v -> %v", e.L1, e.L2)
}

func (e TranslationEntry) Equal(o TranslationEntry) bool {
	return slicesEqual(e.L1, o.L1) && slicesEqual(e.L2, o.L2)
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TranslationData ports org.languagetool.rules.translation.TranslationData.
type TranslationData struct {
	Translations []TranslationEntry
	DataSource   DataSource
}

func NewTranslationData(entries []TranslationEntry, src DataSource) TranslationData {
	return TranslationData{
		Translations: append([]TranslationEntry(nil), entries...),
		DataSource:   src,
	}
}

func (d TranslationData) GetTranslations() []TranslationEntry {
	return append([]TranslationEntry(nil), d.Translations...)
}

// Translator ports org.languagetool.rules.translation.Translator.
type Translator interface {
	Translate(term, fromLang, toLang string) ([]TranslationEntry, error)
	GetDataSource() DataSource
	GetMessage() string
	CleanTranslationForReplace(s, prevWord string) string
	GetTranslationSuffix(s string) string
}

// MapTranslator is an in-memory Translator for tests and offline use.
type MapTranslator struct {
	// key: from|to|term (lower) → entries
	Data    map[string][]TranslationEntry
	Source  DataSource
	Message string
}

func NewMapTranslator(src DataSource) *MapTranslator {
	return &MapTranslator{Data: map[string][]TranslationEntry{}, Source: src}
}

func (m *MapTranslator) key(term, from, to string) string {
	return strings.ToLower(from) + "|" + strings.ToLower(to) + "|" + strings.ToLower(term)
}

func (m *MapTranslator) Add(term, from, to string, entry TranslationEntry) {
	k := m.key(term, from, to)
	m.Data[k] = append(m.Data[k], entry)
}

func (m *MapTranslator) Translate(term, fromLang, toLang string) ([]TranslationEntry, error) {
	return append([]TranslationEntry(nil), m.Data[m.key(term, fromLang, toLang)]...), nil
}

func (m *MapTranslator) GetDataSource() DataSource { return m.Source }
func (m *MapTranslator) GetMessage() string {
	if m.Message != "" {
		return m.Message
	}
	return "Translation suggestion"
}

func (m *MapTranslator) CleanTranslationForReplace(s, _ string) string {
	// strip trailing parenthetical notes: "foo (bar)" → "foo"
	s = tools.JavaStringTrim(s)
	if i := strings.Index(s, " ("); i > 0 && strings.HasSuffix(s, ")") {
		return tools.JavaStringTrim(s[:i])
	}
	return s
}

func (m *MapTranslator) GetTranslationSuffix(s string) string {
	s = tools.JavaStringTrim(s)
	if i := strings.Index(s, " ("); i > 0 && strings.HasSuffix(s, ")") {
		return tools.JavaStringTrim(s[i:])
	}
	return ""
}

var _ Translator = (*MapTranslator)(nil)
