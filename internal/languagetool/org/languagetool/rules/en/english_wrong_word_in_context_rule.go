package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wrongWordInContext.txt
var wrongWordInContextFS embed.FS

var (
	wwicOnce    sync.Once
	wwicEntries []rules.ContextWords
)

func loadWrongWordInContext() []rules.ContextWords {
	wwicOnce.Do(func() {
		f, err := wrongWordInContextFS.Open("data/wrongWordInContext.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		entries, err := rules.LoadWrongWordInContext(f)
		if err != nil {
			panic(err)
		}
		wwicEntries = entries
	})
	return wwicEntries
}

// EnglishWrongWordInContextRule ports org.languagetool.rules.en.EnglishWrongWordInContextRule.
type EnglishWrongWordInContextRule struct {
	*rules.WrongWordInContextRule
}

func NewEnglishWrongWordInContextRule(messages map[string]string) *EnglishWrongWordInContextRule {
	base := &rules.WrongWordInContextRule{
		Messages:           messages,
		ID:                 "ENGLISH_WRONG_WORD_IN_CONTEXT",
		Description:        "commonly confused words: $match",
		MessageString:      "Possibly confused word: Did you mean <suggestion>$SUGGESTION</suggestion> instead of '$WRONGWORD'?",
		ShortMessageString: "Possibly confused word",
		LongMessageString:  "Possibly confused word: Did you mean <suggestion>$SUGGESTION</suggestion> (= $EXPLANATION_SUGGESTION) instead of '$WRONGWORD' (= $EXPLANATION_WRONGWORD)?",
		LanguageCode:       "en",
		Entries:            loadWrongWordInContext(),
	}
	// Java: getCategoryString → "Commonly Confused Words"
	rules.InitWrongWordInContextMeta(base, messages, "Commonly Confused Words")
	// Java: addExamplePair(proscribed → prescribed)
	base.AddExamplePair(
		rules.Wrong("I have <marker>proscribed</marker> you a course of antibiotics."),
		rules.Fixed("I have <marker>prescribed</marker> you a course of antibiotics."),
	)
	return &EnglishWrongWordInContextRule{WrongWordInContextRule: base}
}

func (r *EnglishWrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WrongWordInContextRule.Match(sentence)
}
