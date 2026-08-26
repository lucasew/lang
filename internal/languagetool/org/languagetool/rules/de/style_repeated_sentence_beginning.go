package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleRepeatedSentenceBeginning ports
// org.languagetool.rules.de.StyleRepeatedSentenceBeginning (default off).
// Java: ART:DEF:NOM / ART:IND:NOM or PRO:PER:NOM at tokens[1] only — no surface invent.
// Java: Category CREATIVE_WRITING, setDefaultOff(), ITS Style.
type StyleRepeatedSentenceBeginning struct {
	Messages    map[string]string
	MinRepeated int // Java private static final MIN_REPEATED = 3
	Category    *rules.Category
	IssueType   rules.ITSIssueType
	DefaultOff  bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewStyleRepeatedSentenceBeginning(messages map[string]string) *StyleRepeatedSentenceBeginning {
	// Java ctor:
	//   super.setCategory(new Category(new CategoryId("CREATIVE_WRITING"),
	//       messages.getString("category_creative_writing"), Location.INTERNAL, false));
	//   setDefaultOff();
	//   setLocQualityIssueType(ITSIssueType.Style);
	//   addExamplePair(Example.wrong(... multi-marker ...), Example.fixed(...));
	r := &StyleRepeatedSentenceBeginning{
		Messages:    messages,
		MinRepeated: 3,
		Category:    rules.CreativeWritingCategory(messages),
		IssueType:   rules.ITSStyle,
		DefaultOff:  true,
	}
	r.AddExamplePair(
		rules.Wrong("<marker>Das Auto</marker> kam näher. <marker>Der Hund</marker> lief langsam über die Straße. <marker>Die Reifen</marker> quietschten."),
		rules.Fixed("Das Auto kam näher. Langsam lief der Hund über die Straße. Die Reifen quietschten."),
	)
	return r
}

func (r *StyleRepeatedSentenceBeginning) GetID() string {
	return "STYLE_REPEATED_SENTENCE_BEGINNING"
}

func (r *StyleRepeatedSentenceBeginning) GetDescription() string {
	return "Subjekt als wiederholter Satzanfang"
}

func (r *StyleRepeatedSentenceBeginning) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *StyleRepeatedSentenceBeginning) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSStyle
	}
	return r.IssueType
}

func (r *StyleRepeatedSentenceBeginning) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// AddExamplePair ports Rule.addExamplePair.
func (r *StyleRepeatedSentenceBeginning) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *StyleRepeatedSentenceBeginning) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *StyleRepeatedSentenceBeginning) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// MinToCheckParagraph ports minToCheckParagraph (Java returns MIN_REPEATED).
func (r *StyleRepeatedSentenceBeginning) MinToCheckParagraph() int {
	if r == nil || r.MinRepeated <= 0 {
		return 3
	}
	return r.MinRepeated
}

// MatchList ports StyleRepeatedSentenceBeginning.match(List<AnalyzedSentence>).
//
// Java control flow (bug-for-bug):
//  1. if sentences.size() < MIN_REPEATED → empty
//  2. tokens = getTokensWithoutWhitespace(); inspect tokens[1] only
//  3. ART:DEF:NOM or ART:IND:NOM → scan i=2.. while !VER for SUB; end at SUB or article (noSub)
//  4. else PRO:PER:NOM → span tokens[1] only
//  5. else if nRepeated >= MIN_REPEATED emit RuleMatch(getDescription()) for each; clear
//  6. pos += getCorrectedTextLength() every sentence; flush again after loop
func (r *StyleRepeatedSentenceBeginning) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	minR := 3
	if r != nil && r.MinRepeated > 0 {
		minR = r.MinRepeated
	}
	if len(sentences) < minR {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	pos := 0
	nRepeated := 0
	var startPos, endPos []int
	var repeated []*languagetool.AnalyzedSentence
	flush := func() {
		// Java: if (nRepeated >= MIN_REPEATED) { for ... new RuleMatch(this, sent, start, end, getDescription()) }
		if nRepeated >= minR {
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
	for _, sentence := range sentences {
		// Java assumes non-null sentences with tokens[1]; defensive guards map to the non-match branch.
		if sentence == nil {
			flush()
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		if len(tokens) < 2 || tokens[1] == nil {
			flush()
			pos += sentence.GetCorrectedTextLength()
			continue
		}
		t1 := tokens[1]
		// Java: hasPosTagStartingWith("ART:DEF:NOM") || hasPosTagStartingWith("ART:IND:NOM")
		if t1.HasPosTagStartingWith("ART:DEF:NOM") || t1.HasPosTagStartingWith("ART:IND:NOM") {
			noSub := true
			// Java: for (i = 2; i < tokens.length && !tokens[i].hasPosTagStartingWith("VER"); i++)
			for i := 2; i < len(tokens) && tokens[i] != nil && !tokens[i].HasPosTagStartingWith("VER"); i++ {
				if tokens[i].HasPosTagStartingWith("SUB") {
					noSub = false
					endPos = append(endPos, tokens[i].GetEndPos()+pos)
					break
				}
			}
			if noSub {
				endPos = append(endPos, t1.GetEndPos()+pos)
			}
			repeated = append(repeated, sentence)
			startPos = append(startPos, t1.GetStartPos()+pos)
			nRepeated++
		} else if t1.HasPosTagStartingWith("PRO:PER:NOM") {
			// Java: start/end = tokens[1] only
			repeated = append(repeated, sentence)
			startPos = append(startPos, t1.GetStartPos()+pos)
			endPos = append(endPos, t1.GetEndPos()+pos)
			nRepeated++
		} else {
			flush()
		}
		pos += sentence.GetCorrectedTextLength()
	}
	flush()
	return ruleMatches
}
