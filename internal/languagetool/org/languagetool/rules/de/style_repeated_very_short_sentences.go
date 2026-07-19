package de

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleRepeatedVeryShortSentences ports org.languagetool.rules.de.StyleRepeatedVeryShortSentences.
// Direct-speech exclusion and paragraph-end handling follow Java (default excludeDirectSpeech=true).
// Java: Category CREATIVE_WRITING, setDefaultOff(), ITS Style.
type StyleRepeatedVeryShortSentences struct {
	Messages            map[string]string
	MinWords            int
	MinRepeated         int
	ExcludeDirectSpeech bool
	Category            *rules.Category
	IssueType           rules.ITSIssueType
	DefaultOff          bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

var (
	styleShortOpenQuotes   = regexp.MustCompile(`[\"“„»«]`)
	styleShortEndingQuotes = regexp.MustCompile(`[\"“”»«]`)
)

func NewStyleRepeatedVeryShortSentences(messages map[string]string) *StyleRepeatedVeryShortSentences {
	// Java: CREATIVE_WRITING category + setDefaultOff() + ITS Style.
	r := &StyleRepeatedVeryShortSentences{
		Messages:            messages,
		MinWords:            4,
		MinRepeated:         3,
		ExcludeDirectSpeech: true, // Java EXCLUDE_DIRECT_SPEECH default
		Category:            rules.CreativeWritingCategory(messages),
		IssueType:           rules.ITSStyle,
		DefaultOff:          true,
	}
	// Java multi-marker staccato demo (fixed has no markers).
	r.AddExamplePair(
		rules.Wrong("Das Auto kam <marker>näher.</marker> Der Hund <marker>schlief.</marker> Die Reifen <marker>quietschten.</marker>"),
		rules.Fixed("Das Auto kam näher. Tief und fest schlief der Hund. Die Reifen quietschten."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *StyleRepeatedVeryShortSentences) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *StyleRepeatedVeryShortSentences) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *StyleRepeatedVeryShortSentences) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *StyleRepeatedVeryShortSentences) GetID() string { return "STYLE_REPEATED_SHORT_SENTENCES" }

// GetDescription ports StyleRepeatedVeryShortSentences.getDescription.
func (r *StyleRepeatedVeryShortSentences) GetDescription() string { return "Stakkato-Sätze" }

func (r *StyleRepeatedVeryShortSentences) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *StyleRepeatedVeryShortSentences) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSStyle
	}
	return r.IssueType
}

func (r *StyleRepeatedVeryShortSentences) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// MinToCheckParagraph ports minToCheckParagraph (Java returns minRepeated).
func (r *StyleRepeatedVeryShortSentences) MinToCheckParagraph() int {
	if r == nil || r.MinRepeated <= 0 {
		return 3
	}
	return r.MinRepeated
}

func (r *StyleRepeatedVeryShortSentences) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	minW := r.MinWords
	if minW <= 0 {
		minW = 4
	}
	minR := r.MinRepeated
	if minR <= 0 {
		minR = 3
	}
	if len(sentences) < minR {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	pos := 0
	nRepeated := 0
	nPara := -1
	var startPos, endPos []int
	var repeated []*languagetool.AnalyzedSentence
	beginsWithDirectSpeech := false
	endsWithDirectSpeech := false
	flush := func() {
		if nRepeated >= minR {
			// Java: new RuleMatch(..., getDescription()) — no shortMessage.
			msg := r.GetDescription()
			for i := 0; i < len(repeated); i++ {
				rm := rules.NewRuleMatch(r, repeated[i], startPos[i], endPos[i], msg)
				ruleMatches = append(ruleMatches, rm)
			}
		}
		repeated = nil
		startPos = nil
		endPos = nil
		nRepeated = 0
	}
	for n, sentence := range sentences {
		nPara++
		tokens := sentence.GetTokensWithoutWhitespace()
		if r.ExcludeDirectSpeech {
			if endsWithDirectSpeech {
				beginsWithDirectSpeech = true
			} else {
				beginsWithDirectSpeech = false
			}
			for i := 0; i < len(tokens); i++ {
				tok := tokens[i].GetToken()
				if !beginsWithDirectSpeech && styleShortOpenQuotes.MatchString(tok) &&
					i < len(tokens)-1 && !tokens[i+1].IsWhitespaceBefore() {
					beginsWithDirectSpeech = true
					endsWithDirectSpeech = true
				} else if beginsWithDirectSpeech && styleShortEndingQuotes.MatchString(tok) &&
					i > 1 && !tokens[i].IsWhitespaceBefore() {
					endsWithDirectSpeech = false
				}
			}
		}
		// Java: Tools.isParagraphEnd(sentences, n, lang) — singleLineBreaksMarksPara false for German.
		paragraphEnd := languagetool.IsParagraphEnd(sentences, n, false)
		// Java: !beginnsWithDirectSpeech && (!paragraphEnd || nPara > 0) && tokens.length > 3 && tokens.length <= minWords + 2
		if !beginsWithDirectSpeech && (!paragraphEnd || nPara > 0) && len(tokens) > 3 && len(tokens) <= minW+2 {
			repeated = append(repeated, sentence)
			from := tokens[len(tokens)-2].GetStartPos() + pos
			to := tokens[len(tokens)-1].GetEndPos() + pos
			startPos = append(startPos, from)
			endPos = append(endPos, to)
			nRepeated++
		} else {
			flush()
		}
		pos += sentence.GetCorrectedTextLength()
		if paragraphEnd {
			nPara = -1
		}
	}
	flush()
	return ruleMatches
}
