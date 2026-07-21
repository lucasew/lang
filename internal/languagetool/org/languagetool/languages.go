package languagetool

import (
	"fmt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"path/filepath"
	"strings"
	"sync"
)

// LanguageMeta is a lightweight registered language entry for the Languages registry.
type LanguageMeta struct {
	Name     string
	Code     string // short code with optional variant
	DictPath string // optional dynamic dict path
	// DefaultVariantCode ports Language.getDefaultLanguageVariant().
	// Empty means default is self (Java Language default). When set to another
	// code (e.g. "en-US" on "en"), this entry is not the default variant.
	DefaultVariantCode string
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

// NoopLanguageCode is the short code for Java NoopLanguage ("zz").
const NoopLanguageCode = "zz"

// NoopLanguageMeta is a stand-in for Languages.NOOP_LANGUAGE.
var NoopLanguageMeta = LanguageMeta{Name: "NoopLanguage", Code: NoopLanguageCode}

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
	if tools.JavaStringTrim(langCode) == "" {
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
	// Java empty noop list path
	return L.GetLanguageForShortCodeWithNoop(code, nil)
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

// GetLanguageForShortCodeWithNoop ports getLanguageForShortCode(langCode, noopLanguageCodes).
// Returns NoopLanguageMeta when code is listed in noopLanguageCodes and not otherwise found.
func (L *Languages) GetLanguageForShortCodeWithNoop(code string, noopLanguageCodes []string) LanguageMeta {
	if err := ValidateLanguageCodeFormat(code); err != nil {
		panic(err.Error())
	}
	if m, ok := L.findLanguage(code); ok {
		return m
	}
	for _, n := range noopLanguageCodes {
		if strings.EqualFold(n, code) {
			return NoopLanguageMeta
		}
	}
	panic(fmt.Sprintf("'%s' is not a language code known to LanguageTool. Supported language codes are: %s. See https://dev.languagetool.org/java-api for details.",
		code, strings.Join(L.GetLangCodes(), ", ")))
}

// GetLangCodes ports getLangCodes — sorted shortCodeWithCountryAndVariant list.
func (L *Languages) GetLangCodes() []string {
	L.mu.RLock()
	defer L.mu.RUnlock()
	var codes []string
	seen := map[string]struct{}{}
	for _, l := range append(append([]LanguageMeta{}, L.static...), L.dynamic...) {
		c := l.GetShortCodeWithCountryAndVariant()
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		codes = append(codes, c)
	}
	// sort
	for i := 0; i < len(codes); i++ {
		for j := i + 1; j < len(codes); j++ {
			if codes[j] < codes[i] {
				codes[i], codes[j] = codes[j], codes[i]
			}
		}
	}
	return codes
}

// HasPremiumClass ports Languages.hasPremium(className) — premium language class names.
func HasPremiumClass(className string) bool {
	// Java regex list of premium language classes
	premium := []string{
		"Portuguese", "AngolaPortuguese", "BrazilianPortuguese", "MozambiquePortuguese", "PortugalPortuguese",
		"German", "GermanyGerman", "AustrianGerman", "SwissGerman",
		"Dutch", "French", "Spanish",
		"English", "AustralianEnglish", "AmericanEnglish", "BritishEnglish", "CanadianEnglish", "NewZealandEnglish", "SouthAfricanEnglish",
	}
	prefix := "org.languagetool.language."
	if !strings.HasPrefix(className, prefix) {
		return false
	}
	simple := strings.TrimPrefix(className, prefix)
	for _, p := range premium {
		if simple == p {
			return true
		}
	}
	return false
}

// HasPremium is the Languages.hasPremium alias used by LanguagesTest.
func (L *Languages) HasPremium(className string) bool {
	return HasPremiumClass(className)
}

// IsVariant ports Language.isVariant for registry metas:
// true when code has a country/variant suffix (Java: subclass of another language).
func (l LanguageMeta) IsVariant() bool {
	return strings.Contains(l.Code, "-")
}

// IsTheDefaultVariant ports Language.isTheDefaultVariant.
func (l LanguageMeta) IsTheDefaultVariant() bool {
	if l.DefaultVariantCode == "" {
		return true // Java default: getDefaultLanguageVariant() returns this
	}
	return strings.EqualFold(l.Code, l.DefaultVariantCode)
}

// HasVariant ports Language.hasVariant: another registered language shares short code.
// Java uses class assignability; registry approx: sibling with same GetShortCode().
func (L *Languages) HasVariant(l LanguageMeta) bool {
	sc := l.GetShortCode()
	for _, o := range L.Get() {
		if strings.EqualFold(o.Code, l.Code) {
			continue
		}
		if o.GetShortCode() == sc {
			// Base language (no hyphen) has variants when a country code exists.
			// Variant languages (en-US) do not "have" further variants in Java.
			if !l.IsVariant() {
				return true
			}
		}
	}
	return false
}

// IsHiddenFromGui ports Language.isHiddenFromGui:
// hasVariant && !isVariant && !isTheDefaultVariant.
func (L *Languages) IsHiddenFromGui(l LanguageMeta) bool {
	return L.HasVariant(l) && !l.IsVariant() && !l.IsTheDefaultVariant()
}

// GetOrAddLanguageByClassName ports getOrAddLanguageByClassName for registered metas.
// When not found, panics (classpath ClassBroker not available for arbitrary classes).
func (L *Languages) GetOrAddLanguageByClassName(className string) LanguageMeta {
	// Match by simple name suffix of registered entries is not possible without class map;
	// look for Name matching last segment.
	simple := className
	if i := strings.LastIndex(className, "."); i >= 0 {
		simple = className[i+1:]
	}
	L.mu.RLock()
	for _, l := range append(append([]LanguageMeta{}, L.static...), L.dynamic...) {
		if l.Name == simple || strings.EqualFold(l.Code, simple) {
			L.mu.RUnlock()
			return l
		}
	}
	L.mu.RUnlock()
	panic(fmt.Sprintf("Class '%s' could not be found in classpath", className))
}
