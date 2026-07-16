package languagetool

import (
	"fmt"
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
	// Prefer full code match for private-use tags; short code is still the language part.
	code := l.Code
	if i := strings.Index(code, "-x-"); i >= 0 {
		// e.g. de-DE-x-simple-language → language short code before first hyphen of primary
		if j := strings.IndexByte(code, '-'); j >= 0 {
			return code[:j]
		}
	}
	if i := strings.IndexByte(code, '-'); i >= 0 {
		return code[:i]
	}
	return code
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

// Get returns a copy of static + dynamic languages (excluding demo "xx" / noop "zz").
func (L *Languages) Get() []LanguageMeta {
	L.mu.RLock()
	defer L.mu.RUnlock()
	out := make([]LanguageMeta, 0, len(L.static)+len(L.dynamic))
	for _, l := range L.static {
		if sc := l.GetShortCode(); sc == "xx" || sc == "zz" {
			continue
		}
		// also skip if full code is xx/zz
		if strings.EqualFold(l.Code, "xx") || strings.EqualFold(l.Code, "zz") {
			continue
		}
		out = append(out, l)
	}
	out = append(out, L.dynamic...)
	return out
}

// GetWithDemoLanguage returns static + dynamic including demo (unfiltered copy).
func (L *Languages) GetWithDemoLanguage() []LanguageMeta {
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

// ValidateLanguageCodeFormat checks Java-compatible lang code structure.
// Returns an error for structurally invalid codes (e.g. too many hyphen parts).
func ValidateLanguageCodeFormat(langCode string) error {
	if strings.TrimSpace(langCode) == "" {
		return fmt.Errorf("langCode cannot be empty")
	}
	if strings.Contains(langCode, "-x-") {
		return nil
	}
	if strings.Contains(langCode, "-") {
		parts := strings.Split(langCode, "-")
		// Java String.split discards trailing empties: "de-" → ["de"] → invalid
		// Keep Go Split behavior but treat empty segments as invalid.
		if len(parts) != 2 && len(parts) != 3 {
			return fmt.Errorf("'%s' isn't a valid language code", langCode)
		}
		for _, p := range parts {
			if p == "" {
				return fmt.Errorf("'%s' isn't a valid language code", langCode)
			}
		}
	}
	return nil
}

// GetLanguageForShortCode finds by exact code or short code (case-insensitive); panics if missing.
func (L *Languages) GetLanguageForShortCode(code string) LanguageMeta {
	if err := ValidateLanguageCodeFormat(code); err != nil {
		panic(err.Error())
	}
	if m, ok := L.findLanguage(code); ok {
		return m
	}
	panic(fmt.Sprintf("'%s' is not a language code known to LanguageTool.", code))
}

// IsLanguageSupported is true if a language with the short code is registered.
// Panics for invalid code format (Java IllegalArgumentException).
func (L *Languages) IsLanguageSupported(code string) bool {
	if err := ValidateLanguageCodeFormat(code); err != nil {
		panic(err.Error())
	}
	_, ok := L.findLanguage(code)
	return ok
}

func (L *Languages) findLanguage(code string) (LanguageMeta, bool) {
	L.mu.RLock()
	defer L.mu.RUnlock()
	want := strings.ToLower(code)
	for _, l := range append(append([]LanguageMeta{}, L.static...), L.dynamic...) {
		if strings.EqualFold(l.Code, code) || strings.EqualFold(l.GetShortCode(), code) {
			return l, true
		}
		// also match en-us style against registered en-US
		if strings.ToLower(l.Code) == want {
			return l, true
		}
	}
	return LanguageMeta{}, false
}

// GetLanguageForName finds by display name (case-sensitive, like Java).
func (L *Languages) GetLanguageForName(name string) (LanguageMeta, bool) {
	L.mu.RLock()
	defer L.mu.RUnlock()
	for _, l := range append(append([]LanguageMeta{}, L.static...), L.dynamic...) {
		if l.Name == name {
			return l, true
		}
	}
	return LanguageMeta{}, false
}

// ClearDynamic removes dynamic languages (tests).
func (L *Languages) ClearDynamic() {
	L.mu.Lock()
	L.dynamic = nil
	L.mu.Unlock()
}
