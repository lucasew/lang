package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MissingVerbRule ports org.languagetool.rules.de.MissingVerbRule.
// Off by default (setDefaultOff). Checks that a real sentence contains a verb.
// Ignores very short sentences and special cases "Vielen Dank" / "Herzlichen Glückwunsch".
type MissingVerbRule struct {
	Messages map[string]string
	// Category ports setCategory(GRAMMAR).
	Category *rules.Category
	// DefaultOff mirrors Java setDefaultOff — registration MarkDefaultOff.
	DefaultOff bool
	// rule1 / rule2 port PatternRule fields for isSpecialCase (case-sensitive tokens).
	rule1 *patterns.PatternRule
	rule2 *patterns.PatternRule
	// TagFirstLowercased ports verbAtSentenceStart via language.getTagger().tag:
	// lowercase first char, re-tag, hasPosTagStartingWith("VER").
	// When nil, sentence-start uppercase workaround is skipped (fail-closed).
	TagFirstLowercased func(lower string) bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// MIN_TOKENS_FOR_ERROR ports MissingVerbRule.MIN_TOKENS_FOR_ERROR.
const MIN_TOKENS_FOR_ERROR = 5

// NewMissingVerbRule ports MissingVerbRule(ResourceBundle messages, German language).
// Language is used in Java for PatternRule + tagger; PatternRules are built with "de"
// (same short-code family). Wire tagger via WithTagFirstLowercased / WireMissingVerbRule.
func NewMissingVerbRule(messages map[string]string) *MissingVerbRule {
	r := &MissingVerbRule{
		Messages:   messages,
		Category:   rules.CatGrammar.GetCategory(messages),
		DefaultOff: true,
		// Java:
		//   rule1 = new PatternRule("internal", language, Arrays.asList(
		//     new PatternToken("Vielen", true, false, false),
		//     new PatternToken("Dank", true, false, false)), "", "", "");
		//   rule2 = new PatternRule("internal", language, Arrays.asList(
		//     new PatternToken("Herzlichen", true, false, false),
		//     new PatternToken("Glückwunsch", true, false, false)), "", "", "");
		rule1: patterns.NewPatternRule("internal", "de", []*patterns.PatternToken{
			patterns.NewPatternToken("Vielen", true, false, false),
			patterns.NewPatternToken("Dank", true, false, false),
		}, "", "", ""),
		rule2: patterns.NewPatternRule("internal", "de", []*patterns.PatternToken{
			patterns.NewPatternToken("Herzlichen", true, false, false),
			patterns.NewPatternToken("Glückwunsch", true, false, false),
		}, "", "", ""),
	}
	// Java demo (fixed is illustrative, not a pure replacement of the marker).
	r.AddExamplePair(
		rules.Wrong("<marker>In diesem Satz kein Wort.</marker>"),
		rules.Fixed("In diesem Satz <marker>fehlt</marker> kein Wort."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *MissingVerbRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *MissingVerbRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *MissingVerbRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// WithTagFirstLowercased sets the sentence-start re-tag hook (language.getTagger().tag).
func (r *MissingVerbRule) WithTagFirstLowercased(fn func(lower string) bool) *MissingVerbRule {
	if r != nil {
		r.TagFirstLowercased = fn
	}
	return r
}

func (r *MissingVerbRule) GetID() string { return "MISSING_VERB" }

// GetDescription ports MissingVerbRule.getDescription.
func (r *MissingVerbRule) GetDescription() string { return "Satz ohne Verb" }

func (r *MissingVerbRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *MissingVerbRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// Match ports MissingVerbRule.match.
func (r *MissingVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil || !r.isRealSentence(sentence) || r.isSpecialCase(sentence) {
		return nil
	}
	verbFound := false
	var lastToken *languagetool.AnalyzedTokenReadings
	i := 0
	for _, readings := range sentence.GetTokensWithoutWhitespace() {
		if readings == nil {
			i++
			continue
		}
		// Java operator precedence (&& binds tighter than ||):
		//   hasPosTagStartingWith("VER")
		//   || (!isTagged() && !isCapitalizedWord(token)
		//       || (i == 1 && verbAtSentenceStart(readings)))
		// i.e. VER || (!tagged && !capitalized) || (i==1 && verbAtSentenceStart)
		if readings.HasPosTagStartingWith("VER") ||
			(!readings.IsTagged() && !tools.IsCapitalizedWord(readings.GetToken()) ||
				(i == 1 && r.verbAtSentenceStart(readings))) {
			verbFound = true
			break
		}
		lastToken = readings
		i++
	}
	if !verbFound && lastToken != nil &&
		len(sentence.GetTokensWithoutWhitespace()) >= MIN_TOKENS_FOR_ERROR {
		// Java: end = lastToken.getStartPos() + lastToken.getToken().length() (UTF-16)
		msg := "Dieser Satz scheint kein Verb zu enthalten"
		end := lastToken.GetStartPos() + utf16LenDE(lastToken.GetToken())
		rm := rules.NewRuleMatch(r, sentence, 0, end, msg)
		return []*rules.RuleMatch{rm}
	}
	return nil
}

// isRealSentence ports MissingVerbRule.isRealSentence.
// Java: last non-ws token hasPosTag("PKT") && token is . ? !
func (r *MissingVerbRule) isRealSentence(sentence *languagetool.AnalyzedSentence) bool {
	tokens := sentence.GetTokensWithoutWhitespace()
	if len(tokens) == 0 {
		return false
	}
	lastToken := tokens[len(tokens)-1]
	if lastToken == nil {
		return false
	}
	tok := lastToken.GetToken()
	return lastToken.HasPosTag("PKT") && (tok == "." || tok == "?" || tok == "!")
}

// isSpecialCase ports MissingVerbRule.isSpecialCase:
// rule1.match(sentence).length > 0 || rule2.match(sentence).length > 0
func (r *MissingVerbRule) isSpecialCase(sentence *languagetool.AnalyzedSentence) bool {
	if r == nil {
		return false
	}
	if r.rule1 != nil {
		if ms, err := r.rule1.Match(sentence); err == nil && len(ms) > 0 {
			return true
		}
	}
	if r.rule2 != nil {
		if ms, err := r.rule2.Match(sentence); err == nil && len(ms) > 0 {
			return true
		}
	}
	return false
}

// verbAtSentenceStart ports MissingVerbRule.verbAtSentenceStart.
// Start of sentence is mis-tagged because of uppercase first character; re-tag lowercased.
func (r *MissingVerbRule) verbAtSentenceStart(readings *languagetool.AnalyzedTokenReadings) bool {
	if r == nil || r.TagFirstLowercased == nil || readings == nil {
		return false
	}
	tok := readings.GetToken()
	if tok == "" {
		return false
	}
	// Java: StringTools.lowercaseFirstChar + language.getTagger().tag(singleton)
	return r.TagFirstLowercased(tools.LowercaseFirstChar(tok))
}
