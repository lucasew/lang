package de

import (
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanReadabilityRule ports org.languagetool.rules.de.GermanReadabilityRule
// (extends ReadabilityRule with German FRE and simpleSyllablesCount).
// Default off (Java defaultOn=false). Level default 3.
// Java base: Category TEXT_ANALYSIS, ITS Style.
type GermanReadabilityRule struct {
	Messages   map[string]string
	TooEasy    bool
	Level      int // threshold 0–6; Java default 3
	DefaultOff bool
	MinWords   int // Java MIN_WORDS = 10
	MarkWords  int // Java MARK_WORDS = 3 (span to mark)
	Category   *rules.Category
	IssueType  rules.ITSIssueType
}

func NewGermanReadabilityRule(messages map[string]string, tooEasy bool) *GermanReadabilityRule {
	cat := rules.NewCategoryFull(rules.NewCategoryId("TEXT_ANALYSIS"), "Text Analysis", rules.CategoryInternal, false, "")
	return &GermanReadabilityRule{
		Messages:   messages,
		TooEasy:    tooEasy,
		Level:      3,
		DefaultOff: true,
		MinWords:   10,
		MarkWords:  3,
		Category:   cat,
		IssueType:  rules.ITSStyle,
	}
}

func (r *GermanReadabilityRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *GermanReadabilityRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSStyle
	}
	return r.IssueType
}

func (r *GermanReadabilityRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

func (r *GermanReadabilityRule) GetID() string {
	if r != nil && r.TooEasy {
		return "READABILITY_RULE_SIMPLE_DE"
	}
	return "READABILITY_RULE_DIFFICULT_DE"
}

func (r *GermanReadabilityRule) GetDescription() string {
	if r != nil && r.TooEasy {
		return "Lesbarkeit: Zu einfacher Text"
	}
	return "Lesbarkeit: Zu schwieriger Text"
}

// German FRE (Amstad): 180 - ASL - 58.5 * ASW
func germanFleschReadingEase(asl, asw float64) float64 {
	return 180 - asl - (58.5 * asw)
}

// German readability level from FRE (Java ReadabilityRule.getReadabilityLevel).
// Higher level = easier text.
func germanReadabilityLevel(fre float64) int {
	switch {
	case fre < 30:
		return 0
	case fre < 50:
		return 1
	case fre < 60:
		return 2
	case fre < 70:
		return 3
	case fre < 80:
		return 4
	case fre < 90:
		return 5
	default:
		return 6
	}
}

func printMessageLevelDE(level int) string {
	var sLevel string
	switch level {
	case 0:
		sLevel = "Sehr schwer"
	case 1:
		sLevel = "Schwer"
	case 2:
		sLevel = "Mittelschwer"
	case 3:
		sLevel = "Mittel"
	case 4:
		sLevel = "Mittelleicht"
	case 5:
		sLevel = "Leicht"
	case 6:
		sLevel = "Sehr leicht"
	default:
		return ""
	}
	return " {Grad " + itoaDE(level) + ": " + sLevel + "}"
}

func (r *GermanReadabilityRule) getMessage(level int) string {
	simple, few := "schwierig", "viele"
	if r.TooEasy {
		simple, few = "einfach", "wenige"
	}
	return "Lesbarkeit: Der Text dieses Absatzes ist zu " + simple + printMessageLevelDE(level) +
		". Zu " + few + " Wörter pro Satz und zu " + few + " Silben pro Wort."
}

// simpleSyllablesCount ports GermanReadabilityRule.simpleSyllablesCount:
// word.charAt(i) / word.length() are UTF-16 units; GermanTools.isVowel(char).
func simpleSyllablesCountDE(word string) int {
	n := utf16LenDE(word)
	if n == 0 {
		return 0
	}
	nSyllables := 0
	if IsVowel(javaCharAtDE(word, 0)) {
		nSyllables++
	}
	lastDouble := false
	for i := 1; i < n; i++ {
		c := javaCharAtDE(word, i)
		if IsVowel(c) {
			cl := javaCharAtDE(word, i-1)
			if lastDouble {
				nSyllables++
				lastDouble = false
			} else if ((c == 'i' || c == 'y') && (cl == 'a' || cl == 'e' || cl == 'A' || cl == 'E')) ||
				(c == 'u' && (cl == 'a' || cl == 'e' || cl == 'o' || cl == 'A' || cl == 'E' || cl == 'O')) ||
				(c == 'e' && (cl == 'e' || cl == 'i' || cl == 'E' || cl == 'I')) ||
				(c == 'a' && (cl == 'a' || cl == 'A')) {
				lastDouble = true
			} else {
				nSyllables++
				lastDouble = false
			}
		} else {
			lastDouble = false
		}
	}
	if nSyllables == 0 {
		return 1
	}
	return nSyllables
}

// MatchList ports ReadabilityRule.match with German overrides.
// Processes paragraph ends via IsParagraphEnd; leftover trailing text forms last paragraph.
func (r *GermanReadabilityRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || len(sentences) == 0 {
		return nil
	}
	minWords := r.MinWords
	if minWords <= 0 {
		minWords = 10
	}
	markWords := r.MarkWords
	if markWords <= 0 {
		markWords = 3
	}
	levelThresh := r.Level
	if levelThresh < 0 {
		levelThresh = 3
	}

	var ruleMatches []*rules.RuleMatch
	nSentences, nWords, nSyllables := 0, 0, 0
	pos := 0
	startPos, endPos := -1, -1
	var markSentence *languagetool.AnalyzedSentence

	flushPara := func(sentence *languagetool.AnalyzedSentence) {
		if nWords >= minWords && nSentences > 0 {
			asl := float64(nWords) / float64(nSentences)
			asw := float64(nSyllables) / float64(nWords)
			fre := germanFleschReadingEase(asl, asw)
			rLevel := germanReadabilityLevel(fre)
			if (r.TooEasy && rLevel > levelThresh) || (!r.TooEasy && rLevel < levelThresh) {
				if startPos >= 0 && endPos >= 0 && markSentence != nil {
					// Java ReadabilityRule: new RuleMatch(this, sentence, start, end, msg) only.
					msg := r.getMessage(rLevel)
					rm := rules.NewRuleMatch(r, markSentence, startPos, endPos, msg)
					ruleMatches = append(ruleMatches, rm)
				}
			}
		}
		nSentences, nWords, nSyllables = 0, 0, 0
		startPos, endPos = -1, -1
		markSentence = nil
	}

	for n, sentence := range sentences {
		if sentence == nil {
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		if startPos < 0 && len(tokens) > 1 && tokens[1] != nil {
			startPos = pos + tokens[1].GetStartPos()
			markSentence = sentence
		}
		if endPos < 0 && len(tokens) > markWords && tokens[markWords] != nil {
			endPos = pos + tokens[markWords].GetEndPos()
		} else if endPos < 0 && len(tokens) > 1 {
			// short sentence: mark first few content tokens
			last := tokens[len(tokens)-1]
			if last != nil {
				endPos = pos + last.GetEndPos()
			}
		}
		nSentences++
		for _, token := range tokens {
			if token == nil || token.IsWhitespace() || token.IsNonWord() {
				continue
			}
			// skip sentence markers
			if token.IsSentenceStart() || token.IsSentenceEnd() {
				continue
			}
			sToken := token.GetToken()
			// skip pure non-letters for word count (Java isNonWord already covers punct)
			if sToken == "" {
				continue
			}
			allNonLetter := true
			for _, rr := range sToken {
				if unicode.IsLetter(rr) || unicode.IsDigit(rr) {
					allNonLetter = false
					break
				}
			}
			if allNonLetter {
				continue
			}
			nWords++
			nSyllables += simpleSyllablesCountDE(sToken)
		}
		// Java: Tools.isParagraphEnd
		if languagetool.IsParagraphEnd(sentences, n, false) {
			flushPara(sentence)
		}
		pos += sentence.GetCorrectedTextLength()
	}
	// trailing paragraph without explicit para end
	if nWords > 0 {
		flushPara(nil)
	}
	return ruleMatches
}
