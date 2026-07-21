package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// Variant speller rule IDs and dictionary paths (Morfologik*SpellerRule ports).

const (
	MorfologikAmericanSpellerRuleID     = "MORFOLOGIK_RULE_EN_US"
	MorfologikBritishSpellerRuleID      = "MORFOLOGIK_RULE_EN_GB"
	MorfologikCanadianSpellerRuleID     = "MORFOLOGIK_RULE_EN_CA"
	MorfologikAustralianSpellerRuleID   = "MORFOLOGIK_RULE_EN_AU"
	MorfologikNewZealandSpellerRuleID   = "MORFOLOGIK_RULE_EN_NZ"
	MorfologikSouthAfricanSpellerRuleID = "MORFOLOGIK_RULE_EN_ZA"
	AmericanSpellerDict                 = "/en/hunspell/en_US.dict"
	BritishSpellerDict                  = "/en/hunspell/en_GB.dict"
	CanadianSpellerDict                 = "/en/hunspell/en_CA.dict"
	AustralianSpellerDict               = "/en/hunspell/en_AU.dict"
	NewZealandSpellerDict               = "/en/hunspell/en_NZ.dict"
	SouthAfricanSpellerDict             = "/en/hunspell/en_ZA.dict"
	AmericanVariantSpellingFile         = "en/hunspell/spelling_en-US.txt"
	BritishVariantSpellingFile          = "en/hunspell/spelling_en-GB.txt"
	CanadianVariantSpellingFile         = "en/hunspell/spelling_en-CA.txt"
	AustralianVariantSpellingFile       = "en/hunspell/spelling_en-AU.txt"
	NewZealandVariantSpellingFile       = "en/hunspell/spelling_en-NZ.txt"
	SouthAfricanVariantSpellingFile     = "en/hunspell/spelling_en-ZA.txt"
)

// MorfologikVariantSpellerRule is a thin AbstractEnglishSpellerRule for a locale.
type MorfologikVariantSpellerRule struct {
	*AbstractEnglishSpellerRule
	OtherVariant                map[string]string
	OtherVariantName            string
	LanguageVariantSpellingFile string
}

// englishPlainSpellingRels ports getSpellingFileName + getAdditionalSpellingFileNames.
var englishPlainSpellingRels = []string{
	"en/hunspell/spelling.txt",
	"spelling.txt", // EnglishGlobalSpellingFile
	"en/multiwords.txt",
}

func newVariantSpeller(id, variantCode, dictPath, variantSpellingFile, otherName string, other map[string]string) *MorfologikVariantSpellerRule {
	return newVariantSpellerWithUser(id, variantCode, dictPath, variantSpellingFile, otherName, other, nil)
}

// newVariantSpellerWithUser ports Morfologik*SpellerRule(..., UserConfig).
func newVariantSpellerWithUser(id, variantCode, dictPath, variantSpellingFile, otherName string, other map[string]string, userConfig *languagetool.UserConfig) *MorfologikVariantSpellerRule {
	// Java MorfologikSpellerRule.initSpeller: three Multis at maxEditDistance 1, 2, 3.
	// User dict FSA only when premiumUid + accepted words (Java getUserDictSpellerOrNull).
	var userWords []string
	var premium *int64
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		premium = userConfig.GetPremiumUid()
		userWords = morfologik.UserDictWordsForMulti(accepted, premium)
	}
	plainRels := append([]string(nil), englishPlainSpellingRels...)
	s1 := morfologik.OpenMultiSpellerFromClasspathWithUser(dictPath, plainRels, variantSpellingFile, 1, PrepareLineForSpeller, userWords)
	s2 := morfologik.OpenMultiSpellerFromClasspathWithUser(dictPath, plainRels, variantSpellingFile, 2, PrepareLineForSpeller, userWords)
	s3 := morfologik.OpenMultiSpellerFromClasspathWithUser(dictPath, plainRels, variantSpellingFile, 3, PrepareLineForSpeller, userWords)
	var sp *morfologik.MorfologikSpeller
	if s1 != nil && len(s1.DefaultDictSpellers) > 0 {
		sp = s1.DefaultDictSpellers[0]
	} else if s1 != nil && len(s1.Spellers) > 0 {
		sp = s1.Spellers[0]
	} else {
		sp = morfologik.NewMorfologikSpeller(dictPath, 1)
		_ = sp.TryAttachBinaryFromClasspath(dictPath)
	}
	base := NewAbstractEnglishSpellerRule(id, variantCode, sp)
	base.FileName = dictPath
	base.SetMultiSpellers(s1, s2, s3)
	// Java Morfologik*SpellerRule.getLanguageVariantSpellingFileName → plain-text accept list.
	// ApplyDefaultSpellingWordLists used LanguageShortCode "en" only; load variant file here.
	if base.SpellingCheckRule != nil {
		// Prefer full variant (en-US) for LanguageVariantSpellingClasspath inside ApplyDefault.
		base.SpellingCheckRule.LanguageCode = variantCode
		spelling.ApplyDefaultSpellingWordLists(base.SpellingCheckRule)
		// Explicit variant path (same as Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT constant).
		spelling.ApplyVariantSpellingFile(base.SpellingCheckRule, variantSpellingFile)
		// Java SpellingCheckRule: wordsToBeIgnored.addAll(userConfig.getAcceptedWords()) always.
		base.SpellingCheckRule.ApplyUserAcceptedWords(accepted)
	}
	r := &MorfologikVariantSpellerRule{
		AbstractEnglishSpellerRule:  base,
		OtherVariant:                other,
		OtherVariantName:            otherName,
		LanguageVariantSpellingFile: variantSpellingFile,
	}
	// Java isValidInOtherVariant override used from getRuleMatches post-path.
	r.IsValidInOtherVariantFn = r.IsValidInOtherVariant
	return r
}

func (r *MorfologikVariantSpellerRule) GetFileName() string { return r.FileName }

func (r *MorfologikVariantSpellerRule) GetLanguageVariantSpellingFileName() string {
	return r.LanguageVariantSpellingFile
}

// IsValidInOtherVariant ports isValidInOtherVariant.
func (r *MorfologikVariantSpellerRule) IsValidInOtherVariant(word string) *VariantInfo {
	if r == nil || r.OtherVariant == nil {
		return nil
	}
	if form, ok := r.OtherVariant[strings.ToLower(word)]; ok {
		v := NewVariantInfo(r.OtherVariantName, form)
		return &v
	}
	return nil
}

// usGbVariantMap loads en/en-US-GB.txt (US;GB). column 0: US→GB key; column 1: GB→US key.
// Java loadWordlist(path, column).
func usGbVariantMap(column int) map[string]string {
	return LoadUSGBVariantMap(column)
}

func NewMorfologikAmericanSpellerRule() *MorfologikVariantSpellerRule {
	return NewMorfologikAmericanSpellerRuleWithUser(nil)
}

// NewMorfologikAmericanSpellerRuleWithUser ports MorfologikAmericanSpellerRule(..., UserConfig).
func NewMorfologikAmericanSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikVariantSpellerRule {
	// Java: loadWordlist("en/en-US-GB.txt", 1) — British form as key → American form
	r := newVariantSpellerWithUser(MorfologikAmericanSpellerRuleID, "en-US", AmericanSpellerDict,
		AmericanVariantSpellingFile, "British English", usGbVariantMap(1), userConfig)
	// Java MorfologikAmericanSpellerRule.getAdditionalTopSuggestions: automize*
	if r.AbstractEnglishSpellerRule != nil && r.MorfologikSpellerRule != nil {
		baseFn := r.GetAdditionalTopSuggestionsFn
		r.GetAdditionalTopSuggestionsFn = func(existing []string, word string) []string {
			switch word {
			case "automize":
				return []string{"automate"}
			case "automized":
				return []string{"automated"}
			case "automizing":
				return []string{"automating"}
			case "automizes":
				return []string{"automates"}
			}
			if baseFn != nil {
				return baseFn(existing, word)
			}
			return nil
		}
	}
	return r
}

func NewMorfologikBritishSpellerRule() *MorfologikVariantSpellerRule {
	// Java: loadWordlist("en/en-US-GB.txt", 0) — American form as key → British form
	return newVariantSpeller(MorfologikBritishSpellerRuleID, "en-GB", BritishSpellerDict,
		BritishVariantSpellingFile, "American English", usGbVariantMap(0))
}

func NewMorfologikCanadianSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikCanadianSpellerRuleID, "en-CA", CanadianSpellerDict,
		CanadianVariantSpellingFile, "American English", usGbVariantMap(0))
}

func NewMorfologikAustralianSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikAustralianSpellerRuleID, "en-AU", AustralianSpellerDict,
		AustralianVariantSpellingFile, "American English", usGbVariantMap(0))
}

func NewMorfologikNewZealandSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikNewZealandSpellerRuleID, "en-NZ", NewZealandSpellerDict,
		NewZealandVariantSpellingFile, "American English", usGbVariantMap(0))
}

func NewMorfologikSouthAfricanSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikSouthAfricanSpellerRuleID, "en-ZA", SouthAfricanSpellerDict,
		SouthAfricanVariantSpellingFile, "American English", usGbVariantMap(0))
}

// LoadOtherVariantMap loads "local\tother" lines; column 1 reverses mapping.
func LoadOtherVariantMap(lines []string, column int) map[string]string {
	m := map[string]string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var a, b string
		if i := strings.IndexByte(line, '\t'); i >= 0 {
			a, b = line[:i], line[i+1:]
		} else if i := strings.IndexByte(line, '='); i >= 0 {
			a, b = line[:i], line[i+1:]
		} else {
			continue
		}
		a, b = strings.TrimSpace(a), strings.TrimSpace(b)
		if column == 1 {
			a, b = b, a
		}
		if a != "" && b != "" {
			m[strings.ToLower(a)] = b
		}
	}
	return m
}
