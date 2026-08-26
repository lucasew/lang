package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UpperCaseNgramRule ports org.languagetool.rules.de.UpperCaseNgramRule.
// Finds some words written uppercase that should be lowercase and vice versa.
// @since 5.0
type UpperCaseNgramRule struct {
	messages map[string]string
	// lm ports LanguageModel (Java Objects.requireNonNull in ctor).
	lm ngrams.LanguageModel
	// langCode ports Language lang ctor arg (currently unused in Java Match;
	// reserved for the commented ambiguous_case.txt extension).
	langCode string

	Category  *rules.Category
	IssueType rules.ITSIssueType
	// DefaultOff / DefaultTempOff port setDefaultTempOff() (Java TODO still on).
	DefaultOff     bool
	DefaultTempOff bool

	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// upperCaseNgramThreshold ports UpperCaseNgramRule.THRESHOLD.
const upperCaseNgramThreshold = 50

// upperCaseNgramRelevantWords ports UpperCaseNgramRule.relevantWords.
// Java TODO: optional /de/ambiguous_case.txt load stays disabled.
var upperCaseNgramRelevantWords = map[string]struct{}{
	"tage": {}, "tagen": {},
	"Tage": {}, "Tagen": {},
}

// NewUpperCaseNgramRule ports
// UpperCaseNgramRule(ResourceBundle messages, LanguageModel lm, Language lang).
// lm must be non-nil (Java Objects.requireNonNull).
func NewUpperCaseNgramRule(messages map[string]string, lm ngrams.LanguageModel, langCode string) *UpperCaseNgramRule {
	if lm == nil {
		panic("lm must not be null") // Java Objects.requireNonNull(lm)
	}
	/*
		TODO: potential extension - too many false alarms for now...
		String path = "/de/ambiguous_case.txt";
		List<String> lines = JLanguageTool.getDataBroker().getFromResourceDirAsLines(path);
		for (String line : lines) {
		  if (line.isEmpty() || line.startsWith("#")) {
		    continue;
		  }
		  if (startsWithLowercase(line)) {
		    throw new RuntimeException("Only entries with uppercase first chars expected in " + path + ": " + line);
		  }
		  relevantWords.add(line);
		  relevantWords.add(lowercaseFirstChar(line));
		}
	*/
	r := &UpperCaseNgramRule{
		messages:       messages,
		lm:             lm,
		langCode:       langCode,
		Category:       rules.CatCasing.GetCategory(messages),
		IssueType:      rules.ITSMisspelling,
		DefaultOff:     true, // setDefaultTempOff
		DefaultTempOff: true,
	}
	r.AddExamplePair(
		rules.Wrong("Die Suche endete nach 15 <marker>tagen</marker>."),
		rules.Fixed("Die Suche endete nach 15 <marker>Tagen</marker>."),
	)
	return r
}

func (r *UpperCaseNgramRule) GetID() string { return "DE_UPPER_CASE_NGRAM" }

func (r *UpperCaseNgramRule) GetDescription() string {
	return "Prüft Wörter, ob sie fälschlich groß- oder fälschlich kleingeschrieben sind"
}

func (r *UpperCaseNgramRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *UpperCaseNgramRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSMisspelling
	}
	return r.IssueType
}

// IsDefaultTempOff ports Rule.isDefaultTempOff (set in ctor via setDefaultTempOff).
func (r *UpperCaseNgramRule) IsDefaultTempOff() bool {
	return r != nil && r.DefaultTempOff
}

// IsDefaultOff ports Rule.isDefaultOff (setDefaultTempOff also sets defaultOff).
func (r *UpperCaseNgramRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// AddExamplePair ports Rule.addExamplePair.
func (r *UpperCaseNgramRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *UpperCaseNgramRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *UpperCaseNgramRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// GetAntiPatterns ports Rule.getAntiPatterns default (empty).
// Used by getSentenceWithImmunization identity path in Match.
func (r *UpperCaseNgramRule) GetAntiPatterns() []rules.SentenceReplacer {
	return nil
}

// Match ports UpperCaseNgramRule.match.
func (r *UpperCaseNgramRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	// Java: AnalyzedTokenReadings[] tokens = getSentenceWithImmunization(sentence).getTokensWithoutWhitespace();
	// DE rule does not override getAntiPatterns → identity immunization.
	work := rules.SentenceWithImmunization(sentence, r.GetAntiPatterns())
	tokens := work.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		if token == nil {
			continue
		}
		tokenStr := token.GetToken()
		if i+1 < len(tokens) && upperCaseNgramRelevantContains(tokenStr) && !tools.IsAllUppercase(tokenStr) {
			ucToken := tools.UppercaseFirstChar(tokenStr)
			lcToken := tools.LowercaseFirstChar(tokenStr)
			prevTok := tokens[i-1]
			nextTok := tokens[i+1]
			if prevTok == nil || nextTok == nil {
				continue
			}
			ucList := []string{prevTok.GetToken(), ucToken, nextTok.GetToken()}
			lcList := []string{prevTok.GetToken(), lcToken, nextTok.GetToken()}
			ucProb := r.lm.GetPseudoProbability(ucList)
			lcProb := r.lm.GetPseudoProbability(lcList)
			if tools.StartsWithUppercase(tokenStr) {
				ratio := lcProb.GetProb() / ucProb.GetProb()
				if ratio > upperCaseNgramThreshold {
					msg := "Meinten Sie das Verb '" + lcToken + "'? Nur Nomen und Eigennamen werden großgeschrieben."
					match := rules.NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), msg)
					match.SetSuggestedReplacement(lcToken)
					matches = append(matches, match)
				}
			} else {
				ratio := ucProb.GetProb() / lcProb.GetProb()
				if ratio > upperCaseNgramThreshold {
					msg := "Meinten Sie das Nomen '" + ucToken + "'? Nomen und Eigennamen werden großgeschrieben."
					match := rules.NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), msg)
					match.SetSuggestedReplacement(ucToken)
					matches = append(matches, match)
				}
			}
		}
	}
	return matches
}

func upperCaseNgramRelevantContains(tokenStr string) bool {
	_, ok := upperCaseNgramRelevantWords[tokenStr]
	return ok
}
