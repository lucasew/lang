package en

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// L2ConfusionPair is an inject-friendly false-friend surface (full ngram deferred).
type L2ConfusionPair struct {
	// Wrong is the English surface often misused by L2 speakers.
	Wrong string
	// Better is the preferred English alternative.
	Better string
	// MotherGloss is the L1 meaning of Wrong (e.g. French "réaliser").
	MotherGloss string
	// MessageWord optional lemma for the detail template (defaults to Wrong).
	MessageWord string
}

// EnglishForL2SpeakersFalseFriendRule ports metadata for
// org.languagetool.rules.en.EnglishForL2SpeakersFalseFriendRule variants.
// Full ngram ConfusionProbabilityRule matching is deferred; Pairs inject greens Match.
type EnglishForL2SpeakersFalseFriendRule struct {
	ID           string
	MotherTongue string // short code, e.g. "de"
	Language     string // target language, e.g. "en"
	// Filenames are confusion set resources under the language resource dir.
	Filenames []string
	// ExampleWrong / ExampleFixed surface for documentation / tests (may include markers).
	ExampleWrong string
	ExampleFixed string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
	// Pairs optional inject map for unit tests (Java uses FakeLanguageModel + confusion file).
	Pairs []L2ConfusionPair
}

func (r *EnglishForL2SpeakersFalseFriendRule) GetID() string { return r.ID }

// AddExamplePair ports Rule.addExamplePair and sets ExampleWrong/Fixed surfaces.
func (r *EnglishForL2SpeakersFalseFriendRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
	r.ExampleWrong = incorrect.GetExample()
	r.ExampleFixed = correct.GetExample()
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *EnglishForL2SpeakersFalseFriendRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *EnglishForL2SpeakersFalseFriendRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}
func (r *EnglishForL2SpeakersFalseFriendRule) GetFilenames() []string {
	return append([]string(nil), r.Filenames...)
}

// MotherTongueName returns a display name for mother-tongue code (message surface).
func (r *EnglishForL2SpeakersFalseFriendRule) MotherTongueName() string {
	if r == nil {
		return ""
	}
	switch r.MotherTongue {
	case "de":
		return "German"
	case "fr":
		return "French"
	case "es":
		return "Spanish"
	case "nl":
		return "Dutch"
	default:
		return r.MotherTongue
	}
}

// LanguageName returns display name for the target language.
func (r *EnglishForL2SpeakersFalseFriendRule) LanguageName() string {
	if r == nil || r.Language == "" || r.Language == "en" {
		return "English"
	}
	return r.Language
}

// MessageFor ports the false-friend detail template:
// “$word1” ($L2) means “$word2” ($L1).
func (r *EnglishForL2SpeakersFalseFriendRule) MessageFor(enWord, motherGloss string) string {
	return fmt.Sprintf(`"%s" (%s) means "%s" (%s)`,
		enWord, r.LanguageName(), motherGloss, r.MotherTongueName())
}

// Match flags tokens present in Pairs (inject). Full LM ranking deferred.
func (r *EnglishForL2SpeakersFalseFriendRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || sentence == nil || len(r.Pairs) == 0 {
		return nil, nil
	}
	byWrong := map[string]L2ConfusionPair{}
	for _, p := range r.Pairs {
		if p.Wrong == "" {
			continue
		}
		byWrong[strings.ToLower(p.Wrong)] = p
	}
	var out []*rules.RuleMatch
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		// Skip pure SENT_START only — last content word carries SENT_END in LT.
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		w := tok.GetToken()
		p, ok := byWrong[strings.ToLower(w)]
		if !ok {
			continue
		}
		mw := p.MessageWord
		if mw == "" {
			mw = p.Wrong
		}
		msg := r.MessageFor(mw, p.MotherGloss)
		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
		if p.Better != "" {
			m.SetSuggestedReplacements([]string{p.Better})
		}
		out = append(out, m)
	}
	return out, nil
}

// NewEnglishForGermansFalseFriendRule ports EnglishForGermansFalseFriendRule metadata.
func NewEnglishForGermansFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	r := &EnglishForL2SpeakersFalseFriendRule{
		ID:           "EN_FOR_DE_SPEAKERS_FALSE_FRIENDS",
		MotherTongue: "de",
		Language:     "en",
		Filenames:    []string{"confusion_sets_l2_de.txt"},
	}
	// Java: My <marker>handy</marker> → phone
	r.AddExamplePair(
		rules.Wrong("My <marker>handy</marker> is broken."),
		rules.Fixed("My <marker>phone</marker> is broken."),
	)
	return r
}

// NewEnglishForFrenchFalseFriendRule ports EnglishForFrenchFalseFriendRule.
func NewEnglishForFrenchFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	r := &EnglishForL2SpeakersFalseFriendRule{
		ID:           "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS",
		MotherTongue: "fr",
		Language:     "en",
		Filenames:    []string{"confusion_sets_l2_fr.txt"},
	}
	// Java: achieve → complete
	r.AddExamplePair(
		rules.Wrong("She will <marker>achieve</marker> her task."),
		rules.Fixed("She will <marker>complete</marker> her task."),
	)
	return r
}

// NewEnglishForSpaniardsFalseFriendRule ports EnglishForSpaniardsFalseFriendRule.
func NewEnglishForSpaniardsFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	r := &EnglishForL2SpeakersFalseFriendRule{
		ID:           "EN_FOR_ES_SPEAKERS_FALSE_FRIENDS",
		MotherTongue: "es",
		Language:     "en",
		Filenames:    []string{"confusion_sets_l2_es.txt"},
	}
	// Java: realize → produce
	r.AddExamplePair(
		rules.Wrong("The factory will <marker>realize</marker> computer chips."),
		rules.Fixed("The factory will <marker>produce</marker> computer chips."),
	)
	return r
}

// NewEnglishForDutchmenFalseFriendRule ports EnglishForDutchmenFalseFriendRule.
func NewEnglishForDutchmenFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	r := &EnglishForL2SpeakersFalseFriendRule{
		ID:           "EN_FOR_NL_SPEAKERS_FALSE_FRIENDS",
		MotherTongue: "nl",
		Language:     "en",
		Filenames:    []string{"confusion_sets_l2_nl.txt"},
	}
	// Java: want → wall
	r.AddExamplePair(
		rules.Wrong("The <marker>want</marker> should be painted green."),
		rules.Fixed("The <marker>wall</marker> should be painted green."),
	)
	return r
}
