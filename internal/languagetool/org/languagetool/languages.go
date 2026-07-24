package languagetool

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// LanguageMeta is a lightweight registered language entry for the Languages registry.
// Ports the surface of org.languagetool.Language used by Languages / ApiV2.
type LanguageMeta struct {
	Name     string
	Code     string // getShortCode() — base language code (en, de, …) or full private-use tag
	DictPath string // optional dynamic dict path
	// Countries ports Language.getCountries() (ISO region codes; may include "" like Java).
	Countries []string
	// Variant ports Language.getVariant() (e.g. "valencia", "balear"); empty if none.
	Variant string
	// DefaultVariantCode ports getDefaultLanguageVariant().getShortCodeWithCountryAndVariant().
	// Empty means Java null default (isTheDefaultVariant → false).
	DefaultVariantCode string
	// ClassName is the Java simple class name (e.g. "AmericanEnglish") for isVariant/hasVariant.
	ClassName string
	// SuperClass is the immediate Java superclass simple name when it is an LT Language
	// subclass (e.g. "English"); empty when extends Language / LanguageWithModel directly.
	SuperClass string
	// ForceIsVariant ports overrides like SimpleGerman.isVariant() → true.
	ForceIsVariant bool
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
var NoopLanguageMeta = LanguageMeta{Name: "NoopLanguage", Code: NoopLanguageCode, ClassName: "NoopLanguage"}

// builtInLanguages ports language classes from META-INF language-module.properties
// (name, getShortCode, getCountries, variant, class hierarchy). longCode is derived via
// GetShortCodeWithCountryAndVariant (single-country only).
//
// Country arrays match Java including empty-string entries (affects getLongCodeToLangMapping:
// first country must be non-empty for a mapping key).
var builtInLanguages = []LanguageMeta{
	// English module
	{Name: "English", Code: "en", ClassName: "English", DefaultVariantCode: "en-US"},
	{Name: "English (US)", Code: "en", Countries: []string{"US"}, ClassName: "AmericanEnglish", SuperClass: "English"},
	{Name: "English (GB)", Code: "en", Countries: []string{"GB"}, ClassName: "BritishEnglish", SuperClass: "English"},
	{Name: "English (Australian)", Code: "en", Countries: []string{"AU"}, ClassName: "AustralianEnglish", SuperClass: "English"},
	{Name: "English (Canadian)", Code: "en", Countries: []string{"CA"}, ClassName: "CanadianEnglish", SuperClass: "English"},
	{Name: "English (New Zealand)", Code: "en", Countries: []string{"NZ"}, ClassName: "NewZealandEnglish", SuperClass: "English"},
	{Name: "English (South African)", Code: "en", Countries: []string{"ZA"}, ClassName: "SouthAfricanEnglish", SuperClass: "English"},
	// German module
	{Name: "German", Code: "de", Countries: []string{"LU", "LI", "BE"}, ClassName: "German", DefaultVariantCode: "de-DE"},
	{Name: "German (Germany)", Code: "de", Countries: []string{"DE"}, ClassName: "GermanyGerman", SuperClass: "German"},
	{Name: "German (Austria)", Code: "de", Countries: []string{"AT"}, ClassName: "AustrianGerman", SuperClass: "German"},
	{Name: "German (Swiss)", Code: "de", Countries: []string{"CH"}, ClassName: "SwissGerman", SuperClass: "German"},
	// French — first country FR (LibreOffice fr-FR mapping); empty string in list as Java
	{Name: "French", Code: "fr", Countries: []string{"FR", "", "LU", "MC", "CM", "CI", "HT", "ML", "SN", "CD", "MA", "RE"}, ClassName: "French", DefaultVariantCode: "fr"},
	{Name: "French (Canada)", Code: "fr", Countries: []string{"CA"}, ClassName: "CanadianFrench", SuperClass: "French"},
	{Name: "French (Switzerland)", Code: "fr", Countries: []string{"CH"}, ClassName: "SwissFrench", SuperClass: "French"},
	{Name: "French (Belgium)", Code: "fr", Countries: []string{"BE"}, ClassName: "BelgianFrench", SuperClass: "French"},
	// Spanish — multi-country includes "" as Java; SpanishVoseo only "AR" (commented PA/UY/CR not active)
	{Name: "Spanish", Code: "es", Countries: []string{"ES", "", "MX", "GT", "CR", "PA", "DO", "VE", "PE", "AR", "EC", "CL", "UY", "PY", "BO", "SV", "HN", "NI", "PR", "US", "CU"}, ClassName: "Spanish", DefaultVariantCode: "es"},
	{Name: "Spanish (voseo)", Code: "es", Countries: []string{"AR"}, ClassName: "SpanishVoseo", SuperClass: "Spanish"},
	// Portuguese — first country "" so getLongCodeToLangMapping skips bare Portuguese
	{Name: "Portuguese", Code: "pt", Countries: []string{"", "CV", "GW", "MO", "ST", "TL"}, ClassName: "Portuguese", DefaultVariantCode: "pt-PT"},
	{Name: "Portuguese (Portugal)", Code: "pt", Countries: []string{"PT"}, ClassName: "PortugalPortuguese", SuperClass: "Portuguese"},
	{Name: "Portuguese (Brazil)", Code: "pt", Countries: []string{"BR"}, ClassName: "BrazilianPortuguese", SuperClass: "Portuguese"},
	{Name: "Portuguese (Angola preAO)", Code: "pt", Countries: []string{"AO"}, ClassName: "AngolaPortuguese", SuperClass: "Portuguese"},
	{Name: "Portuguese (Moçambique preAO)", Code: "pt", Countries: []string{"MZ"}, ClassName: "MozambiquePortuguese", SuperClass: "Portuguese"},
	// Dutch
	{Name: "Dutch", Code: "nl", Countries: []string{"NL", "BE"}, ClassName: "Dutch", DefaultVariantCode: "nl"},
	{Name: "Dutch (Belgium)", Code: "nl", Countries: []string{"BE"}, ClassName: "BelgianDutch", SuperClass: "Dutch"},
	// Catalan — base has countries ES so longCode is ca-ES; default is self
	{Name: "Catalan", Code: "ca", Countries: []string{"ES"}, ClassName: "Catalan", DefaultVariantCode: "ca-ES"},
	{Name: "Catalan (Valencian)", Code: "ca", Countries: []string{"ES"}, Variant: "valencia", ClassName: "ValencianCatalan", SuperClass: "Catalan"},
	{Name: "Catalan (Balearic)", Code: "ca", Countries: []string{"ES"}, Variant: "balear", ClassName: "BalearicCatalan", SuperClass: "Catalan"},
	// Serbian (JekavianSerbian not in META-INF; hierarchy still Bosnian→Jekavian→Serbian for assignability)
	{Name: "Serbian", Code: "sr", ClassName: "Serbian"},
	{Name: "Serbian (Serbia)", Code: "sr", Countries: []string{"RS"}, ClassName: "SerbianSerbian", SuperClass: "Serbian"},
	{Name: "Serbian (Bosnia and Herzegovina)", Code: "sr", Countries: []string{"BA"}, ClassName: "BosnianSerbian", SuperClass: "JekavianSerbian"},
	{Name: "Serbian (Croatia)", Code: "sr", Countries: []string{"HR"}, ClassName: "CroatianSerbian", SuperClass: "JekavianSerbian"},
	{Name: "Serbian (Montenegro)", Code: "sr", Countries: []string{"ME"}, ClassName: "MontenegrinSerbian", SuperClass: "JekavianSerbian"},
	// Single / multi modules
	{Name: "Arabic", Code: "ar", Countries: []string{"", "SA", "DZ", "BH", "EG", "IQ", "JO", "KW", "LB", "LY", "MA", "OM", "QA", "SD", "SY", "TN", "AE", "YE"}, ClassName: "Arabic"},
	{Name: "Asturian", Code: "ast", Countries: []string{"ES"}, ClassName: "Asturian"},
	{Name: "Belarusian", Code: "be", Countries: []string{"BY"}, ClassName: "Belarusian"},
	{Name: "Breton", Code: "br", Countries: []string{"FR"}, ClassName: "Breton"},
	{Name: "Crimean Tatar", Code: "crh", Countries: []string{"UA"}, ClassName: "CrimeanTatar"},
	{Name: "Danish", Code: "da", Countries: []string{"DK"}, ClassName: "Danish"},
	{Name: "Greek", Code: "el", Countries: []string{"GR"}, ClassName: "Greek"},
	{Name: "Esperanto", Code: "eo", ClassName: "Esperanto"},
	{Name: "Persian", Code: "fa", Countries: []string{"IR", "AF"}, ClassName: "Persian"},
	{Name: "Irish", Code: "ga", Countries: []string{"IE"}, ClassName: "Irish"},
	{Name: "Galician", Code: "gl", Countries: []string{"ES"}, ClassName: "Galician"},
	{Name: "Icelandic", Code: "is", Countries: []string{"IS"}, ClassName: "Icelandic"},
	{Name: "Italian", Code: "it", Countries: []string{"IT", "CH"}, ClassName: "Italian"},
	{Name: "Japanese", Code: "ja", Countries: []string{"JP"}, ClassName: "Japanese"},
	{Name: "Khmer", Code: "km", Countries: []string{"KH"}, ClassName: "Khmer"},
	{Name: "Lithuanian", Code: "lt", Countries: []string{"LT"}, ClassName: "Lithuanian"},
	{Name: "Malayalam", Code: "ml", Countries: []string{"IN"}, ClassName: "Malayalam"},
	{Name: "Polish", Code: "pl", Countries: []string{"PL"}, ClassName: "Polish"},
	{Name: "Romanian", Code: "ro", Countries: []string{"RO"}, ClassName: "Romanian"},
	{Name: "Russian", Code: "ru", Countries: []string{"RU"}, ClassName: "Russian"},
	{Name: "Slovak", Code: "sk", Countries: []string{"SK"}, ClassName: "Slovak"},
	{Name: "Slovenian", Code: "sl", Countries: []string{"SI"}, ClassName: "Slovenian"},
	{Name: "Swedish", Code: "sv", Countries: []string{"SE", "FI"}, ClassName: "Swedish"},
	{Name: "Tamil", Code: "ta", Countries: []string{"IN"}, ClassName: "Tamil"},
	{Name: "Tagalog", Code: "tl", Countries: []string{"PH"}, ClassName: "Tagalog"},
	{Name: "Ukrainian", Code: "uk", Countries: []string{"UA"}, ClassName: "Ukrainian"},
	{Name: "Chinese", Code: "zh", Countries: []string{"CN"}, ClassName: "Chinese"},
	// Simple German: private-use shortCode; overrides isVariant() → true; extends GermanyGerman
	{Name: "Simple German", Code: "de-DE-x-simple-language", Countries: []string{"DE"}, ClassName: "SimpleGerman", SuperClass: "GermanyGerman", ForceIsVariant: true},
}

// jekavianSuper is the non-META-INF intermediate class JekavianSerbian extends Serbian.
// Used only for Class.isAssignableFrom simulation of Serbian variants.
const jekavianSuper = "JekavianSerbian"

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
// Java: only when countries.length > 0 && countries[0] is non-empty.
func (L *Languages) GetLongCodeToLangMapping() map[string]LanguageMeta {
	L.mu.RLock()
	defer L.mu.RUnlock()
	m := map[string]LanguageMeta{}
	for _, lang := range L.allLocked() {
		sc := lang.GetShortCode()
		if sc == "xx" || sc == "zz" {
			continue
		}
		// Languages.get() excludes xx/zz; mapping iterates Languages.get()
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
	// Idempotent on ClassName (preferred) or longCode+name for ad-hoc test entries.
	// Java registers AmericanEnglish, BritishEnglish, … as distinct Language objects.
	want := lang.GetShortCodeWithCountryAndVariant()
	for _, existing := range L.static {
		if lang.ClassName != "" && existing.ClassName == lang.ClassName {
			return
		}
		if lang.ClassName == "" && existing.ClassName == "" &&
			strings.EqualFold(existing.GetShortCodeWithCountryAndVariant(), want) &&
			existing.Name == lang.Name {
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
// Ports Languages.getLanguageForShortCode(langCode) including long-code mapping fallback.
func (L *Languages) GetLanguageForShortCode(code string) LanguageMeta {
	return L.GetLanguageForShortCodeWithNoop(code, nil)
}

// IsLanguageSupported is true if a language with the short code is registered.
// Ports Languages.isLanguageSupported — uses getLanguageForShortCodeOrNull only
// (no LibreOffice long-code mapping fallback).
// Panics for invalid code format (Java IllegalArgumentException).
func (L *Languages) IsLanguageSupported(code string) bool {
	if err := ValidateLanguageCodeFormat(code); err != nil {
		panic(err.Error())
	}
	_, ok := L.findLanguageOrNull(code)
	return ok
}

// allLocked returns static+dynamic without locking (caller holds lock).
func (L *Languages) allLocked() []LanguageMeta {
	return append(append([]LanguageMeta{}, L.static...), L.dynamic...)
}

// findLanguageOrNull ports Languages.getLanguageForShortCodeOrNull.
func (L *Languages) findLanguageOrNull(code string) (LanguageMeta, bool) {
	L.mu.RLock()
	defer L.mu.RUnlock()
	all := L.allLocked()
	if strings.Contains(code, "-x-") {
		// e.g. "de-DE-x-simple-language"
		for _, element := range all {
			if strings.EqualFold(element.GetShortCode(), code) {
				return element, true
			}
		}
		return LanguageMeta{}, false
	}
	if strings.Contains(code, "-") {
		parts := strings.Split(code, "-")
		if len(parts) == 2 { // e.g. en-US
			// Java: first countries.length==1 + country match (registry order).
			// Catalan is registered before ValencianCatalan so ca-ES → Catalan.
			for _, element := range all {
				if strings.EqualFold(parts[0], element.GetShortCode()) &&
					len(element.Countries) == 1 &&
					strings.EqualFold(parts[1], element.Countries[0]) {
					return element, true
				}
			}
			return LanguageMeta{}, false
		}
		if len(parts) == 3 { // e.g. ca-ES-valencia
			for _, element := range all {
				if strings.EqualFold(parts[0], element.GetShortCode()) &&
					len(element.Countries) == 1 &&
					strings.EqualFold(parts[1], element.Countries[0]) &&
					strings.EqualFold(parts[2], element.Variant) {
					return element, true
				}
			}
			return LanguageMeta{}, false
		}
		// invalid length — ValidateLanguageCodeFormat should have caught this
		return LanguageMeta{}, false
	}
	// bare short code — first match (Java TODO: should return DefaultLanguageVariant)
	for _, element := range all {
		if strings.EqualFold(code, "global") {
			return element, true
		}
		if strings.EqualFold(code, element.GetShortCode()) {
			return element, true
		}
	}
	return LanguageMeta{}, false
}

// GetLanguageForName finds by display name (case-sensitive, like Java).
func (L *Languages) GetLanguageForName(name string) (LanguageMeta, bool) {
	L.mu.RLock()
	defer L.mu.RUnlock()
	for _, l := range L.allLocked() {
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
// Applies getLongCodeToLangMapping after OrNull (LibreOffice fr-FR).
func (L *Languages) GetLanguageForShortCodeWithNoop(code string, noopLanguageCodes []string) LanguageMeta {
	if err := ValidateLanguageCodeFormat(code); err != nil {
		panic(err.Error())
	}
	if m, ok := L.findLanguageOrNull(code); ok {
		return m
	}
	// LibreOffice long-code mapping (Java)
	if m, ok := L.GetLongCodeToLangMapping()[code]; ok {
		return m
	}
	// case-insensitive map key lookup (Java HashMap is case-sensitive; keep exact)
	for _, n := range noopLanguageCodes {
		if n == code {
			return NoopLanguageMeta
		}
	}
	panic(fmt.Sprintf("'%s' is not a language code known to LanguageTool. Supported language codes are: %s. See https://dev.languagetool.org/java-api for details.",
		code, strings.Join(L.GetLangCodes(), ", ")))
}

// GetLangCodes ports getLangCodes — sorted shortCodeWithCountryAndVariant list
// plus long-code mapping extras not already listed.
func (L *Languages) GetLangCodes() []string {
	var codes []string
	seen := map[string]struct{}{}
	for _, l := range L.GetWithDemoLanguage() {
		c := l.GetShortCodeWithCountryAndVariant()
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		codes = append(codes, c)
	}
	for longCode := range L.GetLongCodeToLangMapping() {
		if _, ok := seen[longCode]; ok {
			continue
		}
		seen[longCode] = struct{}{}
		codes = append(codes, longCode)
	}
	sort.Strings(codes)
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

// classAncestry returns ClassName and all SuperClass names (including non-registered intermediates).
func (l LanguageMeta) classAncestry() []string {
	if l.ClassName == "" {
		return nil
	}
	out := []string{l.ClassName}
	// Walk SuperClass chain via known built-in + intermediate (JekavianSerbian → Serbian)
	super := l.SuperClass
	seen := map[string]struct{}{l.ClassName: {}}
	// Map of non-registered intermediates → their super
	intermediates := map[string]string{
		jekavianSuper: "Serbian",
		// LanguageWithModel / Language not registered
	}
	// Index built-in by ClassName for chain walk
	byClass := map[string]string{}
	for _, m := range builtInLanguages {
		if m.ClassName != "" {
			byClass[m.ClassName] = m.SuperClass
		}
	}
	for super != "" {
		if _, ok := seen[super]; ok {
			break
		}
		seen[super] = struct{}{}
		out = append(out, super)
		if next, ok := byClass[super]; ok {
			super = next
			continue
		}
		if next, ok := intermediates[super]; ok {
			super = next
			continue
		}
		break
	}
	return out
}

// isAssignableFrom ports Class.isAssignableFrom for registry metas:
// other.isAssignableFrom(this) ⇔ other is this or a superclass of this.
func (other LanguageMeta) isAssignableFrom(this LanguageMeta) bool {
	if other.ClassName == "" || this.ClassName == "" {
		return false
	}
	if other.ClassName == this.ClassName {
		return true
	}
	for _, a := range this.classAncestry() {
		if a == other.ClassName {
			return true
		}
	}
	return false
}

// IsVariant ports Language.isVariant:
// true when another registered language's class is a superclass of this one
// (or ForceIsVariant, e.g. SimpleGerman).
// With ClassName: SuperClass set ⇒ subclass of a Language (variant); or ForceIsVariant.
// Fallback without ClassName: longCode differs from shortCode (country/variant present).
func (l LanguageMeta) IsVariant() bool {
	return GlobalLanguages.languageIsVariant(l)
}

// languageIsVariant implements isVariant against this registry (Java Class hierarchy).
func (L *Languages) languageIsVariant(l LanguageMeta) bool {
	if l.ForceIsVariant {
		return true
	}
	if l.ClassName == "" {
		// Ad-hoc test registrations without class hierarchy.
		return l.GetShortCodeWithCountryAndVariant() != l.GetShortCode()
	}
	// SuperClass set means extends a Language subclass (or intermediate) → isVariant.
	// Matches Class.isAssignableFrom against registered supers; SuperClass alone is enough
	// when the parent is a known Language module class.
	if l.SuperClass != "" {
		return true
	}
	selfLong := l.GetShortCodeWithCountryAndVariant()
	for _, language := range L.Get() {
		if language.ClassName == l.ClassName &&
			language.GetShortCodeWithCountryAndVariant() == selfLong {
			continue
		}
		// language.getClass().isAssignableFrom(getClass())
		if language.isAssignableFrom(l) && language.ClassName != l.ClassName {
			return true
		}
	}
	return false
}

// IsTheDefaultVariant ports Language.isTheDefaultVariant (private).
// Empty DefaultVariantCode ≡ Java null getDefaultLanguageVariant() → false.
func (l LanguageMeta) IsTheDefaultVariant() bool {
	if l.DefaultVariantCode == "" {
		return false
	}
	return strings.EqualFold(l.GetShortCodeWithCountryAndVariant(), l.DefaultVariantCode)
}

// HasVariant ports Language.hasVariant: a registered subclass implements a variant.
// Fallback without ClassName: another registered language shares short code and this is not a variant.
func (L *Languages) HasVariant(l LanguageMeta) bool {
	if l.ClassName != "" {
		selfLong := l.GetShortCodeWithCountryAndVariant()
		for _, o := range L.Get() {
			if o.GetShortCodeWithCountryAndVariant() == selfLong && o.ClassName == l.ClassName {
				continue
			}
			// getClass().isAssignableFrom(language.getClass()) — l is superclass of o
			if l.isAssignableFrom(o) && o.ClassName != l.ClassName {
				return true
			}
		}
		// SimpleGerman extends GermanyGerman: GermanyGerman hasVariant? subclasses in registry: SimpleGerman
		// German has GermanyGerman etc.
		return false
	}
	// Fallback for ad-hoc test registrations without ClassName
	sc := l.GetShortCode()
	selfLong := l.GetShortCodeWithCountryAndVariant()
	for _, o := range L.Get() {
		if strings.EqualFold(o.GetShortCodeWithCountryAndVariant(), selfLong) {
			continue
		}
		if o.GetShortCode() == sc {
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
	return L.HasVariant(l) && !L.languageIsVariant(l) && !l.IsTheDefaultVariant()
}

// GetOrAddLanguageByClassName ports getOrAddLanguageByClassName for registered metas.
// When not found, panics (classpath ClassBroker not available for arbitrary classes).
func (L *Languages) GetOrAddLanguageByClassName(className string) LanguageMeta {
	simple := className
	if i := strings.LastIndex(className, "."); i >= 0 {
		simple = className[i+1:]
	}
	L.mu.RLock()
	for _, l := range L.allLocked() {
		if l.ClassName == simple || l.Name == simple || strings.EqualFold(l.Code, simple) {
			L.mu.RUnlock()
			return l
		}
	}
	L.mu.RUnlock()
	panic(fmt.Sprintf("Class '%s' could not be found in classpath", className))
}
