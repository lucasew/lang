package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

const TokenAgreementNumrNounRuleID = "UK_NUMR_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementNumrNounRule ports TokenAgreementNumrNounRule.
type TokenAgreementNumrNounRule struct {
	*tokenAgreementMatch
}

func hasNumrReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSNumr.Match(p) || taguk.IPOSNumber.Match(p) {
			return true
		}
	}
	return false
}

// nounForcePattern ports TokenAgreementNumrNounRule.NOUN_FORCE_PATTERN (Java Matcher.matches).
// Full-string only — do not invent extra plurals like "тони" beyond the Java regex.
var nounForcePattern = regexp.MustCompile(
	`^(?:чоловік|солдат|тон|(?:нано|мікро|мілі|дека|кіло|мега|гіга|тера|пета)?(?:герц|байт|біт|бар|бер|ват|вольт|децибел|рентген|моль|мікрон|грам|аршин|лат|карат))$`,
)

// Java TokenAgreementNumrNounRule surface patterns for fractional / half numerals
// (Matcher.matches / String.matches — full-string). Incomplete vs full Java branch
// messages; used only as soft skip on the simplified pair checker (no invent tokens).
var (
	// numrToken.matches("(один-|одне-)?півтора")
	numrPivtoraRE = regexp.MustCompile(`^(?:один-|одне-)?півтора$`)
	// numrToken.matches("(одн.+-)?півтори")
	numrPivtoryRE = regexp.MustCompile(`^(?:одн.+-)?півтори$`)
	// numrToken.matches("пів")
	numrPivRE = regexp.MustCompile(`^пів$`)
	// _FRACT = Pattern.compile(".*,[1-9]+")
	numrFractRE = regexp.MustCompile(`,[1-9]+$`)
)

func NewTokenAgreementNumrNounRule() *TokenAgreementNumrNounRule {
	return NewTokenAgreementNumrNounRuleWithMessages(nil)
}

// NewTokenAgreementNumrNounRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementNumrNounRuleWithMessages(messages map[string]string) *TokenAgreementNumrNounRule {
	r := &TokenAgreementNumrNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID: TokenAgreementNumrNounRuleID,
		// Java getDescription / getShort
		description:  "Узгодження відмінків, роду і числа числівника та іменника",
		shortMsg:     "Узгодження числівника та іменника",
		isLeftToken:  hasNumrReading,
		isRightToken: HasNounReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			if IsForceNounException(left, right) {
				return true
			}
			if IsFractionalNumrException(left, right) {
				return true
			}
			return NumrNounAgree(CollectPOSTags(left), CollectPOSTags(right))
		},
		exception: IsNumrNounException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

// IsForceNounException ports Java NOUN_FORCE_PATTERN.matcher(cleanTokenLower).matches().
func IsForceNounException(numr, noun *languagetool.AnalyzedTokenReadings) bool {
	if noun == nil {
		return false
	}
	// Java uses getCleanToken().toLowerCase() on the noun surface.
	clean := strings.ToLower(noun.GetCleanToken())
	if clean == "" {
		clean = strings.ToLower(noun.GetToken())
	}
	return nounForcePattern.MatchString(clean)
}

// IsFractionalNumrException ports Java half/fractional numeral surfaces used before
// special-case messaging (півтора / півтори / пів / decimal ,N). Simplified matcher
// soft-skips those pairs; does not invent lemma lists beyond Java String.matches.
func IsFractionalNumrException(numr, noun *languagetool.AnalyzedTokenReadings) bool {
	if numr == nil {
		return false
	}
	tok := strings.ToLower(numr.GetToken())
	if numrPivtoraRE.MatchString(tok) || numrPivtoryRE.MatchString(tok) ||
		numrPivRE.MatchString(tok) || numrFractRE.MatchString(tok) {
		return true
	}
	// also try clean token (Java sometimes uses cleanToken for numeric paths)
	clean := strings.ToLower(numr.GetCleanToken())
	if clean != "" && clean != tok {
		if numrPivtoraRE.MatchString(clean) || numrPivtoryRE.MatchString(clean) ||
			numrPivRE.MatchString(clean) || numrFractRE.MatchString(clean) {
			return true
		}
	}
	return false
}

func (r *TokenAgreementNumrNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
