package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanStyleRepeatedWordRule is a surface stand-in for
// org.languagetool.rules.de.GermanStyleRepeatedWordRule.
// Without POS/lemmas, content-like words (letters, len≥4, not stopwords)
// are matched by exact surface form across the same or adjacent sentences.
type GermanStyleRepeatedWordRule struct {
	Messages               map[string]string
	MaxDistanceOfSentences int
}

func NewGermanStyleRepeatedWordRule(messages map[string]string) *GermanStyleRepeatedWordRule {
	return &GermanStyleRepeatedWordRule{
		Messages:               messages,
		MaxDistanceOfSentences: 1,
	}
}

func (r *GermanStyleRepeatedWordRule) GetID() string { return "STYLE_REPEATED_WORD_RULE_DE" }

// German function words / short closed class to reduce false positives.
var styleRepeatStop = map[string]struct{}{
	"der": {}, "die": {}, "das": {}, "den": {}, "dem": {}, "des": {},
	"ein": {}, "eine": {}, "einen": {}, "einem": {}, "einer": {}, "eines": {},
	"und": {}, "oder": {}, "aber": {}, "denn": {}, "weil": {}, "dass": {}, "daß": {},
	"sich": {}, "nicht": {}, "noch": {}, "auch": {}, "nur": {}, "schon": {},
	"mit": {}, "von": {}, "zu": {}, "zum": {}, "zur": {}, "bei": {}, "nach": {},
	"aus": {}, "auf": {}, "für": {}, "über": {}, "unter": {}, "vor": {}, "hinter": {},
	"ich": {}, "du": {}, "er": {}, "sie": {}, "es": {}, "wir": {}, "ihr": {},
	"mir": {}, "dir": {}, "ihm": {}, "uns": {}, "euch": {},
	"mein": {}, "dein": {}, "sein": {}, "unser": {}, "euer": {},
	"ist": {}, "sind": {}, "war": {}, "waren": {}, "hat": {}, "haben": {},
	"wird": {}, "werden": {}, "kann": {}, "muss": {}, "soll": {}, "will": {},
	"als": {}, "wie": {}, "wenn": {}, "dann": {}, "so": {}, "am": {}, "im": {},
	"in": {}, "an": {}, "um": {}, "ob": {}, "was": {}, "wer": {}, "wo": {},
	"all": {}, "alle": {}, "allem": {}, "allen": {}, "aller": {}, "alles": {},
	"diese": {}, "dieser": {}, "dieses": {}, "diesen": {}, "diesem": {},
	"jene": {}, "jener": {}, "kein": {}, "keine": {}, "keinen": {},
	"man": {}, "doch": {}, "sehr": {}, "mehr": {}, "hier": {},
	"dort": {}, "da": {}, "nun": {}, "also": {}, "etwa": {}, "etwas": {},
}

func isStyleContentWord(tok string) bool {
	if utf8.RuneCountInString(tok) < 4 {
		return false
	}
	lc := strings.ToLower(tok)
	if _, stop := styleRepeatStop[lc]; stop {
		return false
	}
	for _, r := range tok {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func hasBreakTokenStyle(tokens []*languagetool.AnalyzedTokenReadings) bool {
	n := len(tokens)
	if n > 5 {
		n = 5
	}
	for i := 0; i < n; i++ {
		t := tokens[i].GetToken()
		if t == "-" || t == "—" || t == "–" {
			return true
		}
	}
	return false
}

func (r *GermanStyleRepeatedWordRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	maxDist := r.MaxDistanceOfSentences
	if maxDist < 0 {
		maxDist = 1
	}
	// Precompute lowercase content tokens per sentence
	type sentInfo struct {
		tokens []*languagetool.AnalyzedTokenReadings
		// lc form → indices of content tokens
		forms map[string][]int
	}
	infos := make([]sentInfo, len(sentences))
	for si, s := range sentences {
		toks := s.GetTokensWithoutWhitespace()
		forms := map[string][]int{}
		for i := 1; i < len(toks); i++ {
			w := toks[i].GetToken()
			if !isStyleContentWord(w) {
				continue
			}
			lc := strings.ToLower(w)
			forms[lc] = append(forms[lc], i)
		}
		infos[si] = sentInfo{tokens: toks, forms: forms}
	}

	var matches []*rules.RuleMatch
	pos := 0
	for n, s := range sentences {
		info := infos[n]
		if hasBreakTokenStyle(info.tokens) {
			pos += s.GetCorrectedTextLength()
			continue
		}
		// For each content token, see if form repeats in window
		seenInSent := map[string]int{} // first index in this sentence
		for i := 1; i < len(info.tokens); i++ {
			tok := info.tokens[i]
			w := tok.GetToken()
			if !isStyleContentWord(w) {
				continue
			}
			lc := strings.ToLower(w)
			isRepeated := 0 // 1 same, 2 before, 3 after
			// same sentence: another occurrence
			if idxs := info.forms[lc]; len(idxs) > 1 {
				// ignore immediate adjacent identical (Java special-case)
				for _, j := range idxs {
					if j == i {
						continue
					}
					if j == i-1 || j == i+1 {
						continue
					}
					isRepeated = 1
					break
				}
			}
			for d := 1; isRepeated == 0 && d <= maxDist; d++ {
				if n-d >= 0 {
					if _, ok := infos[n-d].forms[lc]; ok {
						isRepeated = 2
					}
				}
			}
			for d := 1; isRepeated == 0 && d <= maxDist; d++ {
				if n+d < len(infos) {
					if _, ok := infos[n+d].forms[lc]; ok {
						isRepeated = 3
					}
				}
			}
			_ = seenInSent
			if isRepeated == 0 {
				continue
			}
			var msg string
			switch isRepeated {
			case 1:
				msg = "Mögliches Stilproblem: Das Wort wird noch einmal im selben Satz verwendet."
			case 2:
				msg = "Mögliches Stilproblem: Das Wort wird bereits in einem vorhergehenden Satz verwendet."
			default:
				msg = "Mögliches Stilproblem: Das Wort wird auch in einem nachfolgenden Satz verwendet."
			}
			rm := rules.NewRuleMatch(r, s, pos+tok.GetStartPos(), pos+tok.GetEndPos(), msg)
			rm.ShortMessage = "Wiederholtes Wort"
			matches = append(matches, rm)
		}
		pos += s.GetCorrectedTextLength()
	}
	return matches
}
