package rules

import (
	"fmt"
	"strings"
)

// ITSIssueType ports org.languagetool.rules.ITSIssueType (ITS 2.0 LQ issue types).
type ITSIssueType string

const (
	ITSTerminology           ITSIssueType = "terminology"
	ITSMistranslation        ITSIssueType = "mistranslation"
	ITSOmission              ITSIssueType = "omission"
	ITSUntranslated          ITSIssueType = "untranslated"
	ITSAddition              ITSIssueType = "addition"
	ITSDuplication           ITSIssueType = "duplication"
	ITSInconsistency         ITSIssueType = "inconsistency"
	ITSGrammar               ITSIssueType = "grammar"
	ITSLegal                 ITSIssueType = "legal"
	ITSRegister              ITSIssueType = "register"
	ITSLocaleSpecificContent ITSIssueType = "locale-specific-content"
	ITSLocaleViolation       ITSIssueType = "locale-violation"
	ITSStyle                 ITSIssueType = "style"
	ITSCharacters            ITSIssueType = "characters"
	ITSMisspelling           ITSIssueType = "misspelling"
	ITSTypographical         ITSIssueType = "typographical"
	ITSFormatting            ITSIssueType = "formatting"
	ITSInconsistentEntities  ITSIssueType = "inconsistent-entities"
	ITSNumbers               ITSIssueType = "numbers"
	ITSMarkup                ITSIssueType = "markup"
	ITSPatternProblem        ITSIssueType = "pattern-problem"
	ITSWhitespace            ITSIssueType = "whitespace"
	ITSInternationalization  ITSIssueType = "internationalization"
	ITSLength                ITSIssueType = "length"
	ITSNonConformance        ITSIssueType = "non-conformance"
	ITSUncategorized         ITSIssueType = "uncategorized"
	ITSOther                 ITSIssueType = "other"
)

// AllITSIssueTypes lists every known type.
var AllITSIssueTypes = []ITSIssueType{
	ITSTerminology, ITSMistranslation, ITSOmission, ITSUntranslated, ITSAddition,
	ITSDuplication, ITSInconsistency, ITSGrammar, ITSLegal, ITSRegister,
	ITSLocaleSpecificContent, ITSLocaleViolation, ITSStyle, ITSCharacters,
	ITSMisspelling, ITSTypographical, ITSFormatting, ITSInconsistentEntities,
	ITSNumbers, ITSMarkup, ITSPatternProblem, ITSWhitespace, ITSInternationalization,
	ITSLength, ITSNonConformance, ITSUncategorized, ITSOther,
}

// GetIssueType looks up by ITS 2.0 string name (lowercase/hyphen form).
func GetIssueType(name string) (ITSIssueType, error) {
	for _, t := range AllITSIssueTypes {
		if string(t) == name {
			return t, nil
		}
	}
	return "", fmt.Errorf("no IssueType found for name %q", name)
}

func (t ITSIssueType) String() string {
	return string(t)
}

// ParseIssueTypeCamel maps Java enum-style names (e.g. "Grammar") to ITS strings.
func ParseIssueTypeCamel(name string) (ITSIssueType, error) {
	// try direct first
	if t, err := GetIssueType(name); err == nil {
		return t, nil
	}
	// camelCase / PascalCase → hyphenated lowercase
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('-')
		}
		b.WriteRune(r)
	}
	return GetIssueType(strings.ToLower(b.String()))
}
