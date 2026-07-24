package ar

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/verb_trans_to_untrans2.txt
var transVerbFS embed.FS

var (
	transVerbOnce sync.Once
	// lemma → required preposition(s) (from file; diacritic and stripped keys)
	transVerbMap map[string][]string
)

func stripArabicDiacritics(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 0x064B && r <= 0x065F) || r == 0x0640 || r == 0x0670 {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func loadTransVerbs() map[string][]string {
	transVerbOnce.Do(func() {
		b, err := transVerbFS.ReadFile("data/verb_trans_to_untrans2.txt")
		if err != nil {
			panic(err)
		}
		transVerbMap = map[string][]string{}
		for _, line := range strings.Split(string(b), "\n") {
			line = tools.JavaStringTrim(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = tools.JavaStringTrim(line[:i])
			}
			kv := strings.SplitN(line, "=", 2)
			if len(kv) < 2 {
				continue
			}
			lemma := tools.JavaStringTrim(kv[0])
			preps := strings.Split(tools.JavaStringTrim(kv[1]), "|")
			for i := range preps {
				preps[i] = tools.JavaStringTrim(preps[i])
			}
			transVerbMap[lemma] = preps
			// also index undiacritized for tagger lemmas without tashkeel
			transVerbMap[stripArabicDiacritics(lemma)] = preps
		}
	})
	return transVerbMap
}

// ArabicTransVerbRule ports org.languagetool.rules.ar.ArabicTransVerbRule.
// Match is POS+lemma gated (Java isAttachedTransitiveVerb); without POS/lemma fail closed.
// Form hooks required for suggestions (Java ArabicSynthesizer generateUnattached/Attached);
// without them Match skips the hit (no surface invent of unattached verb / bare prep).
type ArabicTransVerbRule struct {
	Messages map[string]string
	verbs    map[string][]string
	// CorrectVerbForm ports generateUnattachedNewForm (Java synthesizer).
	CorrectVerbForm func(tok *languagetool.AnalyzedTokenReadings) string
	// CorrectPrepForm ports getCorrectPrepositionForm / generateAttachedNewForm.
	CorrectPrepForm func(prep string, verbTok *languagetool.AnalyzedTokenReadings) string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewArabicTransVerbRule(messages map[string]string) *ArabicTransVerbRule {
	r := &ArabicTransVerbRule{Messages: messages, verbs: loadTransVerbs()}
	// Java demo is English placeholder (upstream as-is)
	r.AddExamplePair(
		rules.Wrong("The train arrived <marker>a hour</marker> ago."),
		rules.Fixed("The train arrived <marker>an hour</marker> ago."),
	)
	return r
}

func (r *ArabicTransVerbRule) GetID() string { return "AR_VERB_TRANSITIVE_IINDIRECT" }

func (r *ArabicTransVerbRule) GetDescription() string {
	return "َTransitive verbs corrected to indirect transitive"
}

// AddExamplePair ports Rule.addExamplePair.
func (r *ArabicTransVerbRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ArabicTransVerbRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ArabicTransVerbRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports ArabicTransVerbRule.match.
func (r *ArabicTransVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil || len(r.verbs) == 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	prevTokenIndex := 0
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		if token == nil {
			continue
		}
		var prevToken *languagetool.AnalyzedTokenReadings
		var prevTokenStr string
		if prevTokenIndex > 0 {
			prevToken = tokens[prevTokenIndex]
			prevTokenStr = prevToken.GetToken()
		}
		if prevTokenStr != "" {
			isAttached := r.isAttachedTransitiveVerb(prevToken)
			prepositions := r.getProperPrepositionForTransitiveVerb(prevToken)
			isRight := r.isRightPreposition(token, prepositions)
			if isAttached && !isRight && len(prepositions) > 0 &&
				r.CorrectVerbForm != nil && r.CorrectPrepForm != nil {
				// Java always has synthesizer; without hooks fail closed (no surface invent).
				verb := r.CorrectVerbForm(prevToken)
				newPrep := prepositions[0]
				preposition := r.CorrectPrepForm(newPrep, prevToken)
				if verb != "" && preposition != "" {
					replacement := verb + " " + preposition
					msg := "قل " + replacement + " بدلا من '" + prevTokenStr + "' لأنّ الفعل متعد بحرف."
					rm := rules.NewRuleMatch(r, sentence, prevToken.GetStartPos(), token.GetEndPos(), msg)
					rm.ShortMessage = "خطأ في الفعل المتعدي بحرف"
					rm.SetSuggestedReplacement(replacement)
					matches = append(matches, rm)
				}
			}
		}
		if r.isAttachedTransitiveVerb(token) {
			prevTokenIndex = i
		} else {
			prevTokenIndex = 0
		}
	}
	return matches
}

// isAttachedTransitiveVerb ports Java: POS present and lemma in wrongWords.
func (r *ArabicTransVerbRule) isAttachedTransitiveVerb(mytoken *languagetool.AnalyzedTokenReadings) bool {
	if mytoken == nil {
		return false
	}
	for _, verbTok := range mytoken.GetReadings() {
		if verbTok == nil || verbTok.GetPOSTag() == nil {
			continue
		}
		if verbTok.GetLemma() == nil || *verbTok.GetLemma() == "" {
			continue
		}
		if _, ok := r.verbs[*verbTok.GetLemma()]; ok {
			return true
		}
		// undiacritized lemma key
		if _, ok := r.verbs[stripArabicDiacritics(*verbTok.GetLemma())]; ok {
			return true
		}
	}
	return false
}

func (r *ArabicTransVerbRule) getProperPrepositionForTransitiveVerb(mytoken *languagetool.AnalyzedTokenReadings) []string {
	if mytoken == nil {
		return nil
	}
	for _, verbTok := range mytoken.GetReadings() {
		if verbTok == nil || verbTok.GetPOSTag() == nil || verbTok.GetLemma() == nil {
			continue
		}
		lemma := *verbTok.GetLemma()
		if preps, ok := r.verbs[lemma]; ok {
			return preps
		}
		if preps, ok := r.verbs[stripArabicDiacritics(lemma)]; ok {
			return preps
		}
	}
	return nil
}

// isRightPreposition ports Java: next token first reading lemma in preposition list.
func (r *ArabicTransVerbRule) isRightPreposition(nextToken *languagetool.AnalyzedTokenReadings, prepositionList []string) bool {
	if nextToken == nil || len(prepositionList) == 0 {
		return false
	}
	rds := nextToken.GetReadings()
	if len(rds) == 0 {
		return false
	}
	var nextLemma string
	if rds[0] != nil && rds[0].GetLemma() != nil {
		nextLemma = *rds[0].GetLemma()
	}
	if nextLemma == "" {
		// fail closed without lemma (Java uses getLemma on first reading)
		return false
	}
	for _, p := range prepositionList {
		if nextLemma == p || stripArabicDiacritics(nextLemma) == stripArabicDiacritics(p) {
			return true
		}
	}
	return false
}
