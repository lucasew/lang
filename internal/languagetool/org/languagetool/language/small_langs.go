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

// GetIgnoredCharactersRegex ports Language.getIgnoredCharactersRegex (+ be/uk overrides).
// Java Language default is soft hyphen [\u00AD]; Belarusian/Ukrainian override.
func (s SmallLang) GetIgnoredCharactersRegex() *regexp.Regexp {
	switch s.ShortCode {
	case "be":
		return languagetool.BelarusianIgnoredCharactersRegex
	case "uk":
		return languagetool.UkrainianIgnoredCharactersRegex
	default:
		return languagetool.GermanIgnoredCharactersRegex
	}
}

// GetOpeningDoubleQuote ports Language defaults + Belarusian/Ukrainian overrides.
func (s SmallLang) GetOpeningDoubleQuote() string {
	switch s.ShortCode {
	case "be", "uk":
		return "«"
	default:
		// Language.java default "
		return "\""
	}
}

// GetClosingDoubleQuote ports Language defaults + Belarusian/Ukrainian overrides.
func (s SmallLang) GetClosingDoubleQuote() string {
	switch s.ShortCode {
	case "be", "uk":
		return "»"
	default:
		return "\""
	}
}

// GetOpeningSingleQuote ports Language defaults + Belarusian/Ukrainian overrides.
func (s SmallLang) GetOpeningSingleQuote() string {
	switch s.ShortCode {
	case "be", "uk":
		return "‘"
	default:
		return "'"
	}
}

// GetClosingSingleQuote ports Language defaults + Belarusian/Ukrainian overrides.
func (s SmallLang) GetClosingSingleQuote() string {
	switch s.ShortCode {
	case "be", "uk":
		return "’"
	default:
		return "'"
	}
}

// IsAdvancedTypographyEnabled ports Language default (false) + be true / uk false.
func (s SmallLang) IsAdvancedTypographyEnabled() bool {
	return s.ShortCode == "be" // Belarusian true; Ukrainian false; others default false
}

// ToAdvancedTypography for SmallLang: Belarusian enabled; Ukrainian disabled; others base disabled.
func (s SmallLang) ToAdvancedTypography(input string) string {
	switch s.ShortCode {
	case "be":
		return BelarusianAdvancedTypography(input)
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
	Slovak = SmallLang{"sk", "Slovak", "MORFOLOGIK_RULE_SK_SK", []string{"SK"}}
	// Danish/Swedish/Esperanto/Galician: plain HunspellRule → HUNSPELL_RULE
	Danish    = SmallLang{"da", "Danish", "HUNSPELL_RULE", []string{"DK"}}
	Swedish   = SmallLang{"sv", "Swedish", "HUNSPELL_RULE", []string{"SE"}}
	Romanian  = SmallLang{"ro", "Romanian", "MORFOLOGIK_RULE_RO_RO", []string{"RO"}}
	Greek     = SmallLang{"el", "Greek", "MORFOLOGIK_RULE_EL_GR", []string{"GR"}}
	Galician  = SmallLang{"gl", "Galician", "HUNSPELL_RULE", []string{"ES"}}
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
		Slovak, Danish, Swedish, Romanian, Greek, Galician,
		Japanese, Chinese, Persian, Esperanto, Irish, Ukrainian,
	}
}

// GetMaintainedState ports Language.getMaintainedState for small languages.
// Java default is LookingForNewMaintainer; only languages that override return ActivelyMaintained.
func (s SmallLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	switch s.ShortCode {
	case "sv", "el", "ga", "uk", "eo", "br", "crh":
		// Swedish, Greek, Irish, Ukrainian, Esperanto, Breton, CrimeanTatar
		return languagetool.ActivelyMaintained
	case "gl":
		// Galician explicitly LookingForNewMaintainer (same as default)
		return languagetool.LookingForNewMaintainer
	default:
		// Japanese, Chinese, Persian, Slovak, Danish, Romanian, …: no override → default
		return languagetool.LookingForNewMaintainer
	}
}

// GetMaintainers ports Language.getMaintainers for small-language modules (Java names exact).
func (s SmallLang) GetMaintainers() []Contributor {
	switch s.ShortCode {
	case "be":
		return []Contributor{NewContributor("Alex Buloichik")}
	case "sk":
		return []Contributor{NewContributorWithURL("Zdenko Podobný", "http://sk-spell.sk.cx")}
	case "sv":
		return []Contributor{NewContributor("Leif-Jöran Olsson")}
	case "el":
		return []Contributor{NewContributor("Panagiotis Minos")}
	case "da":
		return []Contributor{
			NewContributor("Esben Aaberg"),
			NewContributor("Henrik Bendt"),
		}
	case "eo":
		return []Contributor{DominiquePelle}
	case "br":
		return []Contributor{
			DominiquePelle,
			NewContributor("Fulup Jakez"),
		}
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
	case "ast":
		return []Contributor{NewContributor("Xesús González Rato")}
	case "sl":
		return []Contributor{NewContributor("Martin Srebotnjak")}
	case "ro":
		return []Contributor{NewContributorWithURL("Ionuț Păduraru", "http://www.archeus.ro")}
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
	case "tl":
		return []Contributor{
			NewContributor("Nathaniel Oco"),
			NewContributor("Allan Borra"),
		}
	case "crh":
		return []Contributor{NewContributor("Andriy Rysin")}
	case "uk":
		return []Contributor{
			NewContributor("Andriy Rysin"),
			NewContributor("Maksym Davydov"),
		}
	default:
		return nil
	}
}

// GetRuleFileNames ports Language.getRuleFileNames + Slovak/Ukrainian extras.
// Slovak adds grammar-typography.xml; Ukrainian adds RULE_FILES list.
func (s SmallLang) GetRuleFileNames() []string {
	return s.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence.
func (s SmallLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := s.ShortCode
	out := languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
	dirBase := "/org/languagetool/rules/" + code + "/"
	switch code {
	case "sk":
		// Slovak.RULE_FILES → grammar-typography.xml
		out = append(out, dirBase+"grammar-typography.xml")
	case "uk":
		// Ukrainian.RULE_FILES (same order as Java)
		for _, f := range UkrainianLanguageDefault.RuleFiles {
			out = append(out, dirBase+f)
		}
	}
	return out
}
