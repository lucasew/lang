package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ItalianLang ports Italian language metadata.
type ItalianLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

func (v ItalianLang) GetName() string { return v.Name }

// GetShortCode ports Italian.getShortCode ("it").
func (v ItalianLang) GetShortCode() string { return "it" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v ItalianLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("it", v.Countries, "")
}

func (v ItalianLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → it/common_words.txt.
func (v ItalianLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetMaintainedState ports Italian.getMaintainedState → ActivelyMaintained.
func (v ItalianLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Italian.getMaintainers → Paolo Bianchini.
func (v ItalianLang) GetMaintainers() []Contributor {
	return []Contributor{NewContributor("Paolo Bianchini")}
}

// SpellerRuleID ports MorfologikItalianSpellerRule.getId → MORFOLOGIK_RULE_IT_IT.
var Italian = ItalianLang{ShortCode: "it", Name: "Italian", Countries: []string{"IT"}, SpellerRuleID: "MORFOLOGIK_RULE_IT_IT"}

// GetDefaultSpellingRuleID ports MorfologikItalianSpellerRule getId.
func (v ItalianLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_IT_IT"
}

// Dutch variants.
type DutchVariant struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

func (v DutchVariant) GetName() string { return v.Name }

// GetShortCode ports Dutch.getShortCode ("nl" for all locales).
func (v DutchVariant) GetShortCode() string { return "nl" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
// BelgianDutch is stored as ShortCode "nl-BE"; Netherlands uses countries NL → "nl-NL".
func (v DutchVariant) GetShortCodeWithCountryAndVariant() string {
	if isBelgianDutch(v) {
		return "nl-BE"
	}
	return BuildShortCodeWithCountryAndVariant("nl", v.Countries, "")
}

func (v DutchVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → nl/common_words.txt.
func (v DutchVariant) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetMaintainedState ports Dutch.getMaintainedState → ActivelyMaintained.
func (v DutchVariant) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Dutch.getMaintainers (OpenTaal, TaalTik).
func (v DutchVariant) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributorWithURL("OpenTaal", "http://www.opentaal.org"),
		NewContributorWithURL("TaalTik", "http://www.taaltik.nl"),
	}
}

// GetDefaultSpellingRuleID ports MorfologikDutchSpellerRule getId (MORFOLOGIK_RULE_NL_NL).
func (v DutchVariant) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_NL_NL"
}

// GetOpeningDoubleQuote ports Dutch.getOpeningDoubleQuote ("“").
func (v DutchVariant) GetOpeningDoubleQuote() string { return "“" }

// GetClosingDoubleQuote ports Dutch.getClosingDoubleQuote ("”").
func (v DutchVariant) GetClosingDoubleQuote() string { return "”" }

// GetOpeningSingleQuote ports Dutch.getOpeningSingleQuote ("‘").
func (v DutchVariant) GetOpeningSingleQuote() string { return "‘" }

// GetClosingSingleQuote ports Dutch.getClosingSingleQuote ("’").
func (v DutchVariant) GetClosingSingleQuote() string { return "’" }

// IsAdvancedTypographyEnabled ports Dutch.isAdvancedTypographyEnabled (true).
func (v DutchVariant) IsAdvancedTypographyEnabled() bool { return true }

// ToAdvancedTypography ports Dutch base Language.toAdvancedTypography with Dutch quotes.
func (v DutchVariant) ToAdvancedTypography(input string) string {
	return DutchAdvancedTypography(input)
}

// nlNLGrammarXML ports Dutch.getRuleFileNames append of nl/nl-NL/grammar.xml.
const nlNLGrammarXML = "/org/languagetool/rules/nl/nl-NL/grammar.xml"

// GetRuleFileNames ports Dutch / BelgianDutch.getRuleFileNames.
// Dutch adds nl/nl-NL/grammar.xml; BelgianDutch removes it after super.
func (v DutchVariant) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists is GetRuleFileNames with a pluggable exists probe.
func (v DutchVariant) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	// Dutch short code is always "nl" (BelgianDutch extends Dutch).
	out := languagetool.GetRuleFileNames("nl", v.ShortCode, "/org/languagetool/rules", exists)
	if isBelgianDutch(v) {
		// BelgianDutch removes nl-NL/grammar.xml that Dutch would add
		return filterPath(out, nlNLGrammarXML)
	}
	// Dutch adds nl/nl-NL/grammar.xml (even if super already listed a variant path)
	if !containsPath(out, nlNLGrammarXML) {
		out = append(out, nlNLGrammarXML)
	}
	return out
}

func isBelgianDutch(v DutchVariant) bool {
	return strings.EqualFold(v.ShortCode, "nl-BE") || strings.HasSuffix(strings.ToUpper(v.ShortCode), "-BE")
}

func containsPath(paths []string, p string) bool {
	for _, x := range paths {
		if x == p {
			return true
		}
	}
	return false
}

func filterPath(paths []string, drop string) []string {
	out := make([]string, 0, len(paths))
	for _, x := range paths {
		if x != drop {
			out = append(out, x)
		}
	}
	return out
}

var (
	DutchNetherlands = DutchVariant{ShortCode: "nl", Name: "Dutch", Countries: []string{"NL"}, SpellerRuleID: "MORFOLOGIK_RULE_NL_NL"}
	BelgianDutch     = DutchVariant{ShortCode: "nl-BE", Name: "Dutch (Belgium)", Countries: []string{"BE"}, SpellerRuleID: "MORFOLOGIK_RULE_NL_NL"}
)

func AllDutchVariants() []DutchVariant {
	return []DutchVariant{DutchNetherlands, BelgianDutch}
}

// PolishLang ports Polish language metadata.
type PolishLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

func (v PolishLang) GetName() string { return v.Name }

// GetShortCode ports Polish.getShortCode ("pl").
func (v PolishLang) GetShortCode() string { return "pl" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v PolishLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("pl", v.Countries, "")
}

func (v PolishLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → pl/common_words.txt.
func (v PolishLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetMaintainedState ports Polish.getMaintainedState → ActivelyMaintained.
func (v PolishLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Polish.getMaintainers → Marcin Miłkowski.
func (v PolishLang) GetMaintainers() []Contributor {
	return []Contributor{MarcinMilkowski}
}

// SpellerRuleID ports MorfologikPolishSpellerRule.getId → MORFOLOGIK_RULE_PL_PL.
var Polish = PolishLang{ShortCode: "pl", Name: "Polish", Countries: []string{"PL"}, SpellerRuleID: "MORFOLOGIK_RULE_PL_PL"}

// GetDefaultSpellingRuleID ports MorfologikPolishSpellerRule getId.
func (v PolishLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_PL_PL"
}

// Catalan variants.
type CatalanVariant struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
	Valencian                      bool
	Balearic                       bool
}

func (v CatalanVariant) GetName() string { return v.Name }

// GetShortCode ports Catalan.getShortCode ("ca" for all locales).
func (v CatalanVariant) GetShortCode() string { return "ca" }

// GetVariant ports Catalan.getVariant — "valencia" / "balear" / "".
func (v CatalanVariant) GetVariant() string {
	if v.Valencian {
		return "valencia"
	}
	if v.Balearic {
		return "balear"
	}
	return ""
}

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant
// (ca-ES, ca-ES-valencia, ca-ES-balear).
func (v CatalanVariant) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("ca", v.Countries, v.GetVariant())
}

func (v CatalanVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → ca/common_words.txt.
func (v CatalanVariant) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetMaintainedState ports Catalan.getMaintainedState → ActivelyMaintained.
func (v CatalanVariant) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Catalan.getMaintainers (Ricard Roca, Jaume Ortolà).
func (v CatalanVariant) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Ricard Roca"),
		NewContributor("Jaume Ortolà"),
	}
}

var (
	// SpellerRuleID ports MorfologikCatalanSpellerRule.getId → MORFOLOGIK_RULE_CA_ES.
	Catalan          = CatalanVariant{ShortCode: "ca", Name: "Catalan", Countries: []string{"ES"}, SpellerRuleID: "MORFOLOGIK_RULE_CA_ES"}
	ValencianCatalan = CatalanVariant{ShortCode: "ca-ES-valencia", Name: "Catalan (Valencian)", Countries: []string{"ES"}, SpellerRuleID: "MORFOLOGIK_RULE_CA_ES", Valencian: true}
	BalearicCatalan  = CatalanVariant{ShortCode: "ca-ES-balear", Name: "Catalan (Balearic)", Countries: []string{"ES"}, SpellerRuleID: "MORFOLOGIK_RULE_CA_ES", Balearic: true}
)

// GetDefaultSpellingRuleID ports MorfologikCatalanSpellerRule getId.
func (v CatalanVariant) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_CA_ES"
}

func AllCatalanVariants() []CatalanVariant {
	return []CatalanVariant{Catalan, ValencianCatalan, BalearicCatalan}
}

// RussianLang ports Russian language metadata.
type RussianLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

func (v RussianLang) GetName() string { return v.Name }

// GetShortCode ports Russian.getShortCode ("ru").
func (v RussianLang) GetShortCode() string { return "ru" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v RussianLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("ru", v.Countries, "")
}

func (v RussianLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → ru/common_words.txt.
func (v RussianLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetMaintainedState ports Russian.getMaintainedState → ActivelyMaintained.
func (v RussianLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Russian.getMaintainers → Yakov Reztsov.
func (v RussianLang) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributorWithURL("Yakov Reztsov", "http://myooo.ru/content/view/83/43/"),
	}
}

// GetIgnoredCharactersRegex ports Russian.getIgnoredCharactersRegex.
func (v RussianLang) GetIgnoredCharactersRegex() *regexp.Regexp {
	return languagetool.RussianIgnoredCharactersRegex
}

// GetDefaultSpellingRuleID ports MorfologikRussianSpellerRule getId.
func (v RussianLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_RU_RU"
}

// GetOpeningDoubleQuote ports Russian.getOpeningDoubleQuote ("«").
func (v RussianLang) GetOpeningDoubleQuote() string { return "«" }

// GetClosingDoubleQuote ports Russian.getClosingDoubleQuote ("»").
func (v RussianLang) GetClosingDoubleQuote() string { return "»" }

// GetOpeningSingleQuote ports Russian.getOpeningSingleQuote ("‘").
func (v RussianLang) GetOpeningSingleQuote() string { return "‘" }

// GetClosingSingleQuote ports Russian.getClosingSingleQuote ("’").
func (v RussianLang) GetClosingSingleQuote() string { return "’" }

// IsAdvancedTypographyEnabled ports Russian.isAdvancedTypographyEnabled (true).
func (v RussianLang) IsAdvancedTypographyEnabled() bool { return true }

// ToAdvancedTypography ports Russian base Language.toAdvancedTypography with Russian quotes.
func (v RussianLang) ToAdvancedTypography(input string) string {
	return RussianAdvancedTypography(input)
}

var Russian = RussianLang{ShortCode: "ru", Name: "Russian", Countries: []string{"RU"}, SpellerRuleID: "MORFOLOGIK_RULE_RU_RU"}