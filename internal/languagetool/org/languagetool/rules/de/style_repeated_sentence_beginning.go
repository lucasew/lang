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
	MinRepeated int // Java MIN_REPEATED = 3
	Category    *rules.Category
	IssueType   rules.ITSIssueType
	DefaultOff  bool
}

func NewStyleRepeatedSentenceBeginning(messages map[string]string) *StyleRepeatedSentenceBeginning {
	// Java: CREATIVE_WRITING category + setDefaultOff() + ITS Style.
	return &StyleRepeatedSentenceBeginning{
		Messages:    messages,
		MinRepeated: 3,
		Category:    rules.CreativeWritingCategory(messages),
		IssueType:   rules.ITSStyle,
		DefaultOff:  true,
	}
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

// MinToCheckParagraph ports minToCheckParagraph (Java returns MIN_REPEATED).
func (r *StyleRepeatedSentenceBeginning) MinToCheckParagraph() int {
	if r == nil || r.MinRepeated <= 0 {
		return 3
	}
	return r.MinRepeated
}

func (r *StyleRepeatedSentenceBeginning) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
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
	var startPos, endPos []int
	var repeated []*languagetool.AnalyzedSentence
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
	for _, sentence := range sentences {
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
		isArt := t1.HasPosTagStartingWith("ART:DEF:NOM") || t1.HasPosTagStartingWith("ART:IND:NOM")
		isPro := t1.HasPosTagStartingWith("PRO:PER:NOM")
		if isArt {
			end := t1.GetEndPos() + pos
			// Java: scan for SUB before VER; else end at article
			noSub := true
			for i := 2; i < len(tokens) && tokens[i] != nil && !tokens[i].HasPosTagStartingWith("VER"); i++ {
				if tokens[i].HasPosTagStartingWith("SUB") {
					noSub = false
					end = tokens[i].GetEndPos() + pos
					break
				}
			}
			if noSub {
				end = t1.GetEndPos() + pos
			}
			repeated = append(repeated, sentence)
			startPos = append(startPos, t1.GetStartPos()+pos)
			endPos = append(endPos, end)
			nRepeated++
		} else if isPro {
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
