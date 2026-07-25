package de

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleRepeatedVeryShortSentences ports org.languagetool.rules.de.StyleRepeatedVeryShortSentences.
// TextLevelRule; CREATIVE_WRITING; setDefaultOff(); ITS Style.
//
// Java fields: minRepeated, minWords, excludeDirectSpeech from UserConfig triple
// (cf[0] Integer, cf[1] Integer, cf[2] Boolean); Tools.isParagraphEnd via lang.
type StyleRepeatedVeryShortSentences struct {
	Messages map[string]string
	// SingleLineBreaksMarksPara ports Language.getSentenceTokenizer().singleLineBreaksMarksPara()
	// (German SRX default false) for Tools.isParagraphEnd.
	SingleLineBreaksMarksPara bool
	MinWords                  int
	MinRepeated               int
	ExcludeDirectSpeech       bool
	Category                  *rules.Category
	IssueType                 rules.ITSIssueType
	DefaultOff                bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// Java private static final defaults.
const (
	styleShortMinRepeated         = 3
	styleShortMinWords            = 4
	styleShortExcludeDirectSpeech = true
)

// Java OPENING_QUOTES / ENDING_QUOTES.
var (
	styleShortOpenQuotes   = regexp.MustCompile(`[\"“„»«]`)
	styleShortEndingQuotes = regexp.MustCompile(`[\"“”»«]`)
)

// NewStyleRepeatedVeryShortSentences ports
// StyleRepeatedVeryShortSentences(messages, lang, null) with German para default.
func NewStyleRepeatedVeryShortSentences(messages map[string]string) *StyleRepeatedVeryShortSentences {
	return NewStyleRepeatedVeryShortSentencesWithUserConfig(messages, nil)
}

// NewStyleRepeatedVeryShortSentencesWithUserConfig ports
// StyleRepeatedVeryShortSentences(ResourceBundle messages, Language lang, UserConfig userConfig).
// lang.singleLineBreaksMarksPara is German-default false; set SingleLineBreaksMarksPara to override.
func NewStyleRepeatedVeryShortSentencesWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *StyleRepeatedVeryShortSentences {
	// Java: CREATIVE_WRITING category + setDefaultOff() + ITS Style.
	r := &StyleRepeatedVeryShortSentences{
		Messages:                  messages,
		SingleLineBreaksMarksPara: false, // German SRX default
		MinWords:                  styleShortMinWords,
		MinRepeated:               styleShortMinRepeated,
		ExcludeDirectSpeech:       styleShortExcludeDirectSpeech,
		Category:                  rules.CreativeWritingCategory(messages),
		IssueType:                 rules.ITSStyle,
		DefaultOff:                true,
	}
	// Java multi-marker staccato demo (fixed has no markers).
	r.AddExamplePair(
		rules.Wrong("Das Auto kam <marker>näher.</marker> Der Hund <marker>schlief.</marker> Die Reifen <marker>quietschten.</marker>"),
		rules.Fixed("Das Auto kam näher. Tief und fest schlief der Hund. Die Reifen quietschten."),
	)
	// Java: Object[] cf = userConfig.getConfigValueByID(getId());
	//   cf[0] instanceof Integer → minRepeated
	//   cf[1] instanceof Integer → minWords
	//   cf[2] instanceof Boolean → excludeDirectSpeech
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID(r.GetID()); cf != nil {
			if len(cf) > 0 && cf[0] != nil {
				if n, ok := cf[0].(int); ok {
					r.MinRepeated = n
				}
			}
			if len(cf) > 1 && cf[1] != nil {
				if n, ok := cf[1].(int); ok {
					r.MinWords = n
				}
			}
			if len(cf) > 2 && cf[2] != nil {
				if b, ok := cf[2].(bool); ok {
					r.ExcludeDirectSpeech = b
				}
			}
		}
	}
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

// MinToCheckParagraph ports minToCheckParagraph (Java returns minRepeated as configured).
func (r *StyleRepeatedVeryShortSentences) MinToCheckParagraph() int {
	if r == nil {
		return styleShortMinRepeated
	}
	return r.MinRepeated
}

// GetRuleOptions ports StyleRepeatedVeryShortSentences.getRuleOptions (3 options).
// Java:
//
//	new RuleOption(MIN_REPEATED, "Anzahl von kurzen Sätzen hintereinander", 0, 10),
//	new RuleOption(MIN_WORDS, "Maximale Länge eines kurzen Satzes", 0, 10),
//	new RuleOption(EXCLUDE_DIRECT_SPEECH, "Direkte Rede und Zitate ausschließen")
func (r *StyleRepeatedVeryShortSentences) GetRuleOptions() []*rules.RuleOption {
	return []*rules.RuleOption{
		rules.NewRuleOption(styleShortMinRepeated, "Anzahl von kurzen Sätzen hintereinander", 0, 10),
		rules.NewRuleOption(styleShortMinWords, "Maximale Länge eines kurzen Satzes", 0, 10),
		rules.NewRuleOption(styleShortExcludeDirectSpeech, "Direkte Rede und Zitate ausschließen", nil, nil),
	}
}

// MatchList ports StyleRepeatedVeryShortSentences.match(List<AnalyzedSentence>).
//
// Java control flow (bug-for-bug):
//  1. if sentences.size() < minRepeated → empty
//  2. per sentence: direct-speech OPENING/ENDING quotes with whitespaceBefore guards
//  3. paragraphEnd = Tools.isParagraphEnd(sentences, n, lang)
//  4. include short when !beginnsWithDirectSpeech && (!paragraphEnd || nPara > 0)
//     && tokens.length > 3 && tokens.length <= minWords + 2
//     span: tokens[len-2].startPos+pos .. tokens[len-1].endPos+pos
//  5. else flush if nRepeated >= minRepeated (message = getDescription())
//  6. pos += getCorrectedTextLength(); if paragraphEnd nPara = -1
//  7. final flush after loop
func (r *StyleRepeatedVeryShortSentences) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	minW := r.MinWords
	minR := r.MinRepeated
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
	singleLine := r.SingleLineBreaksMarksPara
	for n, sentence := range sentences {
		nPara++
		if sentence == nil {
			flush()
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		if r.ExcludeDirectSpeech {
			if endsWithDirectSpeech {
				beginsWithDirectSpeech = true
			} else {
				beginsWithDirectSpeech = false
			}
			for i := 0; i < len(tokens); i++ {
				if tokens[i] == nil {
					continue
				}
				tok := tokens[i].GetToken()
				// Java: !beginnsWithDirectSpeech && OPENING_QUOTES.matches(token)
				//   && i < tokens.length - 1 && !tokens[i + 1].isWhitespaceBefore()
				if !beginsWithDirectSpeech && styleShortOpenQuotes.MatchString(tok) &&
					i < len(tokens)-1 && tokens[i+1] != nil && !tokens[i+1].IsWhitespaceBefore() {
					beginsWithDirectSpeech = true
					endsWithDirectSpeech = true
				} else if beginsWithDirectSpeech && styleShortEndingQuotes.MatchString(tok) &&
					// Java: i > 1 && !tokens[i].isWhitespaceBefore()
					i > 1 && !tokens[i].IsWhitespaceBefore() {
					endsWithDirectSpeech = false
				}
			}
		}
		// Java: Tools.isParagraphEnd(sentences, n, lang)
		paragraphEnd := languagetool.IsParagraphEnd(sentences, n, singleLine)
		// Java: !beginnsWithDirectSpeech && (!paragraphEnd || nPara > 0)
		//   && tokens.length > 3 && tokens.length <= minWords + 2
		// 3 → skip one-word sentences; max = minWords + SENT_START + punctuation
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
