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
	Code     string // getShortCode() — base language code (en, de, …) or full private-use tag
	DictPath string // optional dynamic dict path
	// Countries ports Language.getCountries() (ISO region codes; may be empty or multi).
	Countries []string
	// Variant ports Language.getVariant() (e.g. "valencia", "balear"); empty if none.
	Variant string
	// DefaultVariantCode ports Language.getDefaultLanguageVariant().
	// Empty means default is self (Java Language default). When set to another
	// code (e.g. "en-US" on "en"), this entry is not the default variant.
	DefaultVariantCode string
}

func (l LanguageMeta) GetName() string { return l.Name }

// GetShortCode ports Language.getShortCode().
// For private-use tags (de-DE-x-simple-language) Java returns the full code as shortCode.
func (l LanguageMeta) GetShortCode() string {
	code := l.Code
	if strings.Contains(code, "-x-") {
		// SimpleGerman: getShortCode() returns "de-DE-x-simple-language"
		return code
	}
	// When Code stores long form by mistake (legacy), strip region.
	if i := strings.IndexByte(code, '-'); i >= 0 {
		return code[:i]
	}
	return code
}

// GetCountries ports Language.getCountries().
func (l LanguageMeta) GetCountries() []string {
	return append([]string(nil), l.Countries...)
}

// GetShortCodeWithCountryAndVariant ports Language.getShortCodeWithCountryAndVariant().
// Java: only append country when countries.length == 1 and short code has no "-x-".
func (l LanguageMeta) GetShortCodeWithCountryAndVariant() string {
	name := l.GetShortCode()
	if len(l.Countries) == 1 && !strings.Contains(name, "-x-") {
		name += "-" + l.Countries[0]
		if l.Variant != "" {
			name += "-" + l.Variant
		}
	}
	return name
}

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

// builtInLanguages ports language classes from META-INF language-module.properties
// (name, getShortCode, getCountries, variant). longCode is derived via
// GetShortCodeWithCountryAndVariant (single-country only).
var builtInLanguages = []LanguageMeta{
	// English module
	{Name: "English", Code: "en"},
	{Name: "English (US)", Code: "en", Countries: []string{"US"}},
	{Name: "English (GB)", Code: "en", Countries: []string{"GB"}},
	{Name: "English (Australian)", Code: "en", Countries: []string{"AU"}},
	{Name: "English (Canadian)", Code: "en", Countries: []string{"CA"}},
	{Name: "English (New Zealand)", Code: "en", Countries: []string{"NZ"}},
	{Name: "English (South African)", Code: "en", Countries: []string{"ZA"}},
	// German module
	{Name: "German", Code: "de", Countries: []string{"LU", "LI", "BE"}},
	{Name: "German (Germany)", Code: "de", Countries: []string{"DE"}},
	{Name: "German (Austria)", Code: "de", Countries: []string{"AT"}},
	{Name: "German (Swiss)", Code: "de", Countries: []string{"CH"}},
	// French
	{Name: "French", Code: "fr", Countries: []string{"FR", "LU", "MC", "CM", "CI", "HT", "ML", "SN", "CD", "MA", "RE"}},
	{Name: "French (Canada)", Code: "fr", Countries: []string{"CA"}},
	{Name: "French (Switzerland)", Code: "fr", Countries: []string{"CH"}},
	{Name: "French (Belgium)", Code: "fr", Countries: []string{"BE"}},
	// Spanish
	{Name: "Spanish", Code: "es", Countries: []string{"ES", "MX", "GT", "CR", "PA", "DO", "VE", "PE", "AR", "EC", "CL", "UY", "PY", "BO", "SV", "HN", "NI", "PR", "US", "CU"}},
	{Name: "Spanish (voseo)", Code: "es", Countries: []string{"AR", "PA", "UY", "CR"}},
	// Portuguese
	{Name: "Portuguese", Code: "pt", Countries: []string{"CV", "GW", "MO", "ST", "TL"}},
	{Name: "Portuguese (Portugal)", Code: "pt", Countries: []string{"PT"}},
	{Name: "Portuguese (Brazil)", Code: "pt", Countries: []string{"BR"}},
	{Name: "Portuguese (Angola preAO)", Code: "pt", Countries: []string{"AO"}},
	{Name: "Portuguese (Moçambique preAO)", Code: "pt", Countries: []string{"MZ"}},
	// Dutch
	{Name: "Dutch", Code: "nl", Countries: []string{"NL", "BE"}},
	{Name: "Dutch (Belgium)", Code: "nl", Countries: []string{"BE"}},
	// Catalan
	{Name: "Catalan", Code: "ca", Countries: []string{"ES"}},
	{Name: "Catalan (Valencian)", Code: "ca", Countries: []string{"ES"}, Variant: "valencia"},
	{Name: "Catalan (Balearic)", Code: "ca", Countries: []string{"ES"}, Variant: "balear"},
	// Serbian
	{Name: "Serbian", Code: "sr"},
	{Name: "Serbian (Serbia)", Code: "sr", Countries: []string{"RS"}},
	{Name: "Serbian (Bosnia and Herzegovina)", Code: "sr", Countries: []string{"BA"}},
	{Name: "Serbian (Croatia)", Code: "sr", Countries: []string{"HR"}},
	{Name: "Serbian (Montenegro)", Code: "sr", Countries: []string{"ME"}},
	// Single-country / simple modules
	{Name: "Arabic", Code: "ar", Countries: []string{"SA", "DZ", "BH", "EG", "IQ", "JO", "KW", "LB", "LY", "MA", "OM", "QA", "SD", "SY", "TN", "AE", "YE"}},
	{Name: "Asturian", Code: "ast", Countries: []string{"ES"}},
	{Name: "Belarusian", Code: "be", Countries: []string{"BY"}},
	{Name: "Breton", Code: "br", Countries: []string{"FR"}},
	{Name: "Crimean Tatar", Code: "crh", Countries: []string{"UA"}},
	{Name: "Danish", Code: "da", Countries: []string{"DK"}},
	{Name: "Greek", Code: "el", Countries: []string{"GR"}},
	{Name: "Esperanto", Code: "eo"},
	{Name: "Persian", Code: "fa", Countries: []string{"IR", "AF"}},
	{Name: "Irish", Code: "ga", Countries: []string{"IE"}},
	{Name: "Galician", Code: "gl", Countries: []string{"ES"}},
	{Name: "Icelandic", Code: "is", Countries: []string{"IS"}},
	{Name: "Italian", Code: "it", Countries: []string{"IT", "CH"}},
	{Name: "Japanese", Code: "ja", Countries: []string{"JP"}},
	{Name: "Khmer", Code: "km", Countries: []string{"KH"}},
	{Name: "Lithuanian", Code: "lt", Countries: []string{"LT"}},
	{Name: "Malayalam", Code: "ml", Countries: []string{"IN"}},
	{Name: "Polish", Code: "pl", Countries: []string{"PL"}},
	{Name: "Romanian", Code: "ro", Countries: []string{"RO"}},
	{Name: "Russian", Code: "ru", Countries: []string{"RU"}},
	{Name: "Slovak", Code: "sk", Countries: []string{"SK"}},
	{Name: "Slovenian", Code: "sl", Countries: []string{"SI"}},
	{Name: "Swedish", Code: "sv", Countries: []string{"SE", "FI"}},
	{Name: "Tamil", Code: "ta", Countries: []string{"IN"}},
	{Name: "Tagalog", Code: "tl", Countries: []string{"PH"}},
	{Name: "Ukrainian", Code: "uk", Countries: []string{"UA"}},
	{Name: "Chinese", Code: "zh", Countries: []string{"CN"}},
	// Simple German private-use tag (getShortCode is full tag)
	{Name: "Simple German", Code: "de-DE-x-simple-language"},
}

var builtInLangsOnce sync.Once

// EnsureBuiltInLanguagesRegistered ports Java Languages class-init:
// language modules registered before any detect / canLanguageBeDetected call.
// Safe to call repeatedly (sync.Once).
func EnsureBuiltInLanguagesRegistered() {
	builtInLangsOnce.Do(func() {
		for _, m := range builtInLanguages {
			GlobalLanguages.Register(m)
		}
		// Do not register NoopLanguage "zz" here — Java Languages.get() excludes zz/xx;
		// canLanguageBeDetected("zz") is true only via additionalLanguageCodes (noop list).
	})
}

// GetLongCodeToLangMapping ports Languages.getLongCodeToLangMapping.
// Maps "fr-FR" → French (short fr, countries[0]=FR) for LibreOffice 7.4.
func (L *Languages) GetLongCodeToLangMapping() map[string]LanguageMeta {
	L.mu.RLock()
	defer L.mu.RUnlock()
	m := map[string]LanguageMeta{}
	all := append(append([]LanguageMeta{}, L.static...), L.dynamic...)
	for _, lang := range all {
		sc := lang.GetShortCode()
		if sc == "xx" || sc == "zz" {
			continue
		}
		cs := lang.Countries
		if len(cs) > 0 && cs[0] != "" {
			m[sc+"-"+cs[0]] = lang
		}
	}
	return m
}

// normalizeLanguageMeta accepts legacy Code="en-US" (full long code, empty Countries)
// and expands it to Code=short + Countries like Java Language objects.
func normalizeLanguageMeta(lang LanguageMeta) LanguageMeta {
	if len(lang.Countries) > 0 {
		return lang
	}
	code := lang.Code
	if code == "" || strings.Contains(code, "-x-") {
		return lang
	}
	parts := strings.Split(code, "-")
	if len(parts) == 2 && len(parts[0]) >= 2 && len(parts[1]) == 2 {
		lang.Code = parts[0]
		lang.Countries = []string{strings.ToUpper(parts[1])}
	} else if len(parts) == 3 && len(parts[1]) == 2 {
		// ca-ES-valencia
		lang.Code = parts[0]
		lang.Countries = []string{strings.ToUpper(parts[1])}
		lang.Variant = parts[2]
	}
	return lang
}

// Register adds a static language definition.
func (L *Languages) Register(lang LanguageMeta) {
	L.mu.Lock()
	defer L.mu.Unlock()
	lang = normalizeLanguageMeta(lang)
	// Idempotent on full identity (short+country+variant), not bare short code —
	// Java registers AmericanEnglish, BritishEnglish, … as distinct Language objects.
	want := lang.GetShortCodeWithCountryAndVariant()
	for _, existing := range L.static {
		if strings.EqualFold(existing.GetShortCodeWithCountryAndVariant(), want) {
			return
		}
	}
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
	all := append(append([]LanguageMeta{}, L.static...), L.dynamic...)
	// Prefer exact long-code match (en-US → AmericanEnglish)
	for _, l := range all {
		if strings.EqualFold(l.GetShortCodeWithCountryAndVariant(), code) {
			return l, true
		}
	}
	// Bare short code (en → first English-family match, typically base English)
	for _, l := range all {
		if strings.EqualFold(l.GetShortCode(), code) {
			return l, true
		}
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
// true when shortCodeWithCountryAndVariant differs from shortCode (country/variant present).
func (l LanguageMeta) IsVariant() bool {
	return l.GetShortCodeWithCountryAndVariant() != l.GetShortCode()
}

// IsTheDefaultVariant ports Language.isTheDefaultVariant.
func (l LanguageMeta) IsTheDefaultVariant() bool {
	if l.DefaultVariantCode == "" {
		return true // Java default: getDefaultLanguageVariant() returns this
	}
	return strings.EqualFold(l.GetShortCodeWithCountryAndVariant(), l.DefaultVariantCode)
}

// HasVariant ports Language.hasVariant: another registered language shares short code.
// Java uses class assignability; registry: sibling with same GetShortCode().
func (L *Languages) HasVariant(l LanguageMeta) bool {
	sc := l.GetShortCode()
	selfLong := l.GetShortCodeWithCountryAndVariant()
	for _, o := range L.Get() {
		if strings.EqualFold(o.GetShortCodeWithCountryAndVariant(), selfLong) {
			continue
		}
		if o.GetShortCode() == sc {
			// Base language (no country) has variants when a country code sibling exists.
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
