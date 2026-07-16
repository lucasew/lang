package de

import (
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanReadabilityRule is a surface stand-in for GermanReadabilityRule using a
// simple Flesch-style estimate (words/sentences + syllables/word heuristic).
// TooEasy=true flags very easy paragraphs; false flags very difficult ones.
type GermanReadabilityRule struct {
	Messages     map[string]string
	TooEasy      bool
	Level        int // threshold 0–6; default 3
	MinSentences int
}

func NewGermanReadabilityRule(messages map[string]string, tooEasy bool) *GermanReadabilityRule {
	return &GermanReadabilityRule{
		Messages:     messages,
		TooEasy:      tooEasy,
		Level:        3,
		MinSentences: 2,
	}
}

func (r *GermanReadabilityRule) GetID() string {
	if r.TooEasy {
		return "READABILITY_RULE_SIMPLE_DE"
	}
	return "READABILITY_RULE_DIFFICULT_DE"
}

func countSyllablesDE(word string) int {
	// rough: count vowel groups äöüaeiouy
	n := 0
	inVowel := false
	for _, r := range word {
		rl := unicode.ToLower(r)
		vowel := rl == 'a' || rl == 'e' || rl == 'i' || rl == 'o' || rl == 'u' ||
			rl == 'y' || rl == 'ä' || rl == 'ö' || rl == 'ü'
		if vowel {
			if !inVowel {
				n++
			}
			inVowel = true
		} else {
			inVowel = false
		}
	}
	if n == 0 {
		return 1
	}
	return n
}

func (r *GermanReadabilityRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if len(sentences) < r.MinSentences {
		return nil
	}
	words, syllables := 0, 0
	for _, s := range sentences {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			w := tok.GetToken()
			if utf8.RuneCountInString(w) == 0 {
				continue
			}
			// skip pure punctuation
			allPunct := true
			for _, r0 := range w {
				if unicode.IsLetter(r0) || unicode.IsDigit(r0) {
					allPunct = false
					break
				}
			}
			if allPunct {
				continue
			}
			words++
			syllables += countSyllablesDE(w)
		}
	}
	if words == 0 {
		return nil
	}
	nSent := len(sentences)
	asl := float64(words) / float64(nSent)     // avg sentence length
	asw := float64(syllables) / float64(words) // avg syllables per word
	// German FRE approximation (Amstad): FRE = 180 - ASL - 58.5 * ASW
	fre := 180.0 - asl - 58.5*asw
	// map FRE to level 0–6 roughly
	level := 3
	switch {
	case fre < 0:
		level = 0
	case fre < 30:
		level = 1
	case fre < 50:
		level = 2
	case fre < 60:
		level = 3
	case fre < 70:
		level = 4
	case fre < 80:
		level = 5
	default:
		level = 6
	}
	thresh := r.Level
	if thresh < 0 {
		thresh = 3
	}
	// too easy: level > thresh; too difficult: level < thresh
	flag := false
	if r.TooEasy && level > thresh {
		flag = true
	}
	if !r.TooEasy && level < thresh {
		flag = true
	}
	if !flag {
		return nil
	}
	msg := "Lesbarkeit: Der Text dieses Absatzes ist zu schwierig."
	if r.TooEasy {
		msg = "Lesbarkeit: Der Text dieses Absatzes ist zu einfach."
	}
	// mark whole first sentence
	s0 := sentences[0]
	toks := s0.GetTokensWithoutWhitespace()
	if len(toks) < 2 {
		return nil
	}
	from := toks[1].GetStartPos()
	to := toks[len(toks)-1].GetEndPos()
	rm := rules.NewRuleMatch(r, s0, from, to, msg)
	rm.ShortMessage = "Lesbarkeit"
	return []*rules.RuleMatch{rm}
}
