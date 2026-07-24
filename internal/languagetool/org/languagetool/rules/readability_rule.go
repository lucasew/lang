package rules

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// FleschReadingEase computes English Flesch Reading Ease:
// 206.835 - 1.015*(words/sentences) - 84.6*(syllables/words)
func FleschReadingEase(sentences, words, syllables int) float64 {
	if sentences <= 0 || words <= 0 {
		return 0
	}
	asl := float64(words) / float64(sentences)
	asw := float64(syllables) / float64(words)
	return 206.835 - 1.015*asl - 84.6*asw
}

// CountSyllablesEN is a simple English syllable estimator (vowel-group based).
func CountSyllablesEN(word string) int {
	w := strings.ToLower(word)
	if w == "" {
		return 0
	}
	// strip non-letters
	var b strings.Builder
	for _, r := range w {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}
	w = b.String()
	if w == "" {
		return 0
	}
	vowels := "aeiouy"
	count := 0
	prevV := false
	for _, r := range w {
		isV := strings.ContainsRune(vowels, r)
		if isV && !prevV {
			count++
		}
		prevV = isV
	}
	// silent e
	if strings.HasSuffix(w, "e") && count > 1 {
		count--
	}
	if count < 1 {
		count = 1
	}
	return count
}

// ReadabilityLevel maps Flesch score to school-grade style level 0–6 (simplified).
// Higher level = harder text.
func ReadabilityLevel(flesch float64) int {
	switch {
	case flesch >= 90:
		return 0
	case flesch >= 80:
		return 1
	case flesch >= 70:
		return 2
	case flesch >= 60:
		return 3
	case flesch >= 50:
		return 4
	case flesch >= 30:
		return 5
	default:
		return 6
	}
}

// ReadabilityRule ports metadata for org.languagetool.rules.ReadabilityRule.
// Java: Category TEXT_ANALYSIS (default off category), Style; setDefaultOff when !defaultOn.
type ReadabilityRule struct {
	TooEasyTest bool
	Level       int // threshold level
	MinWords    int
	// Category ports Rule.category (Java TEXT_ANALYSIS).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// DefaultOff ports setDefaultOff (Java when defaultOn=false).
	DefaultOff bool
	// CountSyllables defaults to CountSyllablesEN when nil.
	CountSyllables func(word string) int
}

func NewReadabilityRule(tooEasy bool, level int) *ReadabilityRule {
	// Java: TEXT_ANALYSIS category with onByDefault=false; typically setDefaultOff().
	cat := NewCategoryFull(NewCategoryId("TEXT_ANALYSIS"), "Text Analysis", CategoryInternal, false, "")
	return &ReadabilityRule{
		TooEasyTest: tooEasy,
		Level:       level,
		MinWords:    10,
		Category:    cat,
		IssueType:   ITSStyle,
		DefaultOff:  true,
	}
}

func (r *ReadabilityRule) GetID() string {
	if r.TooEasyTest {
		return "READABILITY_RULE_SIMPLE"
	}
	return "READABILITY_RULE_DIFFICULT"
}

func (r *ReadabilityRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *ReadabilityRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *ReadabilityRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// EvaluateParagraph returns flesch score and level for whitespace-split words/sentences.
// sentences is count of sentences in the paragraph.
func (r *ReadabilityRule) EvaluateParagraph(sentences int, words []string) (flesch float64, level int, tooExtreme bool) {
	if len(words) < r.MinWords || sentences < 1 {
		return 0, 0, false
	}
	sylFn := r.CountSyllables
	if sylFn == nil {
		sylFn = CountSyllablesEN
	}
	syl := 0
	for _, w := range words {
		syl += sylFn(w)
	}
	flesch = FleschReadingEase(sentences, len(words), syl)
	level = ReadabilityLevel(flesch)
	if r.TooEasyTest {
		tooExtreme = level > r.Level && r.Level >= 0
	} else {
		tooExtreme = level < r.Level && r.Level >= 0
	}
	// when level threshold is -1, treat as disabled threshold
	if r.Level < 0 {
		tooExtreme = false
	}
	return flesch, level, tooExtreme
}

// MatchList ports ReadabilityRule.match (paragraph FRE; typically default-off).
func (r *ReadabilityRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil || len(sentences) == 0 {
		return nil
	}
	var words []string
	nSent := 0
	var startSent *languagetool.AnalyzedSentence
	startPos, endPos := -1, -1
	pos := 0
	for _, s := range sentences {
		if s == nil {
			continue
		}
		nSent++
		toks := s.GetTokensWithoutWhitespace()
		if startSent == nil && len(toks) > 1 {
			startSent = s
			startPos = pos + toks[1].GetStartPos()
			if len(toks) > 3 {
				endPos = pos + toks[3].GetEndPos()
			} else {
				endPos = pos + toks[len(toks)-1].GetEndPos()
			}
		}
		for _, t := range toks {
			if t == nil || t.IsSentenceStart() || t.IsSentenceEnd() || t.IsNonWord() {
				continue
			}
			words = append(words, t.GetToken())
		}
		pos += s.GetCorrectedTextLength()
	}
	_, _, too := r.EvaluateParagraph(nSent, words)
	if !too || startSent == nil || startPos < 0 || endPos < 0 {
		return nil
	}
	msg := "This text is too difficult to read"
	if r.TooEasyTest {
		msg = "This text is too easy to read"
	}
	return []*RuleMatch{NewRuleMatch(r, startSent, startPos, endPos, msg)}
}
