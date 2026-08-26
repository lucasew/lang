package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SimilarNameRule ports org.languagetool.rules.de.SimilarNameRule.
// Search for names similar to ones used before in the same text (possible typos).
// Java: Categories.TYPOS, setDefaultOff(), TextLevelRule, minToCheckParagraph=-1.
type SimilarNameRule struct {
	Messages   map[string]string
	Category   *rules.Category
	DefaultOff bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// Java: private static final int MIN_LENGTH = 4; MAX_DIFF = 1;
const (
	similarNameMinLen  = 4
	similarNameMaxDiff = 1
)

func NewSimilarNameRule(messages map[string]string) *SimilarNameRule {
	r := &SimilarNameRule{
		Messages:   messages,
		Category:   rules.CatTypos.GetCategory(messages),
		DefaultOff: true, // Java setDefaultOff()
	}
	// Java: addExamplePair(Example.wrong(...Miller...), Example.fixed(...Müller...))
	r.AddExamplePair(
		rules.Wrong("Angela Müller ist CEO. <marker>Miller</marker> wurde in Hamburg geboren."),
		rules.Fixed("Angela Müller ist CEO. <marker>Müller</marker> wurde in Hamburg geboren."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *SimilarNameRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *SimilarNameRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *SimilarNameRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// GetID ports SimilarNameRule.getId → "DE_SIMILAR_NAMES".
func (r *SimilarNameRule) GetID() string { return "DE_SIMILAR_NAMES" }

// GetDescription ports SimilarNameRule.getDescription.
func (r *SimilarNameRule) GetDescription() string {
	return "Mögliche Tippfehler in Namen finden"
}

func (r *SimilarNameRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// IsDefaultOff ports Rule.isDefaultOff (constructor setDefaultOff).
func (r *SimilarNameRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// MinToCheckParagraph ports SimilarNameRule.minToCheckParagraph (Java returns -1).
func (r *SimilarNameRule) MinToCheckParagraph() int { return -1 }

// EstimateContextForSureMatch ports TextLevelRule.estimateContextForSureMatch → -1.
func (r *SimilarNameRule) EstimateContextForSureMatch() int { return -1 }

// levenshteinSimilar ports Apache StringUtils.getLevenshteinDistance on
// CharSequence — Java String length/charAt are UTF-16 code units, not runes.
func levenshteinSimilar(a, b string) int {
	au, bu := utf16EncodeDE(a), utf16EncodeDE(b)
	la, lb := len(au), len(bu)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if au[i-1] == bu[j-1] {
				cost = 0
			}
			ins, del, sub := cur[j-1]+1, prev[j]+1, prev[j-1]+cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}

func absIntSimilar(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// similarName ports SimilarNameRule.similarName (nullable → "" when none).
// Java HashSet iteration order is undefined; Go map iteration matches that non-determinism.
func (r *SimilarNameRule) similarName(nameHere string, namesSoFar map[string]struct{}) string {
	for name := range namesSoFar {
		if name == nameHere {
			continue
		}
		// Java: Math.abs(name.length() - nameHere.length()) — UTF-16 units
		lenDiff := absIntSimilar(utf16LenDE(name) - utf16LenDE(nameHere))
		// Genitive / dative asymmetries on s/n endings (skip pair)
		nameEndsWithS := endsWithStr(name, "s") && !endsWithStr(nameHere, "s")
		otherNameEndsWithS := !endsWithStr(name, "s") && endsWithStr(nameHere, "s")
		nameEndsWithN := endsWithStr(name, "n") && !endsWithStr(nameHere, "n")
		otherNameEndsWithN := !endsWithStr(name, "n") && endsWithStr(nameHere, "n")
		if nameEndsWithS || otherNameEndsWithS || nameEndsWithN || otherNameEndsWithN {
			// Java: "we assume this is a genitive, e.g. Angela Merkels Ehemann"
			continue
		}
		if lenDiff <= similarNameMaxDiff && levenshteinSimilar(name, nameHere) <= similarNameMaxDiff {
			return name
		}
	}
	return ""
}

func endsWithStr(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}

// isMaybeName ports the boolean isMaybeName expression inside SimilarNameRule.match
// plus the startsWithUppercase gate (Java checks both before treating as name).
//
// Java:
//
//	boolean isMaybeName = word.length() >= MIN_LENGTH
//	  && ((token.hasPartialPosTag("EIG:") && !token.hasPartialPosTag(":COU")) || token.isPosTagUnknown())
//	  && !word.equals("Dein") && ... && !word.equals("Ihrem");
//	if (isMaybeName && StringTools.startsWithUppercase(word)) { ... }
func isMaybeName(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	word := token.GetToken()
	// Java String.length() ≥ MIN_LENGTH (UTF-16 code units)
	if utf16LenDE(word) < similarNameMinLen {
		return false
	}
	// Java chained !word.equals(...) for Dein*/Ihr* forms
	switch word {
	case "Dein", "Deine", "Deinen", "Deiner", "Deines", "Deinem",
		"Ihr", "Ihre", "Ihren", "Ihrer", "Ihres", "Ihrem":
		return false
	}
	// (EIG: && !:COU) || isPosTagUnknown — not !isTagged
	posOK := (token.HasPartialPosTag("EIG:") && !token.HasPartialPosTag(":COU")) || token.IsPosTagUnknown()
	if !posOK {
		return false
	}
	return tools.StartsWithUppercase(word)
}

// MatchList ports SimilarNameRule.match(List<AnalyzedSentence>).
// pos accumulates sentence.getCorrectedTextLength(); RuleMatch suggests similarName.
func (r *SimilarNameRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	namesSoFar := map[string]struct{}{}
	var ruleMatches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		for _, token := range sentence.GetTokensWithoutWhitespace() {
			if token == nil {
				continue
			}
			word := token.GetToken()
			// Java isMaybeName + startsWithUppercase
			if !isMaybeName(token) {
				continue
			}
			if similar := r.similarName(word, namesSoFar); similar != "" {
				msg := "'" + word + "' ähnelt dem vorher benutzten '" + similar + "', handelt es sich evtl. um einen Tippfehler?"
				rm := rules.NewRuleMatch(r, sentence, pos+token.GetStartPos(), pos+token.GetEndPos(), msg)
				rm.SetSuggestedReplacement(similar)
				ruleMatches = append(ruleMatches, rm)
			}
			namesSoFar[word] = struct{}{}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
