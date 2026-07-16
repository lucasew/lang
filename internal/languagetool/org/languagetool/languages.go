package languagetool

import (
	"path/filepath"
	"strings"
	"sync"
)

// LanguageMeta is a lightweight registered language entry for the Languages registry.
type LanguageMeta struct {
	Name     string
	Code     string // short code with optional variant
	DictPath string // optional dynamic dict path
}

func (l LanguageMeta) GetName() string { return l.Name }

func (l LanguageMeta) GetShortCode() string {
	if i := strings.IndexByte(l.Code, '-'); i >= 0 {
		return l.Code[:i]
	}
	return l.Code
}

func (l LanguageMeta) GetShortCodeWithCountryAndVariant() string { return l.Code }

// Languages is a process-level registry (ports org.languagetool.Languages surface).
type Languages struct {
	mu      sync.RWMutex
	static  []LanguageMeta
	dynamic []LanguageMeta
}

// GlobalLanguages is the package singleton.
var GlobalLanguages = &Languages{}

// Register adds a static language definition.
func (L *Languages) Register(lang LanguageMeta) {
	L.mu.Lock()
	defer L.mu.Unlock()
	L.static = append(L.static, lang)
}

// Get returns static + dynamic languages.
func (L *Languages) Get() []LanguageMeta {
	L.mu.RLock()
	defer L.mu.RUnlock()
	out := make([]LanguageMeta, 0, len(L.static)+len(L.dynamic))
	out = append(out, L.static...)
	out = append(out, L.dynamic...)
	return out
}

// AddLanguage ports Languages.addLanguage for dynamic dict-backed languages.
// dictPath must end in .dict (Morfologik) or .dic (Hunspell).
func (L *Languages) AddLanguage(name, code, dictPath string) LanguageMeta {
	ext := filepath.Ext(dictPath)
	if ext != ".dict" && ext != ".dic" {
		panic("Please specify a dictPath that ends in '.dict' (Morfologik binary dictionary) or '.dic' (Hunspell dictionary): " + dictPath)
	}
	lang := LanguageMeta{Name: name, Code: code, DictPath: dictPath}
	L.mu.Lock()
	L.dynamic = append(L.dynamic, lang)
	L.mu.Unlock()
	return lang
}

// GetLanguageForShortCode finds by exact code or short code; panics if missing (like Java).
func (L *Languages) GetLanguageForShortCode(code string) LanguageMeta {
	L.mu.RLock()
	defer L.mu.RUnlock()
	for _, l := range append(append([]LanguageMeta{}, L.static...), L.dynamic...) {
		if l.Code == code || l.GetShortCode() == code {
			return l
		}
	}
	panic("Language not found: " + code)
}

// IsLanguageSupported is true if a language with the short code is registered.
func (L *Languages) IsLanguageSupported(code string) bool {
	L.mu.RLock()
	defer L.mu.RUnlock()
	for _, l := range append(append([]LanguageMeta{}, L.static...), L.dynamic...) {
		if l.Code == code || l.GetShortCode() == code {
			return true
		}
	}
	return false
}

// ClearDynamic removes dynamic languages (tests).
func (L *Languages) ClearDynamic() {
	L.mu.Lock()
	L.dynamic = nil
	L.mu.Unlock()
}
