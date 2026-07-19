package de

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// VerbAgreementRule ports org.languagetool.rules.de.VerbAgreementRule:
// - Morph: VER person/number vs ich/du/er/wir (POS-gated; no surface invent)
// - TextLevel: split on ", <conjunction>" (weil/obwohl/dass/…)
// - Full ANTI_PATTERNS table (VerbAgreementAntiPatterns)
type VerbAgreementRule struct {
	Messages map[string]string
	// Category ports setCategory(GRAMMAR).
	Category *rules.Category
	// Synth optional for verb form suggestions (Java language.getSynthesizer()).
	Synth synthesis.Synthesizer
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewVerbAgreementRule(messages map[string]string) *VerbAgreementRule {
	r := &VerbAgreementRule{
		Messages: messages,
		Category: rules.CatGrammar.GetCategory(messages),
	}
	// Java: Ich bist → Ich bin
	r.AddExamplePair(
		rules.Wrong("Ich <marker>bist</marker> über die Entwicklung sehr froh."),
		rules.Fixed("Ich <marker>bin</marker> über die Entwicklung sehr froh."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *VerbAgreementRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *VerbAgreementRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *VerbAgreementRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// WithSynth sets synthesizer for morph-path suggestions.
func (r *VerbAgreementRule) WithSynth(s synthesis.Synthesizer) *VerbAgreementRule {
	if r != nil {
		r.Synth = s
	}
	return r
}

func (r *VerbAgreementRule) GetID() string { return "DE_VERBAGREEMENT" }

// GetDescription ports VerbAgreementRule.getDescription.
func (r *VerbAgreementRule) GetDescription() string {
	return "Kongruenz von Subjekt und Prädikat (nur 1. u. 2. Person oder m. Personalpronomen), z.B. 'Er bist (ist)'"
}

func (r *VerbAgreementRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// EstimateContextForSureMatch ports VerbAgreementRule.estimateContextForSureMatch → 0.
func (r *VerbAgreementRule) EstimateContextForSureMatch() int { return 0 }

// MinToCheckParagraph ports TextLevelRule.minToCheckParagraph (Java returns 0).
func (r *VerbAgreementRule) MinToCheckParagraph() int { return 0 }

// Java COMMA is the literal low-9 quotation mark ‚ (not ASCII comma).
var verbAgreementSpecialCommaRE = regexp.MustCompile(`‚`)

// CONJUNCTIONS ports VerbAgreementRule.CONJUNCTIONS (active set only).
var verbAgreementConjunctions = map[string]struct{}{
	"weil": {}, "obwohl": {}, "dass": {}, "indem": {}, "sodass": {},
}

var (
	verbAgreementAntiOnce  sync.Once
	verbAgreementAntiRules []*disambigrules.DisambiguationPatternRule
)

func verbAgreementAntiPatternRules() []*disambigrules.DisambiguationPatternRule {
	verbAgreementAntiOnce.Do(func() {
		// Full table from VerbAgreementAntiPatterns (Java ANTI_PATTERNS).
		aps := VerbAgreementAntiPatterns
		verbAgreementAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			verbAgreementAntiRules = append(verbAgreementAntiRules, rule)
		}
	})
	return verbAgreementAntiRules
}

func (r *VerbAgreementRule) getSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := verbAgreementAntiPatternRules()
	if len(aps) == 0 {
		return sentence
	}
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
		if ap != nil {
			immunized = ap.Replace(immunized)
		}
	}
	return immunized
}

// binIgnorePrev ports VerbAgreementRule.BIN_IGNORE (names before "bin").
var binIgnorePrev = map[string]struct{}{
	"Suleiman": {}, "Mohamed": {}, "Muhammad": {}, "Muhammed": {}, "Mohammed": {}, "Mohammad": {},
	"Mansour": {}, "Qaboos": {}, "Qabus": {}, "Tamim": {}, "Majid": {}, "Salman": {}, "Ghazi": {},
	"Mahathir": {}, "Madschid": {}, "Maktum": {}, "al-Aziz": {}, "Asis": {}, "Numan": {},
	"Hussein": {}, "Abdul": {}, "Abdulla": {}, "Abdullah": {}, "Isa": {}, "Osama": {}, "Said": {},
	"Zayid": {}, "Zayed": {}, "Hamad": {}, "Chalifa": {}, "Raschid": {}, "Turki": {}, "/": {},
}

var verbAgreementQuotationMarks = map[string]struct{}{
	"\"": {}, "„": {}, "»": {}, "«": {}, "'": {}, "“": {}, "”": {},
}

func (r *VerbAgreementRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	// Single-sentence entry: also used per partial after MatchList splits.
	return r.matchOne(sentence, 0, sentence)
}

// MatchList ports TextLevelRule.match(List<AnalyzedSentence>):
// within each sentence, split on ", <conjunction>" and run match on each part
// with cumulative character offset for positions.
func (r *VerbAgreementRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || len(sentences) == 0 {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		tokens := sentence.GetTokens() // full token stream including whitespace (Java)
		idx := 0
		for i := 2; i < len(tokens); i++ {
			if tokens[i-2] == nil || tokens[i] == nil {
				continue
			}
			// Java: ",".equals(tokens[i-2].getToken()) && CONJUNCTIONS.contains(tokens[i].getToken())
			if tokens[i-2].GetToken() == "," {
				if _, ok := verbAgreementConjunctions[tokens[i].GetToken()]; ok {
					partial := languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings(nil), tokens[idx:i]...))
					ruleMatches = append(ruleMatches, r.matchOne(partial, pos, sentence)...)
					idx = i
				}
			}
		}
		partial := languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings(nil), tokens[idx:]...))
		ruleMatches = append(ruleMatches, r.matchOne(partial, pos, sentence)...)
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}

func (r *VerbAgreementRule) matchOne(sentence *languagetool.AnalyzedSentence, pos int, whole *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	// Morph-only (Java); untagged AnalyzePlain fails closed.
	return r.matchMorph(sentence, pos, whole)
}

func (r *VerbAgreementRule) matchMorph(sentence *languagetool.AnalyzedSentence, pos int, whole *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	imm := r.getSentenceWithImmunization(sentence)
	tokens := imm.GetTokensWithoutWhitespace()
	if len(tokens) < 4 {
		// Java: ignore one-word sentences (SENT_START, word, SENT_END)
		return nil
	}
	posIch, posDu, posEr, posWir := -1, -1, -1, -1
	posVer1Sin, posVer2Sin, posVer1Plu := -1, -1, -1
	posPossibleVer1Sin, posPossibleVer2Sin, posPossibleVer3Sin, posPossibleVer1Plu := -1, -1, -1, -1

	for i := 1; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		strToken := strings.ToLower(tokens[i].GetToken())
		strToken = verbAgreementSpecialCommaRE.ReplaceAllString(strToken, "")
		switch strToken {
		case "ich":
			posIch = i
		case "du":
			posDu = i
		case "er":
			posEr = i
		case "wir":
			posWir = i
		}
		tok := tokens[i].GetToken()
		if tok == "" {
			continue
		}
		firstUpper := unicode.IsUpper([]rune(tok)[0])
		// Java: lowercase verb, or first content word, or after quotation
		okCase := !firstUpper || i == 1 || (i > 0 && isVerbAgreementQuotation(tokens[i-1]))
		if tokens[i].HasPartialPosTag("VER") && okCase {
			if hasUnambiguouslyPersonAndNumber(tokens[i], "1", "SIN") {
				if !(strToken == "bin" && (binIgnorePredecessor(tokens, i) ||
					(i+1 < len(tokens) && strings.HasPrefix(tokens[i+1].GetToken(), "Laden")))) {
					posVer1Sin = i
				}
			} else if hasUnambiguouslyPersonAndNumber(tokens[i], "2", "SIN") && tokens[i].GetToken() != "Probst" {
				posVer2Sin = i
			} else if hasUnambiguouslyPersonAndNumber(tokens[i], "1", "PLU") {
				posVer1Plu = i
			}
			if tokens[i].HasPartialPosTag(":1:SIN") {
				posPossibleVer1Sin = i
			}
			if tokens[i].HasPartialPosTag(":2:SIN") {
				posPossibleVer2Sin = i
			}
			if tokens[i].HasPartialPosTag(":3:SIN") {
				posPossibleVer3Sin = i
			}
			if tokens[i].HasPartialPosTag(":1:PLU") {
				posPossibleVer1Plu = i
			}
		}
	}

	var ruleMatches []*rules.RuleMatch
	var finiteVerb *languagetool.AnalyzedTokenReadings

	// VER:1:SIN without "ich"
	if posVer1Sin != -1 && posIch == -1 && !isVerbAgreementQuotation(tokens[posVer1Sin-1]) {
		if !tokens[posVer1Sin].IsImmunized() {
			ruleMatches = append(ruleMatches, r.ruleMatchWrongVerb(tokens[posVer1Sin], pos, whole))
		}
	} else if posIch > 0 && !isNearVerb(posPossibleVer1Sin, posIch) &&
		ichLooksLikeSubject(tokens, posIch) &&
		(!isVerbAgreementQuotation(tokens[posIch-1]) || posIch < 3 || (posIch > 1 && tokens[posIch-2].GetToken() == ":")) {
		plus1 := 0
		if posIch+1 < len(tokens) {
			plus1 = 1
		}
		ok, fv := verbDoesMatchPersonAndNumber(tokens[posIch-1], tokens[posIch+plus1], "1", "SIN", finiteVerb)
		finiteVerb = fv
		if !ok && !nextButOneIsModal(tokens, posIch) && (fv == nil || fv.GetToken() != "äußerst") {
			if !tokens[posIch].IsImmunized() && fv != nil {
				ruleMatches = append(ruleMatches, r.ruleMatchWrongVerbSubject(tokens[posIch], fv, "1:SIN", pos, whole))
			}
		}
	}

	if posVer2Sin != -1 && posDu == -1 && !isVerbAgreementQuotation(tokens[posVer2Sin-1]) {
		if !tokens[posVer2Sin].IsImmunized() {
			ruleMatches = append(ruleMatches, r.ruleMatchWrongVerb(tokens[posVer2Sin], pos, whole))
		}
	} else if posDu > 0 && !isNearVerb(posPossibleVer2Sin, posDu) &&
		(!isVerbAgreementQuotation(tokens[posDu-1]) || posDu < 3 || (posDu > 1 && tokens[posDu-2].GetToken() == ":")) {
		plus1 := 0
		if posDu+1 < len(tokens) {
			plus1 = 1
		}
		ok, fv := verbDoesMatchPersonAndNumber(tokens[posDu-1], tokens[posDu+plus1], "2", "SIN", finiteVerb)
		finiteVerb = fv
		if !ok && fv != nil &&
			!tokens[posDu+plus1].HasPosTagStartingWith("VER:1:SIN:KJ2") &&
			!(tokens[posDu+plus1].HasPosTagStartingWith("ADJ:") && !tokens[posDu+plus1].HasPosTag("ADJ:PRD:GRU")) &&
			!tokens[posDu-1].HasPosTagStartingWith("VER:1:SIN:KJ2") &&
			!nextButOneIsModal(tokens, posDu) &&
			!tokens[posDu].IsImmunized() {
			ruleMatches = append(ruleMatches, r.ruleMatchWrongVerbSubject(tokens[posDu], fv, "2:SIN", pos, whole))
		}
	}

	if posEr > 0 && !isNearVerb(posPossibleVer3Sin, posEr) &&
		(!isVerbAgreementQuotation(tokens[posEr-1]) || posEr < 3 || (posEr > 1 && tokens[posEr-2].GetToken() == ":")) {
		plus1 := 0
		if posEr+1 < len(tokens) {
			plus1 = 1
		}
		ok, fv := verbDoesMatchPersonAndNumber(tokens[posEr-1], tokens[posEr+plus1], "3", "SIN", finiteVerb)
		finiteVerb = fv
		if !ok && !nextButOneIsModal(tokens, posEr) && fv != nil &&
			fv.GetToken() != "äußerst" && fv.GetToken() != "regen" &&
			!tokens[posEr].IsImmunized() {
			ruleMatches = append(ruleMatches, r.ruleMatchWrongVerbSubject(tokens[posEr], fv, "3:SIN", pos, whole))
		}
	}

	if posVer1Plu != -1 && posWir == -1 && !isVerbAgreementQuotation(tokens[posVer1Plu-1]) {
		if !tokens[posVer1Plu].IsImmunized() {
			ruleMatches = append(ruleMatches, r.ruleMatchWrongVerb(tokens[posVer1Plu], pos, whole))
		}
	} else if posWir > 0 && !isNearVerb(posPossibleVer1Plu, posWir) && !isVerbAgreementQuotation(tokens[posWir-1]) {
		plus1 := 0
		if posWir+1 < len(tokens) {
			plus1 = 1
		}
		ok, fv := verbDoesMatchPersonAndNumber(tokens[posWir-1], tokens[posWir+plus1], "1", "PLU", finiteVerb)
		if !ok && !nextButOneIsModal(tokens, posWir) && !tokens[posWir].IsImmunized() &&
			(fv == nil || fv.GetToken() != "äußerst") && fv != nil {
			ruleMatches = append(ruleMatches, r.ruleMatchWrongVerbSubject(tokens[posWir], fv, "1:PLU", pos, whole))
		}
	}
	return ruleMatches
}

func binIgnorePredecessor(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i <= 0 || tokens[i-1] == nil {
		return false
	}
	_, ok := binIgnorePrev[tokens[i-1].GetToken()]
	return ok
}

// ichLooksLikeSubject ports the Java subject gate for posIch:
// token.equals("ich") || getStartPos() <= 1 || ("Ich" after ":")
func ichLooksLikeSubject(tokens []*languagetool.AnalyzedTokenReadings, posIch int) bool {
	if posIch < 0 || posIch >= len(tokens) || tokens[posIch] == nil {
		return false
	}
	t := tokens[posIch]
	tok := t.GetToken()
	if tok == "ich" {
		return true
	}
	// Java: startPos <= 1 applies to any surface at sentence start (incl. "Ich")
	if t.GetStartPos() <= 1 {
		return true
	}
	if tok == "Ich" {
		if posIch >= 2 && tokens[posIch-2] != nil && tokens[posIch-2].GetToken() == ":" {
			return true
		}
		if posIch >= 1 && tokens[posIch-1] != nil && tokens[posIch-1].GetToken() == ":" {
			return true
		}
	}
	return false
}

func nextButOneIsModal(tokens []*languagetool.AnalyzedTokenReadings, pos int) bool {
	return pos < len(tokens)-2 && tokens[pos+2] != nil && tokens[pos+2].HasPartialPosTag(":MOD:")
}

func isNearVerb(a, b int) bool {
	return a != -1 && verbAgreementAbs(a-b) < 5
}

func verbAgreementAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func isVerbAgreementQuotation(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	_, ok := verbAgreementQuotationMarks[token.GetToken()]
	return ok
}

func hasUnambiguouslyPersonAndNumber(tokenReadings *languagetool.AnalyzedTokenReadings, person, number string) bool {
	if tokenReadings == nil || tokenReadings.GetToken() == "" {
		return false
	}
	tok := tokenReadings.GetToken()
	if unicode.IsUpper([]rune(tok)[0]) && tokenReadings.GetStartPos() != 0 {
		return false
	}
	if !tokenReadings.HasPosTagStartingWith("VER") {
		return false
	}
	needle := ":" + person + ":" + number
	for _, at := range tokenReadings.GetReadings() {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		postag := *at.GetPOSTag()
		if strings.HasSuffix(postag, "_END") {
			continue
		}
		if !strings.Contains(postag, needle) {
			return false
		}
	}
	return true
}

func isFiniteVerb(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil || token.GetToken() == "" {
		return false
	}
	tok := token.GetToken()
	if unicode.IsUpper([]rune(tok)[0]) && token.GetStartPos() != 0 {
		return false
	}
	if !token.HasPosTagStartingWith("VER") {
		return false
	}
	if token.HasPartialPosTag("PA2") || token.HasPartialPosTag("PRO:") || token.HasPartialPosTag("ZAL") {
		return false
	}
	if token.GetToken() == "einst" {
		return false
	}
	return token.HasPartialPosTag(":1:") || token.HasPartialPosTag(":2:") || token.HasPartialPosTag(":3:")
}

func verbDoesMatchPersonAndNumber(token1, token2 *languagetool.AnalyzedTokenReadings, person, number string,
	_ *languagetool.AnalyzedTokenReadings) (match bool, finiteVerb *languagetool.AnalyzedTokenReadings) {
	if token1 != nil {
		t := token1.GetToken()
		if t == "," || t == "und" || t == "sowie" || t == "&" {
			return true, finiteVerb
		}
	}
	if token2 != nil {
		t := token2.GetToken()
		if t == "," || t == "und" || t == "sowie" || t == "&" {
			return true, finiteVerb
		}
	}
	found := false
	needle := ":" + person + ":" + number
	if isFiniteVerb(token1) {
		found = true
		finiteVerb = token1
		if token1.HasPartialPosTag(needle) {
			return true, finiteVerb
		}
	}
	if isFiniteVerb(token2) {
		found = true
		finiteVerb = token2
		if token2.HasPartialPosTag(needle) {
			return true, finiteVerb
		}
	}
	// Java: !foundFiniteVerb means match=true (no finite verb nearby → no error)
	return !found, finiteVerb
}

func (r *VerbAgreementRule) ruleMatchWrongVerb(token *languagetool.AnalyzedTokenReadings, pos int, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	msg := "Möglicherweise fehlende grammatische Übereinstimmung zwischen Subjekt und Prädikat (" +
		token.GetToken() + ") bezüglich Person oder Numerus (Einzahl, Mehrzahl - Beispiel: " +
		"'Max bist' statt 'Max ist')."
	return rules.NewRuleMatch(r, sentence, pos+token.GetStartPos(), pos+token.GetEndPos(), msg)
}

func (r *VerbAgreementRule) ruleMatchWrongVerbSubject(subject, verb *languagetool.AnalyzedTokenReadings, expectedVerbPOS string, pos int, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	msg := "Möglicherweise fehlende grammatische Übereinstimmung zwischen Subjekt (" + subject.GetToken() +
		") und Prädikat (" + verb.GetToken() + ") bezüglich Person oder Numerus (Einzahl, Mehrzahl - Beispiel: " +
		"'ich sind' statt 'ich bin')."
	var suggestions []string
	// Java: RuleMatch without shortMessage (null).
	if subject.GetStartPos() < verb.GetStartPos() {
		rm := rules.NewRuleMatch(r, sentence, pos+subject.GetStartPos(), pos+verb.GetEndPos(), msg)
		for _, vs := range r.getVerbSuggestions(verb, expectedVerbPOS, false) {
			suggestions = append(suggestions, subject.GetToken()+" "+vs)
		}
		toUpper := utf8.RuneCountInString(subject.GetToken()) > 0 && unicode.IsUpper([]rune(subject.GetToken())[0])
		for _, ps := range getPronounSuggestions(verb, toUpper) {
			suggestions = append(suggestions, ps+" "+verb.GetToken())
		}
		sortSuggestionsBySimilarity(suggestions, subject.GetToken()+" "+verb.GetToken())
		if len(suggestions) > 0 {
			rm.SetSuggestedReplacements(suggestions)
		}
		return rm
	}
	rm := rules.NewRuleMatch(r, sentence, pos+verb.GetStartPos(), pos+subject.GetEndPos(), msg)
	toUpper := utf8.RuneCountInString(verb.GetToken()) > 0 && unicode.IsUpper([]rune(verb.GetToken())[0])
	for _, vs := range r.getVerbSuggestions(verb, expectedVerbPOS, toUpper) {
		suggestions = append(suggestions, vs+" "+subject.GetToken())
	}
	for _, ps := range getPronounSuggestions(verb, false) {
		suggestions = append(suggestions, verb.GetToken()+" "+ps)
	}
	sortSuggestionsBySimilarity(suggestions, verb.GetToken()+" "+subject.GetToken())
	if len(suggestions) > 0 {
		rm.SetSuggestedReplacements(suggestions)
	}
	return rm
}

func (r *VerbAgreementRule) getVerbSuggestions(verb *languagetool.AnalyzedTokenReadings, expectedVerbPOS string, toUppercase bool) []string {
	// Java uses synthesizer only — fail closed without Synth / VER reading (no surface invent).
	if r == nil || r.Synth == nil || verb == nil {
		return nil
	}
	var verbToken *languagetool.AnalyzedToken
	for _, token := range verb.GetReadings() {
		if token != nil && token.GetPOSTag() != nil && strings.HasPrefix(*token.GetPOSTag(), "VER:") {
			verbToken = token
			break
		}
	}
	if verbToken == nil {
		return nil
	}
	// Java: synthesize(verbToken, "VER.*:"+expectedVerbPOS+".*", true)
	tagRE := "VER.*:" + expectedVerbPOS + ".*"
	forms, err := r.Synth.Synthesize(verbToken, tagRE)
	if err != nil || len(forms) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, f := range forms {
		if toUppercase {
			f = tools.UppercaseFirstChar(f)
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	return out
}

func getPronounSuggestions(verb *languagetool.AnalyzedTokenReadings, toUppercase bool) []string {
	var result []string
	if verb.HasPartialPosTag(":1:SIN") {
		result = append(result, "ich")
	}
	if verb.HasPartialPosTag(":2:SIN") {
		result = append(result, "du")
	}
	if verb.HasPartialPosTag(":3:SIN") {
		result = append(result, "er", "sie", "es")
	}
	if verb.HasPartialPosTag(":1:PLU") {
		result = append(result, "wir")
	}
	if verb.HasPartialPosTag(":2:PLU") {
		result = append(result, "ihr")
	}
	if verb.HasPartialPosTag(":3:PLU") {
		hasSie := false
		for _, r := range result {
			if r == "sie" {
				hasSie = true
				break
			}
		}
		if !hasSie {
			result = append(result, "sie")
		}
	}
	if toUppercase {
		for i := range result {
			result[i] = tools.UppercaseFirstChar(result[i])
		}
	}
	return result
}

func sortSuggestionsBySimilarity(suggestions []string, markedText string) {
	sort.SliceStable(suggestions, func(i, j int) bool {
		return agreementLevenshtein(markedText, suggestions[i]) < agreementLevenshtein(markedText, suggestions[j])
	})
}
