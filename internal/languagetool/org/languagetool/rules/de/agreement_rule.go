package de

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// AgreementRule ports AgreementRule (morph DET/PRO–NOUN paths):
// - DET/PRO + SUB mismatch when both carry POS tags (:STV empty-set1 for "Meiner Chef")
// - DET/PRO + ADJ + SUB (and ADJ+ADJ+SUB) via retainCommonCategories
// - anti-pattern immunization (getSentenceWithImmunization + AllAgreementAntiPatterns)
// - optional AgreementSuggestor2 when Synth is set (incl. setSkipped)
// - modifiers (sehr/…), ignored pronouns/nouns, relative-clause skips
// - HERR_FRAU only when next is untagged/EIG; allowSuggestion gates replMap
// - getCompoundError via prep_compound (requires CompoundPhraseValid = Java lt.check)
type AgreementRule struct {
	Messages map[string]string
	// Category ports setCategory(GRAMMAR).
	Category *rules.Category
	// Synth is optional (Java language.getSynthesizer()). Nil → no suggestions.
	Synth synthesis.Synthesizer
	// CompoundPhraseValid ports lt.check in getRuleMatch for open compounds
	// (Java: only DE_AGREEMENT + GERMAN_SPELLER enabled). Nil → fail-closed (no invent).
	CompoundPhraseValid func(phrase string) bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewAgreementRule(messages map[string]string) *AgreementRule {
	r := &AgreementRule{
		Messages: messages,
		Category: rules.CatGrammar.GetCategory(messages),
	}
	// Java: Der Haus → Das Haus
	r.AddExamplePair(
		rules.Wrong("<marker>Der Haus</marker> wurde letztes Jahr gebaut."),
		rules.Fixed("<marker>Das Haus</marker> wurde letztes Jahr gebaut."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AgreementRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AgreementRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AgreementRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// WithSynth sets the synthesizer used by AgreementSuggestor2.
func (r *AgreementRule) WithSynth(s synthesis.Synthesizer) *AgreementRule {
	if r != nil {
		r.Synth = s
	}
	return r
}

func (r *AgreementRule) GetID() string { return "DE_AGREEMENT" }

// GetDescription ports AgreementRule.getDescription.
func (r *AgreementRule) GetDescription() string {
	return "Kongruenz von Nominalphrasen (unvollständig!), z.B. 'mein kleiner (kleines) Haus'"
}

// GetURL ports AgreementRule constructor setUrl.
func (r *AgreementRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/deklination/"
}

func (r *AgreementRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// EstimateContextForSureMatch ports estimateContextForSureMatch:
// max length of all anti-pattern lists.
func (r *AgreementRule) EstimateContextForSureMatch() int {
	max := 0
	for _, ap := range AllAgreementAntiPatterns() {
		if n := len(ap); n > max {
			max = n
		}
	}
	return max
}

// Java MSG / MSG2 / SHORT_MSG (AgreementRule.java).
const (
	agreementMsg   = "Möglicherweise passen das Nomen und die Wörter, die das Nomen beschreiben, grammatisch nicht zusammen."
	agreementMsg2  = "Möglicherweise passen das Nomen und die Wörter, die das Nomen beschreiben, grammatisch nicht zusammen."
	agreementShort = "Evtl. passen Wörter grammatisch nicht zusammen."
)

var (
	agreementAntiPatternsOnce sync.Once
	agreementAntiPatternRules []*disambigrules.DisambiguationPatternRule
)

// agreementAntiPatterns ports AgreementRule.getAntiPatterns (cached IMMUNIZE rules).
func agreementAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	agreementAntiPatternsOnce.Do(func() {
		aps := AllAgreementAntiPatterns()
		agreementAntiPatternRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			// Java makeAntiPatterns: INTERNAL_ANTIPATTERN + IMMUNIZE
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			agreementAntiPatternRules = append(agreementAntiPatternRules, rule)
		}
	})
	return agreementAntiPatternRules
}

// getSentenceWithImmunization ports Rule.getSentenceWithImmunization for this rule.
// Clones tokens so immunization does not mutate the caller's sentence.
func (r *AgreementRule) getSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := agreementAntiPatterns()
	if len(aps) == 0 {
		return sentence
	}
	// Clone via FromOld (chunk tags, immunization, typographic apostrophe) then
	// NewAnalyzedSentence so non-blank index points at the clones (not Java Copy's
	// shared nonBlankTokens). Anti-pattern Replace mutates then rebuilds.
	src := sentence.GetTokens()
	cloned := make([]*languagetool.AnalyzedTokenReadings, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}
		cloned[i] = languagetool.NewAnalyzedTokenReadingsFromOld(t, t.GetReadings(), "")
	}
	immunized := languagetool.NewAnalyzedSentence(cloned)
	for _, ap := range aps {
		if ap == nil {
			continue
		}
		immunized = ap.Replace(immunized)
	}
	return immunized
}

func (r *AgreementRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	// Java: tokens = getSentenceWithImmunization(sentence).getTokensWithoutWhitespace()
	imm := r.getSentenceWithImmunization(sentence)
	tokens := imm.GetTokensWithoutWhitespace()
	// Snapshot original non-blank tokens (surface forms before ins/zur rewrite).
	// Shallow-copy the slice; replacePrepositionsByArticle replaces elements in tokens only.
	origTokens := append([]*languagetool.AnalyzedTokenReadings(nil), tokens...)
	replMap := replacePrepositionsByArticle(tokens)
	var matches []*rules.RuleMatch

	// Morphological DET/PRO … NOUN paths (need POS tags on tokens)
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		// Java: skip SENTENCE_START and immunized tokens (also check pre-rewrite)
		if tok.IsSentenceStart() || tok.IsImmunized() {
			continue
		}
		if i < len(origTokens) && origTokens[i] != nil && origTokens[i].IsImmunized() {
			continue
		}
		if couldBeRelativeOrDependentClause(tokens, i) {
			continue
		}
		// "der eine" / "die eine" false-alarm skip (Java match loop)
		if i > 0 {
			prev := strings.ToLower(tokens[i-1].GetToken())
			cur := tokens[i].GetToken()
			if (prev == "der" || prev == "die" || prev == "das" || prev == "des" || prev == "dieses") &&
				(cur == "eine" || cur == "einen") {
				continue
			}
		}
		// Art. abbrev / following participle skips
		if shouldSkipDetAbbrevOrParticiple(tokens, i) {
			continue
		}
		if !isDeterminer(tok) && !isRelevantPronoun(tokens, i) {
			continue
		}
		// Java getPosAfterModifier: "ein sehr hohes Haus" / "ein 500 Meter hohes Haus"
		afterMod := getPosAfterModifier(i+1, tokens)
		if afterMod >= len(tokens) {
			break
		}
		skippedStr := ""
		if afterMod > i+1 && i+1 < len(tokens) && tokens[i+1] != nil {
			// Java: substring of skipped modifiers between det and first non-modifier
			from := tokens[i+1].GetStartPos()
			to := tokens[afterMod-1].GetEndPos()
			if to > from {
				skippedStr = sentence.GetText()
				if from < len(skippedStr) && to <= len(skippedStr) {
					skippedStr = skippedStr[from:to]
				} else {
					// fallback: join token surfaces
					var parts []string
					for j := i + 1; j < afterMod; j++ {
						if tokens[j] != nil {
							parts = append(parts, tokens[j].GetToken())
						}
					}
					skippedStr = strings.Join(parts, " ")
				}
			}
		}
		next := tokens[afterMod]
		if next == nil || next.IsImmunized() {
			continue
		}
		repl := replMap[i]
		// Java maybePreposition = tokens[i-1] (unless "was für …")
		var maybePrep *languagetool.AnalyzedTokenReadings
		if i-1 >= 0 {
			maybePrep = tokens[i-1]
			if i-2 >= 0 && tokens[i-2] != nil && strings.EqualFold(tokens[i-2].GetToken(), "was") {
				maybePrep = nil
			}
		}
		// DET + ADJ/PA + SUB
		if isNonPredicativeAdjOrParticiple(next) {
			k := afterMod + 1
			if k >= len(tokens) {
				continue
			}
			// DET + ADJ + ADJ + SUB
			if isNonPredicativeAdjOrParticiple(tokens[k]) && k+1 < len(tokens) && isNounTagged(tokens[k+1]) {
				if tokens[k].IsImmunized() || tokens[k+1].IsImmunized() {
					continue
				}
				if rm := r.checkDetAdjAdjNoun(tokens[i], next, tokens[k], tokens[k+1], sentence, i, repl, skippedStr, origTokens, maybePrep); rm != nil {
					matches = append(matches, rm)
				}
				continue
			}
			// Java: DET+ADJ+NOUN does not skip "Herr" at the call site (HERR_FRAU handled inside).
			if isNounTagged(tokens[k]) {
				if tokens[k].IsImmunized() {
					continue
				}
				// "als das" false alarm: "weniger farbenprächtig als das anderer Papageien"
				if i >= 2 && isAdjectiveTagged(tokens[i-2]) &&
					tokens[i-1].GetToken() == "als" && tokens[i].GetToken() == "das" {
					continue
				}
				// Java allowSuggestion = tokenPos == i+2 (no modifiers between det and adj)
				// When false, replMap is null so ReplacementType is omitted from suggestor.
				allowSug := afterMod == i+1
				var replArg ReplacementType
				if allowSug {
					replArg = repl
				}
				if rm := r.checkDetAdjNoun(tokens[i], next, tokens[k], sentence, i, replArg, allowSug, skippedStr, origTokens, maybePrep); rm != nil {
					matches = append(matches, rm)
				}
			}
			continue
		}
		// DET + SUB (Java: skip bare "Herr" title at call site)
		if isNounTagged(next) && next.GetToken() != "Herr" {
			if rm := r.checkDetNoun(tokens[i], next, sentence, i, repl, skippedStr, origTokens, maybePrep); rm != nil {
				matches = append(matches, rm)
			}
		}
	}
	// Open-compound errors are handled via getCompoundError paths (dict/lt.check),
	// not by inventing hits on any two capitalized tokens.
	return matches
}

func (r *AgreementRule) checkDetNoun(det, noun *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence,
	tokenPos int, repl ReplacementType, skippedStr string, origTokens []*languagetool.AnalyzedTokenReadings, maybePrep *languagetool.AnalyzedTokenReadings) *rules.RuleMatch {
	// Java: token2.isImmunized() || NOUNS_TO_BE_IGNORED || "-"
	if noun == nil || noun.IsImmunized() || noun.GetToken() == "-" {
		return nil
	}
	if _, ok := nounsToBeIgnored[noun.GetToken()]; ok {
		return nil
	}
	// Java: single reading ending :STV → empty set1 to catch "Meiner Chef raucht."
	forceSTVMismatch := isSingleSTVReading(det)
	var set1, set2 map[string]struct{}
	if forceSTVMismatch {
		set1 = map[string]struct{}{}
	} else {
		set1 = GetAgreementCategories(det, nil, false)
	}
	set2 = GetAgreementCategories(noun, nil, false)
	if !forceSTVMismatch {
		// Without STV: fail-closed when either side has no categories (no invent on untagged).
		if len(set1) == 0 || len(set2) == 0 {
			return nil
		}
		if CategoriesIntersect(set1, set2) {
			return nil
		}
	} else {
		// STV: Java set1 empty → retainAll → always empty → match (unless exception).
		// Still require noun categories so we don't invent on untagged nouns.
		if len(set2) == 0 {
			return nil
		}
	}
	if isDetNounException(det, noun) {
		return nil
	}
	// Prefer compound-error match when next token looks like open compound (Java).
	if cm := getCompoundErrorDetNoun(det, noun, tokenPos, origTokens, sentence, r); cm != nil {
		return cm
	}
	rm := rules.NewRuleMatch(r, sentence, det.GetStartPos(), noun.GetEndPos(), agreementMsg)
	rm.ShortMessage = agreementShort
	r.attachSuggestions(rm, det, nil, nil, noun, repl, maybePrep, skippedStr)
	return rm
}

// isSingleSTVReading ports the :STV empty-set1 gate in checkDetNounAgreement.
func isSingleSTVReading(t *languagetool.AnalyzedTokenReadings) bool {
	if t == nil {
		return false
	}
	rds := t.GetReadings()
	if len(rds) != 1 || rds[0] == nil || rds[0].GetPOSTag() == nil {
		return false
	}
	return strings.HasSuffix(*rds[0].GetPOSTag(), ":STV")
}

// isDetNounException ports AgreementRule.isException.
func isDetNounException(det, noun *languagetool.AnalyzedTokenReadings) bool {
	if det == nil || noun == nil {
		return false
	}
	return det.GetToken() == "allen" && noun.GetToken() == "Grund"
}

func (r *AgreementRule) checkDetAdjNoun(det, adj, noun *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence,
	tokenPos int, repl ReplacementType, allowSuggestion bool, skippedStr string, origTokens []*languagetool.AnalyzedTokenReadings, maybePrep *languagetool.AnalyzedTokenReadings) *rules.RuleMatch {
	if noun == nil || utf16LenDE(noun.GetToken()) < 2 {
		return nil
	}
	if noun.IsImmunized() {
		return nil
	}
	common := retainCommonCategories3(det, adj, noun)
	if len(common) > 0 {
		return nil
	}
	// Java HERR_FRAU: skip only when next token is untagged or EIG ("das ignorierte Herr Grey")
	if isHerrFrau(noun.GetToken()) && tokenPos+3 < len(origTokens) {
		t4 := origTokens[tokenPos+3]
		if t4 == nil || !t4.IsTagged() || t4.HasPosTagStartingWith("EIG:") {
			return nil
		}
	}
	// Java: try getCompoundError(t[pos..pos+3], …) before 3-token compound
	if tokenPos+4 < len(origTokens) && origTokens[tokenPos] != nil &&
		origTokens[tokenPos+1] != nil && origTokens[tokenPos+2] != nil && origTokens[tokenPos+3] != nil {
		if cm := getCompoundErrorDetAdjAdjNoun(
			origTokens[tokenPos], origTokens[tokenPos+1], origTokens[tokenPos+2], origTokens[tokenPos+3],
			tokenPos, "", origTokens, sentence, r); cm != nil {
			return cm
		}
	}
	if cm := getCompoundErrorDetAdjNoun(det, adj, noun, tokenPos, origTokens, sentence, r); cm != nil {
		return cm
	}
	if noun.HasPosTagStartingWith("ABK") {
		return nil
	}
	rm := rules.NewRuleMatch(r, sentence, det.GetStartPos(), noun.GetEndPos(), agreementMsg)
	rm.ShortMessage = agreementShort
	// Java always builds suggestor; allowSuggestion only gates replMap (already applied by caller).
	_ = allowSuggestion
	r.attachSuggestions(rm, det, adj, nil, noun, repl, maybePrep, skippedStr)
	return rm
}

func isHerrFrau(s string) bool {
	return s == "Herr" || s == "Frau"
}

func (r *AgreementRule) checkDetAdjAdjNoun(det, adj1, adj2, noun *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence,
	tokenPos int, repl ReplacementType, skippedStr string, origTokens []*languagetool.AnalyzedTokenReadings, maybePrep *languagetool.AnalyzedTokenReadings) *rules.RuleMatch {
	if noun == nil || noun.IsImmunized() {
		return nil
	}
	common := retainCommonCategories4(det, adj1, adj2, noun)
	if len(common) > 0 {
		return nil
	}
	if cm := getCompoundErrorDetAdjAdjNoun(det, adj1, adj2, noun, tokenPos, skippedStr, origTokens, sentence, r); cm != nil {
		return cm
	}
	if noun.HasPosTagStartingWith("ABK") {
		return nil
	}
	rm := rules.NewRuleMatch(r, sentence, det.GetStartPos(), noun.GetEndPos(), agreementMsg2)
	rm.ShortMessage = agreementShort
	// Java only attaches suggestor for adj-adj-noun when replMap != null (always true here)
	r.attachSuggestions(rm, det, adj1, adj2, noun, repl, maybePrep, skippedStr)
	return rm
}

// attachSuggestions wires AgreementSuggestor2 when a synthesizer is available.
func (r *AgreementRule) attachSuggestions(rm *rules.RuleMatch, det, adj1, adj2, noun *languagetool.AnalyzedTokenReadings, repl ReplacementType, maybePrep *languagetool.AnalyzedTokenReadings, skippedStr string) {
	if r == nil || rm == nil || r.Synth == nil || noun == nil {
		return
	}
	s := NewAgreementSuggestor2(r.Synth, det, noun).WithReplacementType(repl).WithPreposition(maybePrep).WithSkipped(skippedStr)
	if adj1 != nil {
		s = s.WithAdjectives(adj1, adj2)
	}
	// Java: suggestor.getSuggestions(true) — keep lowest token-edit tier only.
	sugs := s.GetSuggestionsFiltered(true)
	if len(sugs) > 0 {
		// Cap like typical UI; Java returns full filtered list from suggestor.
		if len(sugs) > 20 {
			sugs = sugs[:20]
		}
		rm.SetSuggestedReplacements(sugs)
	}
}

func retainCommonCategories3(t1, t2, t3 *languagetool.AnalyzedTokenReadings) map[string]struct{} {
	skipSol := true
	if t1 != nil {
		if _, ok := vieleWenige[strings.ToLower(t1.GetToken())]; ok {
			skipSol = false
		}
	}
	return intersectMaps(
		GetAgreementCategories(t1, nil, skipSol),
		GetAgreementCategories(t2, nil, skipSol),
		GetAgreementCategories(t3, nil, true),
	)
}

func retainCommonCategories4(t1, t2, t3, t4 *languagetool.AnalyzedTokenReadings) map[string]struct{} {
	skipSol := true
	if t1 != nil {
		if _, ok := vieleWenige[strings.ToLower(t1.GetToken())]; ok {
			skipSol = false
		}
	}
	return intersectMaps(
		GetAgreementCategories(t1, nil, skipSol),
		GetAgreementCategories(t2, nil, skipSol),
		GetAgreementCategories(t3, nil, skipSol),
		GetAgreementCategories(t4, nil, true),
	)
}

func intersectMaps(sets ...map[string]struct{}) map[string]struct{} {
	if len(sets) == 0 {
		return nil
	}
	out := map[string]struct{}{}
	for k := range sets[0] {
		ok := true
		for _, s := range sets[1:] {
			if _, has := s[k]; !has {
				ok = false
				break
			}
		}
		if ok {
			out[k] = struct{}{}
		}
	}
	return out
}

// isDeterminer ports hasReadingOfType(…, DETERMINER) ≈ ART: tags.
func isDeterminer(t *languagetool.AnalyzedTokenReadings) bool {
	return t != nil && t.HasPosTagStartingWith("ART:")
}

// isDetOrPro is kept for older call sites; prefer isDeterminer / isRelevantPronoun.
func isDetOrPro(t *languagetool.AnalyzedTokenReadings) bool {
	if t == nil {
		return false
	}
	return t.HasPosTagStartingWith("ART:") || t.HasPosTagStartingWith("PRO:")
}

func isNounTagged(t *languagetool.AnalyzedTokenReadings) bool {
	if t == nil {
		return false
	}
	return t.HasPosTagStartingWith("SUB:") || t.HasPosTagStartingWith("EIG:")
}

func isAdjectiveTagged(t *languagetool.AnalyzedTokenReadings) bool {
	return t != nil && t.HasPosTagStartingWith("ADJ:")
}

// isNonPredicativeAdjOrParticiple ports isNonPredicativeAdjective || isParticiple.
func isNonPredicativeAdjOrParticiple(t *languagetool.AnalyzedTokenReadings) bool {
	if t == nil {
		return false
	}
	if t.HasPartialPosTag("PA1") || t.HasPartialPosTag("PA2") {
		return true
	}
	// Java isNonPredicativeAdjective: ADJ without :PRD
	for _, r := range t.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		pos := *r.GetPOSTag()
		if strings.HasPrefix(pos, "ADJ") && !strings.Contains(pos, "PRD") {
			return true
		}
	}
	return false
}

// isRelevantPronoun ports AgreementRule.isRelevantPronoun.
func isRelevantPronoun(tokens []*languagetool.AnalyzedTokenReadings, pos int) bool {
	if pos < 0 || pos >= len(tokens) || tokens[pos] == nil {
		return false
	}
	t := tokens[pos]
	if !t.HasPosTagStartingWith("PRO:") {
		return false
	}
	low := strings.ToLower(t.GetToken())
	if _, ok := pronounsToBeIgnored[low]; ok {
		return false
	}
	// "vor allem"
	if pos > 0 && tokens[pos-1] != nil &&
		strings.EqualFold(tokens[pos-1].GetToken(), "vor") &&
		strings.EqualFold(t.GetToken(), "allem") {
		return false
	}
	return true
}

// couldBeRelativeOrDependentClause ports AgreementRule.couldBeRelativeOrDependentClause.
func couldBeRelativeOrDependentClause(tokens []*languagetool.AnalyzedTokenReadings, pos int) bool {
	if pos >= 1 && tokens[pos-1] != nil && tokens[pos] != nil {
		// ", das Frauen zugesprochen bekamen"
		if tokens[pos-1].GetToken() == "," && tokens[pos].HasAnyLemma(relPronounLemmas...) {
			if pos+3 < len(tokens) {
				return true
			}
		}
	}
	if pos >= 2 && tokens[pos-2] != nil && tokens[pos-1] != nil && tokens[pos] != nil {
		if tokens[pos-2].GetToken() != "," {
			return false
		}
		// ", in dem …" prep + rel pronoun
		prep := tokens[pos-1].HasPosTagStartingWith("PRP:")
		rel := tokens[pos].HasAnyLemma(relPronounLemmas...)
		if prep && rel {
			return true
		}
		// ", weil diese …" KON:UNT + jen/dies/ebendies
		if tokens[pos-1].HasPosTag("KON:UNT") &&
			(tokens[pos].HasAnyLemma("jen") || tokens[pos].HasAnyLemma("dies") || tokens[pos].HasAnyLemma("ebendies")) {
			return true
		}
	}
	return false
}

// getPosAfterModifier ports AgreementRule.getPosAfterModifier.
func getPosAfterModifier(startAt int, tokens []*languagetool.AnalyzedTokenReadings) int {
	if startAt < 0 {
		return 0
	}
	if startAt < len(tokens) && tokens[startAt] != nil &&
		tokens[startAt].GetToken() == "relativ" &&
		startAt+1 < len(tokens) && tokens[startAt+1] != nil &&
		tokens[startAt+1].GetToken() == "gesehen" {
		startAt += 2
	}
	// Java: if (viel|weit + weniger|eher) +=2; else if MODIFIERS +=1
	if startAt < len(tokens) && tokens[startAt] != nil {
		tok := tokens[startAt].GetToken()
		if (tok == "viel" || tok == "weit") &&
			startAt+1 < len(tokens) && tokens[startAt+1] != nil &&
			(tokens[startAt+1].GetToken() == "weniger" || tokens[startAt+1].GetToken() == "eher") {
			startAt += 2
		} else if startAt+1 < len(tokens) {
			if _, ok := agreementModifiers[tok]; ok {
				startAt++
			}
		}
	}
	if startAt+1 < len(tokens) && tokens[startAt] != nil && tokens[startAt+1] != nil {
		phrase := strings.ToLower(tokens[startAt].GetToken() + " " + tokens[startAt+1].GetToken())
		switch phrase {
		case "mit mir", "mit dir", "mit ihm", "mit ihr", "mit ihnen", "mit uns", "mit euch",
			"ohne mich", "ohne dich", "ohne ihn", "ohne sie", "ohne uns", "ohne euch":
			startAt += 2
		}
	}
	// "500 Meter" / "1,4 Meter" measure modifiers
	if startAt+1 < len(tokens) && tokens[startAt] != nil {
		numTok := tokens[startAt]
		isNum := isDigits(numTok.GetToken()) || numTok.HasPosTag("ZAL")
		if isNum {
			posAfter := startAt + 1
			if startAt+3 < len(tokens) && tokens[startAt+1] != nil && tokens[startAt+2] != nil &&
				tokens[startAt+1].GetToken() == "," && isDigits(tokens[startAt+2].GetToken()) {
				posAfter = startAt + 3
			}
			if posAfter < len(tokens) && tokens[posAfter] != nil {
				u := tokens[posAfter].GetToken()
				if strings.HasSuffix(u, "gramm") || strings.HasSuffix(u, "Gramm") ||
					strings.HasSuffix(u, "Meter") || strings.HasSuffix(u, "meter") {
					return posAfter + 1
				}
			}
		}
	}
	return startAt
}

// shouldSkipDetAbbrevOrParticiple ports Art./participle false-alarm guards in match().
func shouldSkipDetAbbrevOrParticiple(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	// "Art. 1" / "bisherigen Art. 1"
	if i+2 < len(tokens) && tokens[i+1] != nil && tokens[i+2] != nil &&
		tokens[i+1].GetToken() == "Art" && tokens[i+2].GetToken() == "." {
		return true
	}
	if i+3 < len(tokens) && tokens[i+2] != nil && tokens[i+3] != nil &&
		tokens[i+2].GetToken() == "Art" && tokens[i+3].GetToken() == "." {
		return true
	}
	// "einen Hochwasser führenden Fluss"
	if i+2 < len(tokens) && tokens[i+2] != nil {
		if tokens[i+2].HasPartialPosTag("PA1") {
			return true
		}
		low := strings.ToLower(tokens[i+2].GetToken())
		if low == "zugeschriebenen" || low == "zugeschriebene" ||
			low == "genannten" || low == "genannte" {
			return true
		}
	}
	return false
}
