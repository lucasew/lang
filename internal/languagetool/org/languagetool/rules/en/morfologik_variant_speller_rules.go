package en

import (
	"strings"

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

func newVariantSpeller(id, variantCode, dictPath, variantSpellingFile, otherName string, other map[string]string) *MorfologikVariantSpellerRule {
	base := NewAbstractEnglishSpellerRule(id, variantCode, morfologik.NewMorfologikSpeller(dictPath, 1))
	base.FileName = dictPath
	return &MorfologikVariantSpellerRule{
		AbstractEnglishSpellerRule:  base,
		OtherVariant:                other,
		OtherVariantName:            otherName,
		LanguageVariantSpellingFile: variantSpellingFile,
	}
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

func NewMorfologikAmericanSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikAmericanSpellerRuleID, "en-US", AmericanSpellerDict,
		AmericanVariantSpellingFile, "British English", nil)
}

func NewMorfologikBritishSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikBritishSpellerRuleID, "en-GB", BritishSpellerDict,
		BritishVariantSpellingFile, "American English", nil)
}

func NewMorfologikCanadianSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikCanadianSpellerRuleID, "en-CA", CanadianSpellerDict,
		CanadianVariantSpellingFile, "American English", nil)
}

func NewMorfologikAustralianSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikAustralianSpellerRuleID, "en-AU", AustralianSpellerDict,
		AustralianVariantSpellingFile, "American English", nil)
}

func NewMorfologikNewZealandSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikNewZealandSpellerRuleID, "en-NZ", NewZealandSpellerDict,
		NewZealandVariantSpellingFile, "American English", nil)
}

func NewMorfologikSouthAfricanSpellerRule() *MorfologikVariantSpellerRule {
	return newVariantSpeller(MorfologikSouthAfricanSpellerRuleID, "en-ZA", SouthAfricanSpellerDict,
		SouthAfricanVariantSpellingFile, "American English", nil)
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
