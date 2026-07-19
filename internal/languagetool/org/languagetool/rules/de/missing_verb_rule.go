package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MissingVerbRule ports org.languagetool.rules.de.MissingVerbRule.
// Off by default in Java (setDefaultOff). Checks that a real sentence has a verb.
// Requires VER tags for a positive verbFound; untagged non-capitalized tokens
// also count as possible verbs (Java: !isTagged && !isCapitalizedWord).
// Sentence-start uppercase verb needs TagFirstLowercased (GermanTagger).
type MissingVerbRule struct {
	Messages map[string]string
	// Category ports setCategory(GRAMMAR).
	Category *rules.Category
	// DefaultOff mirrors Java setDefaultOff — registration MarkDefaultOff.
	DefaultOff bool
	// TagFirstLowercased ports verbAtSentenceStart: tag lowercased first word.
	// When nil, first-token uppercase workaround is skipped (fail-closed).
	TagFirstLowercased func(lower string) bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

const missingVerbMinTokens = 5

func NewMissingVerbRule(messages map[string]string) *MissingVerbRule {
	r := &MissingVerbRule{
		Messages:   messages,
		Category:   rules.CatGrammar.GetCategory(messages),
		DefaultOff: true,
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

// WithTagFirstLowercased sets the sentence-start re-tag hook.
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

func (r *MissingVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil || !isRealSentenceMissingVerb(sentence) || isSpecialCaseMissingVerb(sentence) {
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
		// Java operator precedence:
		// VER || ((!tagged && !capitalized) || (i==1 && verbAtSentenceStart))
		if readings.HasPosTagStartingWith("VER") ||
			(!readings.IsTagged() && !tools.IsCapitalizedWord(readings.GetToken())) ||
			(i == 1 && r.verbAtSentenceStart(readings)) {
			verbFound = true
			break
		}
		lastToken = readings
		i++
	}
	if !verbFound && lastToken != nil &&
		len(sentence.GetTokensWithoutWhitespace()) >= missingVerbMinTokens {
		// Java: new RuleMatch(..., msg) only — no shortMessage.
		msg := "Dieser Satz scheint kein Verb zu enthalten"
		rm := rules.NewRuleMatch(r, sentence, 0, lastToken.GetStartPos()+len(lastToken.GetToken()), msg)
		return []*rules.RuleMatch{rm}
	}
	return nil
}

// isRealSentenceMissingVerb ports MissingVerbRule.isRealSentence.
// Java: hasPosTag("PKT") && token is . ? !
func isRealSentenceMissingVerb(sentence *languagetool.AnalyzedSentence) bool {
	tokens := sentence.GetTokensWithoutWhitespace()
	if len(tokens) == 0 {
		return false
	}
	last := tokens[len(tokens)-1]
	if last == nil {
		return false
	}
	tok := last.GetToken()
	if tok != "." && tok != "?" && tok != "!" {
		return false
	}
	return last.HasPosTag("PKT")
}

// isSpecialCaseMissingVerb ports rule1/rule2: "Vielen Dank" / "Herzlichen Glückwunsch"
func isSpecialCaseMissingVerb(sentence *languagetool.AnalyzedSentence) bool {
	tokens := sentence.GetTokensWithoutWhitespace()
	start := 0
	if len(tokens) > 0 && tokens[0] != nil && tokens[0].IsSentenceStart() {
		start = 1
	}
	if start+1 >= len(tokens) {
		return false
	}
	a, b := tokens[start].GetToken(), tokens[start+1].GetToken()
	if a == "Vielen" && b == "Dank" {
		return true
	}
	if a == "Herzlichen" && b == "Glückwunsch" {
		return true
	}
	return false
}

func (r *MissingVerbRule) verbAtSentenceStart(readings *languagetool.AnalyzedTokenReadings) bool {
	if r == nil || r.TagFirstLowercased == nil || readings == nil {
		return false
	}
	tok := readings.GetToken()
	if tok == "" {
		return false
	}
	return r.TagFirstLowercased(tools.LowercaseFirstChar(tok))
}
