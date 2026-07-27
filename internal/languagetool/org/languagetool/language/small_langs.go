package language

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// SmallLang is metadata for languages ported with tagger/speller surfaces only.
type SmallLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

func (s SmallLang) GetName() string      { return s.Name }
func (s SmallLang) GetShortCode() string { return s.ShortCode }

// GetCountries returns a defensive copy of country codes.
func (s SmallLang) GetCountries() []string {
	return append([]string(nil), s.Countries...)
}

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (s SmallLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant(s.ShortCode, s.Countries, "")
}

// GetCommonWordsPath ports Language.getCommonWordsPath; km/ja return null (empty).
func (s SmallLang) GetCommonWordsPath() string {
	switch s.ShortCode {
	case "km", "ja":
		// Java Khmer/Japanese.getCommonWordsPath → null (TODO upstream)
		return CommonWordsPathNone
	default:
		return DefaultCommonWordsPath(s.ShortCode)
	}
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / Morfologik* getId when set.
func (s SmallLang) GetDefaultSpellingRuleID() string {
	return s.SpellerRuleID
}

// GetIgnoredCharactersRegex ports Language.getIgnoredCharactersRegex (+ uk override).
// Java Language default is soft hyphen [\u00AD]; Ukrainian overrides.
// Belarusian override: BelarusianLang.GetIgnoredCharactersRegex (belarusian.go).
func (s SmallLang) GetIgnoredCharactersRegex() *regexp.Regexp {
	switch s.ShortCode {
	case "uk":
		return languagetool.UkrainianIgnoredCharactersRegex
	default:
		return languagetool.GermanIgnoredCharactersRegex
	}
}

// GetOpeningDoubleQuote ports Language defaults + Ukrainian overrides.
// Belarusian quotes: BelarusianLang (belarusian.go).
func (s SmallLang) GetOpeningDoubleQuote() string {
	switch s.ShortCode {
	case "uk":
		return "«"
	default:
		// Language.java default "
		return "\""
	}
}

// GetClosingDoubleQuote ports Language defaults + Ukrainian overrides.
// Belarusian quotes: BelarusianLang (belarusian.go).
func (s SmallLang) GetClosingDoubleQuote() string {
	switch s.ShortCode {
	case "uk":
		return "»"
	default:
		return "\""
	}
}

// GetOpeningSingleQuote ports Language defaults + Ukrainian overrides.
// Belarusian quotes: BelarusianLang (belarusian.go).
func (s SmallLang) GetOpeningSingleQuote() string {
	switch s.ShortCode {
	case "uk":
		return "‘"
	default:
		return "'"
	}
}

// GetClosingSingleQuote ports Language defaults + Ukrainian overrides.
// Belarusian quotes: BelarusianLang (belarusian.go).
func (s SmallLang) GetClosingSingleQuote() string {
	switch s.ShortCode {
	case "uk":
		return "’"
	default:
		return "'"
	}
}

// IsAdvancedTypographyEnabled ports Language default (false).
// Belarusian true: BelarusianLang.IsAdvancedTypographyEnabled (belarusian.go).
// Ukrainian explicitly false (quotes set but advanced typography disabled).
func (s SmallLang) IsAdvancedTypographyEnabled() bool {
	return false
}

// ToAdvancedTypography for SmallLang: Ukrainian disabled; others base disabled.
// Belarusian: BelarusianLang.ToAdvancedTypography (belarusian.go).
func (s SmallLang) ToAdvancedTypography(input string) string {
	switch s.ShortCode {
	case "uk":
		return UkrainianAdvancedTypography(input)
	default:
		cfg := languagetool.TypographyConfig{
			Enabled:            false,
			OpeningDoubleQuote: s.GetOpeningDoubleQuote(),
			ClosingDoubleQuote: s.GetClosingDoubleQuote(),
			OpeningSingleQuote: s.GetOpeningSingleQuote(),
			ClosingSingleQuote: s.GetClosingSingleQuote(),
		}
		return languagetool.ToAdvancedTypography(input, cfg)
	}
}

// SpellerRuleID is createDefaultSpellingRule / registered speller getId only.
// Languages without a Java default speller use "" (do not invent Morfologik IDs).
var (
	// Slovak promoted to SlovakLang (see slovak.go) — not SmallLang.
	// Danish promoted to DanishLang (see danish.go) — not SmallLang.
	// Swedish/Esperanto/Galician: plain HunspellRule → HUNSPELL_RULE
	Swedish = SmallLang{"sv", "Swedish", "HUNSPELL_RULE", []string{"SE"}}
	// Romanian promoted to RomanianLang (see romanian.go) — not SmallLang.
	// Greek promoted to GreekLang (see greek.go) — not SmallLang.
	Galician = SmallLang{"gl", "Galician", "HUNSPELL_RULE", []string{"ES"}}
	// Japanese/Chinese/Persian: no createDefaultSpellingRule / no speller in getRelevantRules
	Japanese  = SmallLang{"ja", "Japanese", "", []string{"JP"}}
	Chinese   = SmallLang{"zh", "Chinese", "", []string{"CN"}}
	Persian   = SmallLang{"fa", "Persian", "", []string{"IR"}}
	Esperanto = SmallLang{"eo", "Esperanto", "HUNSPELL_RULE", nil}
	// MorfologikIrishSpellerRule.getId
	Irish     = SmallLang{"ga", "Irish", "MORFOLOGIK_RULE_GA_IE", []string{"IE"}}
	Ukrainian = SmallLang{"uk", "Ukrainian", "MORFOLOGIK_RULE_UK_UA", []string{"UA"}}
)

func AllSmallLangs() []SmallLang {
	return []SmallLang{
		Swedish, Galician,
		Japanese, Chinese, Persian, Esperanto, Irish, Ukrainian,
	}
}

// GetMaintainedState ports Language.getMaintainedState for small languages.
// Java default is LookingForNewMaintainer; only languages that override return ActivelyMaintained.
func (s SmallLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	switch s.ShortCode {
	case "sv", "ga", "uk", "eo":
		// Swedish, Irish, Ukrainian, Esperanto
		// Greek/Breton/CrimeanTatar are *Lang — ActivelyMaintained there
		return languagetool.ActivelyMaintained
	case "gl":
		// Galician explicitly LookingForNewMaintainer (same as default)
		return languagetool.LookingForNewMaintainer
	default:
		// Japanese, Chinese, Persian, …: no override → default
		// (Danish/Romanian/Slovak/Slovenian/Tagalog/Asturian/Belarusian LookingForNewMaintainer on *Lang; Greek/Breton/CrimeanTatar ActivelyMaintained on *Lang)
		return languagetool.LookingForNewMaintainer
	}
}

// GetMaintainers ports Language.getMaintainers for small-language modules (Java names exact).
func (s SmallLang) GetMaintainers() []Contributor {
	switch s.ShortCode {
	// Belarusian maintainers: BelarusianLang.GetMaintainers (belarusian.go)
	// Slovak maintainers: SlovakLang.GetMaintainers (slovak.go)
	case "sv":
		return []Contributor{NewContributor("Leif-Jöran Olsson")}
	// Greek maintainers: GreekLang.GetMaintainers (greek.go)
	// Danish maintainers: DanishLang.GetMaintainers (danish.go)
	case "eo":
		return []Contributor{DominiquePelle}
	// Breton maintainers: BretonLang.GetMaintainers (breton.go)
	case "ga":
		return []Contributor{
			NewContributor("Jim O'Regan"),
			NewContributor("Emily Barnes"),
			NewContributor("Mícheál J. Ó Meachair"),
			NewContributor("Seanán Ó Coistín"),
		}
	case "gl":
		return []Contributor{
			NewContributor("Susana Sotelo Docío"),
			NewContributorWithURL("Tiago F. Santos (4.0-4.7)", "https://github.com/TiagoSantos81"),
		}
	case "lt":
		return []Contributor{NewContributor("Mantas Kriaučiūnas")}
	case "is":
		return []Contributor{NewContributor("Anton Karl Ingason")}
	// Asturian maintainers: AsturianLang.GetMaintainers (asturian.go)
	// Slovenian maintainers: SlovenianLang.GetMaintainers (slovenian.go)
	// Romanian maintainers: RomanianLang.GetMaintainers (romanian.go)
	case "fa":
		return []Contributor{
			NewContributor("Reza1615"),
			NewContributor("Alireza Eskandarpour Shoferi"),
			NewContributor("Ebrahim Byagowi"),
		}
	case "zh":
		return []Contributor{NewContributor("Tao Lin")}
	case "ja":
		return []Contributor{NewContributor("Takahiro Shinkai")}
	case "km":
		return []Contributor{NewContributor("Nathan Wells")}
	case "ta":
		return []Contributor{NewContributor("Elanjelian Venugopal")}
	case "ml":
		return []Contributor{NewContributor("Jithesh.V.S")}
	// Tagalog maintainers: TagalogLang.GetMaintainers (tagalog.go)
	// CrimeanTatar maintainers: CrimeanTatarLang.GetMaintainers (crimean_tatar.go)
	case "uk":
		return []Contributor{
			NewContributor("Andriy Rysin"),
			NewContributor("Maksym Davydov"),
		}
	default:
		return nil
	}
}

// GetRuleFileNames ports Language.getRuleFileNames + Ukrainian extras.
// Slovak RULE_FILES: SlovakLang.GetRuleFileNames (slovak.go).
// Ukrainian adds RULE_FILES list.
func (s SmallLang) GetRuleFileNames() []string {
	return s.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence.
func (s SmallLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := s.ShortCode
	out := languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
	dirBase := "/org/languagetool/rules/" + code + "/"
	switch code {
	case "uk":
		// Ukrainian.RULE_FILES (same order as Java)
		for _, f := range UkrainianLanguageDefault.RuleFiles {
			out = append(out, dirBase+f)
		}
	}
	return out
}
