package de

import (
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CaseRule ports org.languagetool.rules.de.CaseRule.
// Morph/POS path only (Java) — no surface invent for untagged AnalyzePlain.
// Lookup / IsMisspelled are optional hooks (WireCaseRule attaches GermanTagger when available).
// Java: setCategory(CASING); setUrl (GetURL overrides constructor setUrl).
type CaseRule struct {
	Messages     map[string]string
	Category     *rules.Category
	Lookup       func(word string) *languagetool.AnalyzedTokenReadings // GermanTagger.lookup
	IsMisspelled func(lowerWord string) bool                           // default spelling rule
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewCaseRule(messages map[string]string) *CaseRule {
	r := &CaseRule{
		Messages: messages,
		Category: rules.CatCasing.GetCategory(messages),
	}
	// Java: Das laufen → Das Laufen
	r.AddExamplePair(
		rules.Wrong("<marker>Das laufen</marker> fällt mir schwer."),
		rules.Fixed("<marker>Das Laufen</marker> fällt mir schwer."),
	)
	return r
}

func (r *CaseRule) GetID() string { return "DE_CASE" }

func (r *CaseRule) GetDescription() string {
	return "Großschreibung von Nomen und substantivierten Verben"
}

func (r *CaseRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// AddExamplePair ports Rule.addExamplePair.
func (r *CaseRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *CaseRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *CaseRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// EstimateContextForSureMatch ports CaseRule.estimateContextForSureMatch:
// max length of ANTI_PATTERNS lists.
func (r *CaseRule) EstimateContextForSureMatch() int {
	max := 0
	for _, ap := range CaseRuleAntiPatterns {
		if n := len(ap); n > max {
			max = n
		}
	}
	return max
}

// GetURL ports CaseRule.getUrl (overrides constructor setUrl).
func (r *CaseRule) GetURL() string {
	return "https://dict.leo.org/grammatik/deutsch/Rechtschreibung/Regeln/Gross-klein/index.html"
}

const (
	caseUppercaseMessage = "Außer am Satzanfang werden nur Nomen und Eigennamen großgeschrieben."
	caseLowercaseMessage = "Falls es sich um ein substantiviertes Verb handelt, wird es großgeschrieben."
	caseColonMessage     = "Folgt dem Doppelpunkt weder ein Substantiv noch eine wörtliche Rede oder ein vollständiger Hauptsatz, schreibt man klein weiter."
)

var (
	caseNumeralsEN   = regexp.MustCompile(`(?i)^([a-z]|[0-9]+|(m{0,4}(c[md]|d?c{0,3})(x[cl]|l?x{0,3})(i[xv]|v?i{0,3})))$`)
	caseTwoUppercase = regexp.MustCompile(`^[A-ZÖÄÜ][A-ZÖÄÜ][a-zöäüß-]+$`)
	caseVerhaltenRE  = regexp.MustCompile(`.+verhalten`)
	caseIrgendEtcRE  = regexp.MustCompile(`^(irgendwelche|irgendwas|irgendein|weniger?|einiger?|mehr|aufs)$`)
	caseAllnmRE      = regexp.MustCompile(`^Alle[nm]$`)

	caseNounIndicators = map[string]struct{}{
		"das": {}, "sein": {}, "mein": {}, "dein": {}, "euer": {}, "unser": {},
	}

	caseSentenceStartExceptions = map[string]struct{}{
		"(": {}, "\"": {}, "'": {}, "‘": {}, "„": {}, "«": {}, "»": {}, "‚": {},
		".": {}, "!": {}, "?": {},
	}

	caseUndefinedQuantifiers = map[string]struct{}{
		"viel": {}, "nichts": {}, "nix": {}, "wenig": {}, "allerlei": {},
	}

	caseInterrogativeParticles = map[string]struct{}{
		"was": {}, "wodurch": {}, "wofür": {}, "womit": {}, "woran": {},
		"worauf": {}, "woraus": {}, "wovon": {}, "wie": {},
	}

	casePossessiveIndicators = map[string]struct{}{
		"einer": {}, "eines": {}, "der": {}, "des": {}, "dieser": {}, "dieses": {},
	}

	caseDasVerbExceptions = map[string]struct{}{
		"nur": {}, "sogar": {}, "auch": {}, "die": {}, "alle": {}, "viele": {}, "zu": {},
	}

	caseColonQuestionWords = map[string]struct{}{
		"warum": {}, "wieso": {}, "weshalb": {}, "wer": {}, "was": {},
		"wann": {}, "wo": {}, "wie": {}, "wozu": {},
	}

	caseColonQuestionConjunctions = map[string]struct{}{
		"und": {}, "oder": {}, "aber": {}, "denn": {},
	}

	caseAntiOnce  sync.Once
	caseAntiRules []*disambigrules.DisambiguationPatternRule

	caseExcPatternsOnce sync.Once
	caseExcPatterns     [][]*patterns.StringMatcher
)

func caseAntiPatternRules() []*disambigrules.DisambiguationPatternRule {
	caseAntiOnce.Do(func() {
		aps := CaseRuleAntiPatterns
		caseAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			caseAntiRules = append(caseAntiRules, rule)
		}
	})
	return caseAntiRules
}

func (r *CaseRule) getSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := caseAntiPatternRules()
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

func (r *CaseRule) lookup(word string) *languagetool.AnalyzedTokenReadings {
	if r != nil && r.Lookup != nil {
		return r.Lookup(word)
	}
	return nil
}

func (r *CaseRule) isMisspelled(lower string) bool {
	if r != nil && r.IsMisspelled != nil {
		return r.IsMisspelled(lower)
	}
	// Java: language.getDefaultSpellingRule().isMisspelled. WireCaseRule sets FilterDict.
	if FilterDictAvailable() {
		return FilterDictIsMisspelled(lower)
	}
	// Without dict: treat as not misspelled (Java always has speller; untagged twin path).
	return false
}

func (r *CaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	imm := r.getSentenceWithImmunization(sentence)
	tokens := imm.GetTokensWithoutWhitespace()
	if len(tokens) <= 1 {
		return nil
	}
	// Morph only (Java); untagged AnalyzePlain fails closed without inventing case errors.
	return r.matchMorph(tokens, sentence)
}

func (r *CaseRule) matchMorph(tokens []*languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	prevTokenIsDas := false
	isPrecededByModalOrAuxiliary := false
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		// sentence start
		if at := tokens[i].GetAnalyzedToken(0); at != nil {
			if pt := at.GetPOSTag(); pt != nil && *pt == languagetool.SentenceStartTagName {
				continue
			}
		}
		if tokens[i].IsSentenceStart() {
			continue
		}
		if i == 1 {
			prevTokenIsDas = isCaseNounIndicator(tokens[1].GetToken())
			continue
		}
		if i > 0 && (isCaseSalutation(tokens[i-1].GetToken()) || isCaseCompany(tokens[i-1].GetToken())) {
			continue
		}
		// 1.1 Technische Dokumentation
		if i > 2 && tokens[i-1] != nil && tokens[i-2] != nil && tokens[i-3] != nil &&
			caseNumeralsEN.MatchString(tokens[i-1].GetToken()) &&
			tokens[i-2].GetToken() == "." &&
			caseNumeralsEN.MatchString(tokens[i-3].GetToken()) {
			continue
		}

		analyzedToken := tokens[i]
		token := analyzedToken.GetToken()
		isBaseform := analyzedToken.GetReadingsLength() >= 1 && atrHasLemma(analyzedToken, token)
		firstPos := firstPOSTag(analyzedToken)
		if (firstPos == "" || HasReadingOfType(analyzedToken, POSVerb)) && isBaseform {
			nextTokenIsPersonalOrReflexivePronoun := false
			if i < len(tokens)-1 && tokens[i+1] != nil {
				nextToken := tokens[i+1]
				nextTokenIsPersonalOrReflexivePronoun = nextToken.HasPartialPosTag("PRO:PER") ||
					nextToken.GetToken() == "sich" || nextToken.GetToken() == "Sie"
				if nextToken.HasPosTag("PKT") {
					continue
				}
				// Java operator precedence:
				// (prevTokenIsDas && (DAS_VERB_EXCEPTIONS || relativeClause)) || (i>1 && VER:AUX|MOD)
				// The AUX/MOD arm is independent of prevTokenIsDas so baseform verbs after
				// modal/auxiliary skip the rest of the iteration (including uppercase match).
				if (prevTokenIsDas &&
					(isInSet(nextToken.GetToken(), caseDasVerbExceptions) ||
						isFollowedByRelativeOrSubordinateClause(i, tokens))) ||
					(i > 1 && hasPartialTagCase(tokens[i-2], "VER:AUX", "VER:MOD")) {
					continue
				}
			}
			if isPrevProbablyRelativePronoun(tokens, i) ||
				(prevTokenIsDas && getTokensWithPosTagStartingWithCount(tokens, "VER") == 1) {
				continue
			}
			r.potentiallyAddLowercaseMatch(&ruleMatches, tokens[i], prevTokenIsDas, token, nextTokenIsPersonalOrReflexivePronoun, sentence)
		}
		prevTokenIsDas = isCaseNounIndicator(tokens[i].GetToken())
		if analyzedToken.MatchesPosTagRegex(`VER:(MOD|AUX):[1-3]:.*`) {
			isPrecededByModalOrAuxiliary = true
		}

		lowercaseReadings := r.lookup(strings.ToLower(token))
		if r.hasNounReading(analyzedToken) {
			if !r.isPotentialUpperCaseError(i, tokens, lowercaseReadings, isPrecededByModalOrAuxiliary) {
				continue
			}
		} else if analyzedToken.HasPosTagStartingWith("SUB:") &&
			i < len(tokens)-1 && tokens[i+1] != nil &&
			len(tokens[i+1].GetToken()) > 0 && unicode.IsLower([]rune(tokens[i+1].GetToken())[0]) &&
			tokens[i+1].MatchesPosTagRegex(`(VER:[123]:|PA2).+`) {
			continue
		}
		if firstPos == "" && lowercaseReadings == nil {
			continue
		}
		if firstPos == "" && lowercaseReadings != nil {
			lcFirst := firstPOSTag(lowercaseReadings)
			if lcFirst == "" || strings.HasSuffix(analyzedToken.GetToken(), "innen") {
				continue
			}
		}
		r.potentiallyAddUppercaseMatch(&ruleMatches, tokens, i, analyzedToken, token, lowercaseReadings, sentence)
	}
	return ruleMatches
}

func firstPOSTag(r *languagetool.AnalyzedTokenReadings) string {
	if r == nil {
		return ""
	}
	at := r.GetAnalyzedToken(0)
	if at == nil {
		return ""
	}
	pt := at.GetPOSTag()
	if pt == nil {
		return ""
	}
	return *pt
}

func atrHasLemma(r *languagetool.AnalyzedTokenReadings, lemma string) bool {
	// Java AnalyzedTokenReadings.hasLemma
	return r != nil && r.HasLemma(lemma)
}

func isCaseNounIndicator(tok string) bool {
	_, ok := caseNounIndicators[strings.ToLower(tok)]
	return ok
}

func isInSet(s string, set map[string]struct{}) bool {
	_, ok := set[s]
	return ok
}

func getTokensWithPosTagStartingWithCount(tokens []*languagetool.AnalyzedTokenReadings, partial string) int {
	n := 0
	for _, t := range tokens {
		if t != nil && t.HasPosTagStartingWith(partial) {
			n++
		}
	}
	return n
}

func (r *CaseRule) isPotentialUpperCaseError(pos int, tokens []*languagetool.AnalyzedTokenReadings, lowercaseReadings *languagetool.AnalyzedTokenReadings, isPrecededByModalOrAuxiliary bool) bool {
	if pos <= 1 {
		return false
	}
	if tokens[pos-1] != nil && tokens[pos-1].GetToken() == "zu" &&
		tokens[pos] != nil && !tokens[pos].MatchesPosTagRegex(`.*(NEU|MAS|FEM)$`) &&
		lowercaseReadings != nil &&
		lowercaseReadings.HasPosTagStartingWith("VER:INF") {
		return true
	}
	if tokens[pos] != nil && caseVerhaltenRE.MatchString(tokens[pos].GetToken()) {
		return false
	}
	isPotentialError := pos < len(tokens)-3 &&
		tokens[pos+1] != nil && tokens[pos+1].GetToken() == "," &&
		tokens[pos+2] != nil && isInSet(tokens[pos+2].GetToken(), caseInterrogativeParticles) &&
		tokens[pos-1] != nil && tokens[pos-1].HasPosTagStartingWith("VER:MOD") &&
		!tokens[pos-1].HasAnyLemma("mögen") &&
		tokens[pos+3] != nil && tokens[pos+3].GetToken() != "zum"
	if !isPotentialError &&
		lowercaseReadings != nil &&
		tokens[pos] != nil &&
		tokens[pos-1] != nil &&
		(tokens[pos].HasPartialPosTag("SUB:NOM:SIN:NEU:INF") || tokens[pos].HasPartialPosTag("SUB:DAT:PLU:")) &&
		(tokens[pos-1].GetToken() == "zu" || hasPartialTagCase(tokens[pos-1], "SUB", "EIG", "VER:AUX:3:", "ADV:TMP", "ABK")) {
		if lowercaseReadings.HasPosTag("PA2:PRD:GRU:VER") &&
			!tokens[pos-1].HasPosTagStartingWith("VER:AUX:3") &&
			!lowercaseReadings.HasPosTag("VER:3:PLU:PRT:NON") {
			isPotentialError = true
		}
		if (pos >= len(tokens)-2 || (tokens[pos+1] != nil && tokens[pos+1].GetToken() == ",")) &&
			(tokens[pos-1].GetToken() == "zu" || isPrecededByModalOrAuxiliary) &&
			strings.HasPrefix(tokens[pos].GetToken(), "Über") &&
			(lowercaseReadings.HasPartialPosTag("VER:INF:") || lowercaseReadings.HasPartialPosTag("PA2:PRD:GRU:VER")) {
			isPotentialError = true
		}
	}
	return isPotentialError
}

func isPrevProbablyRelativePronoun(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	return i >= 3 &&
		tokens[i-1] != nil && tokens[i-1].GetToken() == "das" &&
		tokens[i-2] != nil && tokens[i-2].GetToken() == "," &&
		tokens[i-3] != nil && tokens[i-3].MatchesPosTagRegex(`SUB:...:SIN:NEU`)
}

func isCaseSalutation(token string) bool {
	switch token {
	case "Herr", "Hr", "Herrn", "Frau", "Fr", "Fräulein":
		return true
	}
	return false
}

func isCaseCompany(token string) bool {
	switch token {
	case "Firma", "Familie", "Unternehmen", "Firmen", "Bäckerei", "Metzgerei", "Fa":
		return true
	}
	return false
}

func (r *CaseRule) hasNounReading(readings *languagetool.AnalyzedTokenReadings) bool {
	if readings == nil {
		return false
	}
	if readings.HasPosTagStartingWith("ABK") && readings.HasPartialPosTag("SUB") {
		return true
	}
	tok := strings.ReplaceAll(readings.GetToken(), "\u00AD", "")
	allReadings := r.lookup(tok)
	if allReadings != nil {
		for _, reading := range allReadings.GetReadings() {
			if reading == nil {
				continue
			}
			pt := reading.GetPOSTag()
			if pt != nil && strings.Contains(*pt, "SUB:") && !strings.Contains(*pt, ":ADJ") {
				return true
			}
		}
	}
	return false
}

func (r *CaseRule) potentiallyAddLowercaseMatch(ruleMatches *[]*rules.RuleMatch, tokenReadings *languagetool.AnalyzedTokenReadings, prevTokenIsDas bool, token string, nextTokenIsPersonalOrReflexivePronoun bool, sentence *languagetool.AnalyzedSentence) {
	if prevTokenIsDas &&
		!nextTokenIsPersonalOrReflexivePronoun &&
		token != "" && unicode.IsLower([]rune(token)[0]) &&
		!isInSet(token, caseRuleSubstVerbenExceptions) &&
		tokenReadings.HasPosTagStartingWith("VER:INF") &&
		!tokenReadings.IsIgnoredBySpeller() &&
		!tokenReadings.IsImmunized() {
		r.addRuleMatch(ruleMatches, sentence, caseLowercaseMessage, tokenReadings, tools.UppercaseFirstChar(tokenReadings.GetToken()))
	}
}

func (r *CaseRule) potentiallyAddUppercaseMatch(ruleMatches *[]*rules.RuleMatch, tokens []*languagetool.AnalyzedTokenReadings, i int, analyzedToken *languagetool.AnalyzedTokenReadings, token string, lowercaseReadings *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) {
	if token == "" {
		return
	}
	isUpperFirst := unicode.IsUpper([]rune(token)[0])
	lcWord := tools.LowercaseFirstChar(tokens[i].GetToken())
	if !(isUpperFirst &&
		utf16LenDE(token) > 1 &&
		!tokens[i].IsIgnoredBySpeller() &&
		!tokens[i].IsImmunized() &&
		!isInSet(tokens[i-1].GetToken(), caseSentenceStartExceptions) &&
		!isInSet(token, caseRuleExceptionsWords) &&
		!tools.IsAllUppercase(token) &&
		!r.isLanguage(i, tokens, token) &&
		!isProbablyCity(i, tokens, token) &&
		!hasProperNounReading(analyzedToken) &&
		!analyzedToken.IsSentenceEnd() &&
		!isEllipsisCase(i, tokens) &&
		!isNumberingCase(i, tokens) &&
		!r.isNominalization(i, tokens, token, lowercaseReadings) &&
		!r.isAdverbAndNominalization(i, tokens) &&
		!r.isSpecialCase(i, tokens) &&
		!r.isAdjectiveAsNoun(i, tokens, lowercaseReadings) &&
		!isSingularImperative(lowercaseReadings, tokens[i]) &&
		!isExceptionPhrase(i, tokens) &&
		!(i == 2 && tokens[i-1] != nil && tokens[i-1].GetToken() == "“") &&
		!isCaseTypo(tokens[i].GetToken()) &&
		!followedByGenderGap(tokens, i) &&
		!isNounWithVerbReading(i, tokens) &&
		!isInvisibleSeparator(i-1, tokens) &&
		!r.isMisspelled(lcWord)) {
		return
	}
	if tokens[i-1] != nil && tokens[i-1].GetToken() == ":" {
		if isQuestionEquivalentAfterColon(i, tokens) {
			return
		}
		subarray := tokens[:i]
		if isVerbFollowing(i, tokens, lowercaseReadings) || getTokensWithPosTagStartingWithCount(subarray, "VER") == 0 {
			// no match
		} else {
			r.addRuleMatch(ruleMatches, sentence, caseColonMessage, tokens[i], lcWord)
		}
		return
	}
	r.addRuleMatch(ruleMatches, sentence, caseUppercaseMessage, tokens[i], lcWord)
}

func hasProperNounReading(r *languagetool.AnalyzedTokenReadings) bool {
	// Java: GermanHelper.hasReadingOfType(analyzedToken, POSType.PROPER_NOUN)
	// (EIG in a ≥3-part STTS tag via AnalyzedGermanToken — not bare HasPosTagStartingWith)
	return HasReadingOfType(r, POSProperNoun)
}

func followedByGenderGap(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	return i+2 < len(tokens) && tokens[i+1] != nil && tokens[i+1].GetToken() == ":" &&
		tokens[i+2] != nil && (tokens[i+2].GetToken() == "in" || tokens[i+2].GetToken() == "innen")
}

func isCaseTypo(token string) bool {
	return caseTwoUppercase.MatchString(token)
}

func isSingularImperative(lowercaseReadings, token *languagetool.AnalyzedTokenReadings) bool {
	return lowercaseReadings != nil && lowercaseReadings.HasPosTagStartingWith("VER:IMP:SIN") &&
		token != nil && token.GetToken() != "Ein" && token.GetToken() != "Eine"
}

func isNounWithVerbReading(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	return tokens[i] != nil && tokens[i].HasPosTagStartingWith("SUB") &&
		tokens[i].HasPosTagStartingWith("VER:INF")
}

func isInvisibleSeparator(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return false
	}
	t := tokens[i].GetToken()
	return t != "" && []rune(t)[0] == '\u2063'
}

func isVerbFollowing(i int, tokens []*languagetool.AnalyzedTokenReadings, lowercaseReadings *languagetool.AnalyzedTokenReadings) bool {
	sub := make([]*languagetool.AnalyzedTokenReadings, len(tokens)-i)
	copy(sub, tokens[i:])
	if lowercaseReadings != nil && len(sub) > 0 {
		sub[0] = lowercaseReadings
	}
	return getTokensWithPosTagStartingWithCount(sub, "VER:") != 0
}

func isColonQuestionWord(word string) bool {
	_, ok := caseColonQuestionWords[strings.ToLower(word)]
	return ok
}

func (r *CaseRule) addRuleMatch(ruleMatches *[]*rules.RuleMatch, sentence *languagetool.AnalyzedSentence, msg string, tokenReadings *languagetool.AnalyzedTokenReadings, fixedWord string) {
	// Java: new RuleMatch(rule, sentence, from, to, msg) + setSuggestedReplacement only (no shortMessage).
	rm := rules.NewRuleMatch(r, sentence, tokenReadings.GetStartPos(), tokenReadings.GetEndPos(), msg)
	rm.SetSuggestedReplacement(fixedWord)
	*ruleMatches = append(*ruleMatches, rm)
}

func isNumberingCase(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	if i < 2 || tokens[i-1] == nil || tokens[i-2] == nil {
		return false
	}
	p := tokens[i-1].GetToken()
	if p != ")" && p != "]" {
		return false
	}
	if !caseNumeralsEN.MatchString(tokens[i-2].GetToken()) {
		return false
	}
	if i > 3 && tokens[i-3] != nil && tokens[i-3].GetToken() == "(" &&
		tokens[i-4] != nil && tokens[i-4].HasPosTagStartingWith("SUB:") {
		return false
	}
	return true
}

func isEllipsisCase(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	if i < 1 || tokens[i-1] == nil {
		return false
	}
	p := tokens[i-1].GetToken()
	if p != "]" && p != ")" {
		return false
	}
	if i == 4 && tokens[i-2] != nil && tokens[i-2].GetToken() == "…" {
		return true
	}
	if i == 6 && tokens[i-2] != nil && tokens[i-2].GetToken() == "." {
		return true
	}
	return false
}

func (r *CaseRule) isNominalization(i int, tokens []*languagetool.AnalyzedTokenReadings, token string, lowercaseReadings *languagetool.AnalyzedTokenReadings) bool {
	var nextReadings *languagetool.AnalyzedTokenReadings
	if i < len(tokens)-1 {
		nextReadings = tokens[i+1]
	}
	if !(tools.StartsWithUppercase(token) && !r.isNumber(token) &&
		!(r.hasNounReading(nextReadings) || (nextReadings != nil && isNumericTokenCase(nextReadings.GetToken()))) &&
		!caseAllnmRE.MatchString(token)) {
		return false
	}
	if lowercaseReadings != nil && lowercaseReadings.HasPosTag("PRP:LOK+TMP+CAU:DAT+AKK") {
		return false
	}
	var prevToken, prevPrevToken, prevPrevPrevToken *languagetool.AnalyzedTokenReadings
	if i > 0 {
		prevToken = tokens[i-1]
	}
	if i >= 2 {
		prevPrevToken = tokens[i-2]
	}
	if i >= 3 {
		prevPrevPrevToken = tokens[i-3]
	}
	prevTokenStr := ""
	if prevToken != nil {
		prevTokenStr = prevToken.GetToken()
	}
	if (prevTokenStr == "und" || prevTokenStr == "oder" || prevTokenStr == "beziehungsweise") && prevPrevToken != nil &&
		((tokens[i].HasPartialPosTag("SUB") && tokens[i].HasPartialPosTag(":ADJ")) ||
			(prevPrevToken.HasPartialPosTag("SUB") && !r.hasNounReading(nextReadings) &&
				lowercaseReadings != nil && lowercaseReadings.HasPartialPosTag("ADJ") && prevTokenStr != ",")) {
		return true
	}
	if lowercaseReadings != nil && lowercaseReadings.HasPosTag("PA1:PRD:GRU:VER") {
		return false
	}
	return (prevToken != nil && caseIrgendEtcRE.MatchString(prevTokenStr) && tokens[i].HasPartialPosTag("SUB")) ||
		r.isNumber(prevTokenStr) ||
		(hasPartialTagCase(prevToken, "ART", "PRO:") &&
			!(((i < 4 && len(tokens) > 4) || (prevToken != nil && prevToken.GetReadingsLength() == 1) ||
				(prevPrevToken != nil && atrHasLemma(prevPrevToken, "sein"))) &&
				prevToken.HasPosTagStartingWith("PRO:PER:NOM:")) &&
			!prevToken.HasPartialPosTag(":STD")) ||
		(hasPartialTagCase(prevPrevPrevToken, "ART") && hasPartialTagCase(prevPrevToken, "PRP") && hasPartialTagCase(prevToken, "SUB")) ||
		(hasPartialTagCase(prevPrevToken, "PRO:", "PRP") && hasPartialTagCase(prevToken, "ADJ", "ADV", "PA2", "PA1")) ||
		(hasPartialTagCase(prevPrevPrevToken, "PRO:", "PRP") && hasPartialTagCase(prevPrevToken, "ADJ", "ADV") && hasPartialTagCase(prevToken, "ADJ", "ADV", "PA2")) ||
		(tokens[i].HasPosTagStartingWith("SUB:") && hasPartialTagCase(prevToken, "GEN") && !hasPartialTagCase(nextReadings, "PKT"))
}

func isNumericTokenCase(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (r *CaseRule) isNumber(token string) bool {
	if isNumericTokenCase(token) {
		return true
	}
	lookup := r.lookup(tools.LowercaseFirstChar(token))
	return lookup != nil && lookup.HasPosTag("ZAL")
}

func (r *CaseRule) isAdverbAndNominalization(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	prevPrevToken := ""
	if i > 1 && tokens[i-2] != nil {
		prevPrevToken = tokens[i-2].GetToken()
	}
	var prevToken *languagetool.AnalyzedTokenReadings
	if i > 0 {
		prevToken = tokens[i-1]
	}
	token := tokens[i].GetToken()
	var nextReadings *languagetool.AnalyzedTokenReadings
	if i < len(tokens)-1 {
		nextReadings = tokens[i+1]
	}
	return strings.EqualFold(prevPrevToken, "das") && hasPartialTagCase(prevToken, "ADV") &&
		tools.StartsWithUppercase(token) && !r.hasNounReading(nextReadings)
}

func hasPartialTagCase(token *languagetool.AnalyzedTokenReadings, posTags ...string) bool {
	if token == nil {
		return false
	}
	for _, posTag := range posTags {
		if token.HasPartialPosTag(posTag) {
			return true
		}
	}
	return false
}

func (r *CaseRule) isSpecialCase(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	prevToken := ""
	if i > 0 && tokens[i-1] != nil {
		prevToken = tokens[i-1].GetToken()
	}
	token := tokens[i].GetToken()
	var nextReadings *languagetool.AnalyzedTokenReadings
	if i < len(tokens)-1 {
		nextReadings = tokens[i+1]
	}
	return strings.EqualFold(prevToken, "im") && token == "Allgemeinen" && !r.hasNounReading(nextReadings)
}

func (r *CaseRule) isAdjectiveAsNoun(i int, tokens []*languagetool.AnalyzedTokenReadings, lowercaseReadings *languagetool.AnalyzedTokenReadings) bool {
	var prevToken, nextReadings *languagetool.AnalyzedTokenReadings
	if i > 0 {
		prevToken = tokens[i-1]
	}
	if i < len(tokens)-1 {
		nextReadings = tokens[i+1]
	}
	var prevLowercaseReadings *languagetool.AnalyzedTokenReadings
	if i > 1 && prevToken != nil && isInSet(tokens[i-2].GetToken(), caseSentenceStartExceptions) {
		prevLowercaseReadings = r.lookup(strings.ToLower(prevToken.GetToken()))
	}

	isPossiblyFollowedByInfinitive := nextReadings != nil && nextReadings.GetToken() == "zu"
	isFollowedByInfinitive := nextReadings != nil && !isPossiblyFollowedByInfinitive && nextReadings.HasPartialPosTag("EIZ")
	isFollowedByPossessiveIndicator := nextReadings != nil && isInSet(nextReadings.GetToken(), casePossessiveIndicators)

	isUndefQuantifier := prevToken != nil && isInSet(strings.ToLower(prevToken.GetToken()), caseUndefinedQuantifiers)
	isPrevDeterminer := prevToken != nil &&
		(hasPartialTagCase(prevToken, "ART", "PRP", "ZAL") || hasPartialTagCase(prevLowercaseReadings, "ART", "PRP", "ZAL")) &&
		!prevToken.HasPartialPosTag(":STD")
	isPrecededByVerb := prevToken != nil && prevToken.MatchesPosTagRegex(`VER:(MOD:|AUX:)?[1-3]:.*`) && !atrHasLemma(prevToken, "sein")

	if !isPrevDeterminer && !isUndefQuantifier && !(isPossiblyFollowedByInfinitive || isFollowedByInfinitive) &&
		!(isPrecededByVerb && lowercaseReadings != nil && hasPartialTagCase(lowercaseReadings, "ADJ:", "PA") && nextReadings != nil &&
			nextReadings.GetToken() != "und" && nextReadings.GetToken() != "oder" && nextReadings.GetToken() != ",") &&
		!(isFollowedByPossessiveIndicator && hasPartialTagCase(lowercaseReadings, "ADJ", "VER")) &&
		!(prevToken != nil && prevToken.HasPosTag("KON:UNT") && !r.hasNounReading(nextReadings) && nextReadings != nil && !nextReadings.HasPosTag("KON:NEB")) {
		var prevPrevToken *languagetool.AnalyzedTokenReadings
		if i > 1 && prevToken != nil && prevToken.HasPartialPosTag("ADJ") {
			prevPrevToken = tokens[i-2]
		}
		if !isPrecededByVerb && lowercaseReadings != nil && prevToken != nil {
			if prevToken.HasPartialPosTag("SUB:") && lowercaseReadings.MatchesPosTagRegex(`(ADJ|PA2):GEN:PLU:MAS:GRU:SOL.*`) {
				return nextReadings != nil && !nextReadings.HasPartialPosTag("SUB:")
			} else if nextReadings != nil && nextReadings.GetReadingsLength() == 1 &&
				prevToken.HasPosTagStartingWith("PRO:PER:NOM:") && nextReadings.HasPosTag("ADJ:PRD:GRU") {
				return true
			}
		}
		if !hasPartialTagCase(prevPrevToken, "ART", "PRP", "ZAL") {
			return false
		}
	}

	for _, reading := range tokens[i].GetReadings() {
		if reading == nil {
			continue
		}
		var posTag string
		if pt := reading.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		nextTok := ""
		if nextReadings != nil {
			nextTok = nextReadings.GetToken()
		}
		if (posTag == "" || strings.Contains(posTag, "ADJ")) && !r.hasNounReading(nextReadings) && !isNumericTokenCase(nextTok) {
			if posTag == "" && hasPartialTagCase(lowercaseReadings, "PRP:LOK", "PA2:PRD:GRU:VER", "PA1:PRD:GRU:VER", "ADJ:PRD:KOM", "ADV:TMP") {
				// skip
			} else {
				return true
			}
		}
	}
	return false
}

func (r *CaseRule) isLanguage(i int, tokens []*languagetool.AnalyzedTokenReadings, token string) bool {
	base := strings.TrimSuffix(strings.TrimSuffix(token, "n"), "e")
	maybeLanguage := (strings.HasSuffix(token, "sch") && IsLanguageName(token)) || IsLanguageName(base)
	var prevToken, nextReadings *languagetool.AnalyzedTokenReadings
	if i > 0 {
		prevToken = tokens[i-1]
	}
	if i < len(tokens)-1 {
		nextReadings = tokens[i+1]
	}
	return maybeLanguage && (!r.hasNounReading(nextReadings) || (prevToken != nil && prevToken.GetToken() == "auf"))
}

func isProbablyCity(i int, tokens []*languagetool.AnalyzedTokenReadings, token string) bool {
	if token != "Klein" && token != "Groß" && token != "Neu" {
		return false
	}
	if i >= len(tokens)-1 || tokens[i+1] == nil {
		return false
	}
	next := tokens[i+1]
	return !next.IsTagged() || next.HasPosTagStartingWith("EIG")
}

func isFollowedByRelativeOrSubordinateClause(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	if i >= len(tokens)-4 {
		return false
	}
	return tokens[i+1] != nil && tokens[i+1].GetToken() == "," &&
		tokens[i+2] != nil &&
		(isInSet(tokens[i+2].GetToken(), caseInterrogativeParticles) || tokens[i+2].HasPosTag("KON:UNT"))
}

func caseExceptionPatterns() [][]*patterns.StringMatcher {
	caseExcPatternsOnce.Do(func() {
		ex := CaseRuleExceptions()
		caseExcPatterns = make([][]*patterns.StringMatcher, 0, len(ex))
		for phrase := range ex {
			parts := strings.Fields(phrase)
			if len(parts) == 0 {
				continue
			}
			ms := make([]*patterns.StringMatcher, len(parts))
			ok := true
			for j, p := range parts {
				// phrases may be regexps; skip invalid rather than panic
				func() {
					defer func() {
						if recover() != nil {
							ok = false
						}
					}()
					ms[j] = patterns.NewStringMatcherRegexp(p)
				}()
				if !ok {
					break
				}
			}
			if ok {
				caseExcPatterns = append(caseExcPatterns, ms)
			}
		}
	})
	return caseExcPatterns
}

func isExceptionPhrase(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return false
	}
	tok := tokens[i].GetToken()
	for _, pats := range caseExceptionPatterns() {
		for j, p := range pats {
			if p == nil || !p.Matches(tok) {
				continue
			}
			startIndex := i - j
			if CaseRuleCompareListsMatchers(tokens, startIndex, startIndex+len(pats)-1, pats) {
				return true
			}
		}
	}
	return false
}

// CaseRuleCompareListsMatchers ports CaseRule.compareLists with StringMatcher.
func CaseRuleCompareListsMatchers(tokens []*languagetool.AnalyzedTokenReadings, startIndex, endIndex int, patterns []*patterns.StringMatcher) bool {
	if startIndex < 0 {
		return false
	}
	ii := 0
	for j := startIndex; j <= endIndex; j++ {
		if ii >= len(patterns) || j >= len(tokens) || tokens[j] == nil || patterns[ii] == nil || !patterns[ii].Matches(tokens[j].GetToken()) {
			return false
		}
		ii++
	}
	return true
}

// CaseRuleCompareLists ports CaseRule.compareLists for regexp patterns (tests).
func CaseRuleCompareLists(tokens []*languagetool.AnalyzedTokenReadings, startIndex, endIndex int, patterns []*regexp.Regexp) bool {
	if startIndex < 0 || endIndex >= len(tokens) || endIndex-startIndex+1 != len(patterns) {
		return false
	}
	for i := 0; i < len(patterns); i++ {
		if tokens[startIndex+i] == nil || !patterns[i].MatchString(tokens[startIndex+i].GetToken()) {
			return false
		}
	}
	return true
}

func isQuestionEquivalentAfterColon(i int, tokens []*languagetool.AnalyzedTokenReadings) bool {
	if i >= len(tokens)-1 || tokens[i] == nil || tokens[i+1] == nil {
		return false
	}
	word := tokens[i].GetToken()
	next := tokens[i+1].GetToken()
	if isColonQuestionWord(word) && next == "?" {
		return true
	}
	if _, ok := caseColonQuestionConjunctions[strings.ToLower(word)]; ok &&
		i < len(tokens)-2 && tokens[i+1] != nil && tokens[i+2] != nil &&
		isColonQuestionWord(tokens[i+1].GetToken()) && tokens[i+2].GetToken() == "?" {
		return true
	}
	return false
}

// isDet is used by AgreementRule helpers (determiner surface check for open compounds).
var caseDeterminers = map[string]struct{}{
	"der": {}, "die": {}, "das": {}, "dem": {}, "den": {}, "des": {},
	"ein": {}, "eine": {}, "einem": {}, "einen": {}, "einer": {}, "eines": {},
	"mein": {}, "meine": {}, "meinem": {}, "meinen": {}, "meiner": {}, "meines": {},
	"dein": {}, "deine": {}, "deinem": {}, "deinen": {}, "deiner": {}, "deines": {},
	"sein": {}, "seine": {}, "seinem": {}, "seinen": {}, "seiner": {}, "seines": {},
	"ihr": {}, "ihre": {}, "ihrem": {}, "ihren": {}, "ihrer": {}, "ihres": {},
	"unser": {}, "unsere": {}, "unserem": {}, "unseren": {}, "unserer": {}, "unseres": {},
	"kein": {}, "keine": {}, "keinem": {}, "keinen": {}, "keiner": {}, "keines": {},
	"dieser": {}, "diese": {}, "dieses": {}, "diesem": {}, "diesen": {},
	"jener": {}, "jene": {}, "jenes": {}, "jenem": {}, "jenen": {},
	"alle": {}, "allem": {}, "allen": {}, "aller": {}, "alles": {},
	"viele": {}, "vieler": {}, "wenige": {},
}

func isDet(w string) bool {
	_, ok := caseDeterminers[strings.ToLower(w)]
	return ok
}
