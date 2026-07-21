package de

import (
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// MissingCommaRelativeClauseRule ports org.languagetool.rules.de.MissingCommaRelativeClauseRule.
// Behind=false → COMMA_IN_FRONT_RELATIVE_CLAUSE; Behind=true → COMMA_BEHIND_RELATIVE_CLAUSE.
// Morph/POS only (no surface invent for untagged AnalyzePlain).
// Java: Category HILFESTELLUNG_KOMMASETZUNG ("Kommasetzung").
type MissingCommaRelativeClauseRule struct {
	Messages map[string]string
	Behind   bool
	Category *rules.Category
}

func newMissingCommaCategory(messages map[string]string) *rules.Category {
	// Java: new Category(HILFESTELLUNG_KOMMASETZUNG, "Kommasetzung", INTERNAL, true)
	return rules.NewCategoryFull(rules.NewCategoryId("HILFESTELLUNG_KOMMASETZUNG"), "Kommasetzung", rules.CategoryInternal, true, "")
}

func NewMissingCommaRelativeClauseRule(messages map[string]string) *MissingCommaRelativeClauseRule {
	return &MissingCommaRelativeClauseRule{Messages: messages, Category: newMissingCommaCategory(messages)}
}

func NewMissingCommaRelativeClauseRuleBehind(messages map[string]string) *MissingCommaRelativeClauseRule {
	return &MissingCommaRelativeClauseRule{Messages: messages, Behind: true, Category: newMissingCommaCategory(messages)}
}

func (r *MissingCommaRelativeClauseRule) GetID() string {
	// Java IDs (not the previous Go stand-in names)
	if r != nil && r.Behind {
		return "COMMA_BEHIND_RELATIVE_CLAUSE"
	}
	return "COMMA_IN_FRONT_RELATIVE_CLAUSE"
}

// GetDescription ports MissingCommaRelativeClauseRule.getDescription.
func (r *MissingCommaRelativeClauseRule) GetDescription() string {
	if r != nil && r.Behind {
		return "Fehlendes Komma nach Relativsatz"
	}
	return "Fehlendes Komma vor Relativsatz"
}

func (r *MissingCommaRelativeClauseRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

var (
	// Java: Pattern.compile("[,;.:?•!-–—’'\"„“”…»«‚‘›‹()\\/\\[\\]]")
	// Note: between '!' (U+0021) and en-dash (U+2013) sits ASCII '-' which is a
	// character-class range operator → matches every single-unit code point from
	// '!' through en-dash (letters, digits, most punctuation). Twin that quirk
	// bug-for-bug; multi-char tokens still fail full-match (.matches()).
	// Characters outside that range that appear after en-dash in Java are listed
	// explicitly (em-dash, guillemets, brackets, …, •).
	missingCommaMarksRE = regexp.MustCompile(`^[•\x{0021}-\x{2013}\x{2014}’'"„“”…»«‚‘›‹()/\[\]\\]$`)
	missingCommaPronounRE = regexp.MustCompile(`^(d(e[mnr]|ie|as|e([nr]|ss)en)|welche[mrs]?|wessen|was)$`)

	missingCommaAntiOnce  sync.Once
	missingCommaAntiRules []*disambigrules.DisambiguationPatternRule
)

func missingCommaAntiPatternRules() []*disambigrules.DisambiguationPatternRule {
	missingCommaAntiOnce.Do(func() {
		aps := MissingCommaRelativeAntiPatterns
		missingCommaAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			missingCommaAntiRules = append(missingCommaAntiRules, rule)
		}
	})
	return missingCommaAntiRules
}

func (r *MissingCommaRelativeClauseRule) getSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := missingCommaAntiPatternRules()
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

func (r *MissingCommaRelativeClauseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	imm := r.getSentenceWithImmunization(sentence)
	tokens := imm.GetTokensWithoutWhitespace()
	if len(tokens) <= 1 {
		return nil
	}
	// Morph path only (Java POS-gated). Untagged AnalyzePlain fails closed.
	if r.Behind {
		return r.matchMorphBehind(tokens, sentence)
	}
	return r.matchMorphFront(tokens, sentence)
}

func hasAnyVERTags(tokens []*languagetool.AnalyzedTokenReadings) bool {
	for _, t := range tokens {
		if t != nil && t.HasPosTagStartingWith("VER") {
			return true
		}
	}
	return false
}

func isSeparatorMissingComma(token string) bool {
	return missingCommaMarksRE.MatchString(token) || token == "und" || token == "oder"
}

func nextSeparatorMissingComma(tokens []*languagetool.AnalyzedTokenReadings, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i] != nil && isSeparatorMissingComma(tokens[i].GetToken()) {
			return i
		}
	}
	return len(tokens) - 1
}

func isPrpMissingComma(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && token.HasPosTagStartingWith("PRP:") && !token.IsImmunized()
}

func isVerbMissingComma(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	if n < 0 || n >= len(tokens) || tokens[n] == nil || tokens[n].IsImmunized() {
		return false
	}
	// Java isVerb: verbPattern && !zalEtc && !(VER:INF: && prev=="zu") && !immunized
	if !tokens[n].MatchesPosTagRegex(`(VER:[1-3]:|VER:.*:[1-3]:).*`) {
		return false
	}
	if tokens[n].MatchesPosTagRegex(`(ZAL|AD[JV]|ART|SUB|PRO:POS|PRP).*`) {
		return false
	}
	if tokens[n].HasPosTagStartingWith("VER:INF:") && n > 0 && tokens[n-1] != nil && tokens[n-1].GetToken() == "zu" {
		return false
	}
	return true
}

func isAnyVerbMissingComma(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	if n < 0 || n >= len(tokens) || tokens[n] == nil {
		return false
	}
	if tokens[n].HasPosTagStartingWith("VER:") {
		return true
	}
	if n < len(tokens)-1 && tokens[n+1] != nil {
		if tokens[n].GetToken() == "zu" && tokens[n+1].HasPosTagStartingWith("VER:INF:") {
			return true
		}
		if tokens[n].HasPosTag("NEG") && tokens[n+1].HasPosTagStartingWith("VER:") {
			return true
		}
	}
	return false
}

func verbPosMissingComma(tokens []*languagetool.AnalyzedTokenReadings, start, end int) []int {
	var verbs []int
	for i := start; i < end && i < len(tokens); i++ {
		if !isVerbMissingComma(tokens, i) {
			continue
		}
		if tokens[i].HasPosTagStartingWith("PA") {
			gender := getGenderMissingComma(tokens[i])
			sStr := "(ADJ|PA[12]):.*" + gender + ".*"
			j := i + 1
			for j < end && j < len(tokens) && tokens[j] != nil && tokens[j].MatchesPosTagRegex(sStr) {
				j++
			}
			if j < end && j < len(tokens) && tokens[j] != nil &&
				!tokens[j].MatchesPosTagRegex("(SUB|EIG):.*"+gender+".*") && !isPosTagUnknown(tokens[j]) {
				verbs = append(verbs, i)
			}
		} else {
			verbs = append(verbs, i)
		}
	}
	return verbs
}

func isPosTagUnknown(t *languagetool.AnalyzedTokenReadings) bool {
	if t == nil {
		return true
	}
	// Java isPosTagUnknown — treat untagged as unknown
	return !t.IsTagged()
}

func isKonUntMissingComma(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	if token.HasPosTag("KON:UNT") {
		return true
	}
	switch strings.ToLower(token.GetToken()) {
	case "wer", "wo", "wohin":
		return true
	}
	return false
}

func hasPotentialSubclause(tokens []*languagetool.AnalyzedTokenReadings, start, end int) int {
	verbs := verbPosMissingComma(tokens, start, end)
	if len(verbs) == 1 && end < len(tokens)-2 && verbs[0] == end-1 {
		nextEnd := nextSeparatorMissingComma(tokens, end+1)
		nextVerbs := verbPosMissingComma(tokens, end+1, nextEnd)
		if isKonUntMissingComma(tokens[start]) {
			if len(nextVerbs) > 1 || (len(nextVerbs) == 1 && nextVerbs[0] == end-1) {
				return verbs[0]
			}
		} else if len(nextVerbs) > 0 {
			return verbs[0]
		}
		return -1
	}
	if len(verbs) == 2 {
		if tokens[verbs[0]].MatchesPosTagRegex("VER:(MOD|AUX):.*") && tokens[verbs[1]].HasPosTagStartingWith("VER:INF:") {
			return verbs[0]
		}
		if tokens[verbs[0]].HasPosTagStartingWith("VER:AUX:") && tokens[verbs[1]].HasPosTagStartingWith("VER:PA2:") {
			return -1
		}
		if end == len(tokens)-1 && verbs[0] == end-2 &&
			tokens[verbs[0]].HasPosTagStartingWith("VER:INF:") && tokens[verbs[1]].HasPosTagStartingWith("VER:MOD:") {
			return -1
		}
	}
	if len(verbs) == 3 {
		// Java: MOD + (INF/PA2 then INF) or weder/noch INF pair → no subclause signal
		if tokens[verbs[0]].HasPosTagStartingWith("VER:MOD:") &&
			((verbs[2]-1 >= 0 && tokens[verbs[2]-1] != nil &&
				tokens[verbs[2]-1].MatchesPosTagRegex(`VER:(INF|PA2):.*`) &&
				tokens[verbs[2]].HasPosTagStartingWith("VER:INF:")) ||
				(verbs[1]-1 >= 0 && tokens[verbs[1]-1] != nil && tokens[verbs[1]-1].GetToken() == "weder" &&
					tokens[verbs[1]].HasPosTagStartingWith("VER:INF:") &&
					verbs[2]-1 >= 0 && tokens[verbs[2]-1] != nil && tokens[verbs[2]-1].GetToken() == "noch" &&
					tokens[verbs[1]].HasPosTagStartingWith("VER:INF:"))) {
			return -1
		}
	}
	if len(verbs) > 1 {
		return verbs[len(verbs)-1]
	}
	return -1
}

func isPronounMissingComma(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	if n < 1 || n >= len(tokens) || tokens[n] == nil {
		return false
	}
	return missingCommaPronounRE.MatchString(strings.ToLower(tokens[n].GetToken())) &&
		tokens[n-1] != nil && tokens[n-1].GetToken() != "sowie"
}

func getGenderMissingComma(token *languagetool.AnalyzedTokenReadings) string {
	if token == nil {
		return ""
	}
	var parts []string
	if token.MatchesPosTagRegex(".*:SIN:FEM.*") {
		parts = append(parts, "SIN:FEM")
	}
	if token.MatchesPosTagRegex(".*:SIN:MAS.*") {
		parts = append(parts, "SIN:MAS")
	}
	if token.MatchesPosTagRegex(".*:SIN:NEU.*") {
		parts = append(parts, "SIN:NEU")
	}
	if token.MatchesPosTagRegex(".*:PLU.*") {
		parts = append(parts, "PLU")
	}
	if len(parts) > 1 {
		return "(" + strings.Join(parts, "|") + ")"
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return ""
}

func matchesGenderMissingComma(gender string, tokens []*languagetool.AnalyzedTokenReadings, from, to int) bool {
	mStr := "(SUB|EIG):.*" + gender + ".*"
	if gender == "" {
		mStr = "PRO:DEM:.*SIN:NEU.*"
	}
	for i := to - 1; i >= from; i-- {
		if tokens[i] == nil {
			continue
		}
		if tokens[i].MatchesPosTagRegex(mStr) && (i != 1 || !tokens[i].HasPosTagStartingWith("VER:")) {
			return true
		}
	}
	return false
}

// isArticleWithoutSub ports Java isArticleWithoutSub.
func isArticleWithoutSubMissingComma(gender string, tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	if gender == "" || n < 1 || n >= len(tokens) || tokens[n] == nil || tokens[n-1] == nil {
		return false
	}
	return tokens[n].HasPosTagStartingWith("VER:") &&
		tokens[n-1].MatchesPosTagRegex(`(ADJ|PA[12]|PRO:POS):.*`+gender+`.*`)
}

// skipSubMissingComma ports Java skipSub — next SUB/EIG matching gender of token n.
func skipSubMissingComma(tokens []*languagetool.AnalyzedTokenReadings, n, to int) int {
	if n < 0 || n >= len(tokens) || tokens[n] == nil {
		return -1
	}
	gender := getGenderMissingComma(tokens[n])
	sSub := `(SUB|EIG):.*` + gender + `.*`
	for i := n + 1; i < to && i < len(tokens); i++ {
		if tokens[i] != nil && tokens[i].MatchesPosTagRegex(sSub) {
			return i
		}
	}
	return -1
}

// skipToSubMissingComma ports Java skipToSub.
func skipToSubMissingComma(gender string, tokens []*languagetool.AnalyzedTokenReadings, n, to int) int {
	if n+1 < len(tokens) && tokens[n+1] != nil &&
		tokens[n+1].MatchesPosTagRegex(`PA[12]:.*`+gender+`.*`) {
		return n + 1
	}
	for i := n + 1; i < to && i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		if tokens[i].MatchesPosTagRegex(`(ADJ|PA[12]):.*`+gender+`.*`) || isPosTagUnknown(tokens[i]) {
			return i
		}
		if tokens[i].HasPosTagStartingWith("ART") {
			i = skipSubMissingComma(tokens, i, to)
			if i < 0 {
				return i
			}
		}
	}
	return -1
}

// isArticleMissingComma ports Java isArticle.
func isArticleMissingComma(gender string, tokens []*languagetool.AnalyzedTokenReadings, from, to int) bool {
	if gender == "" {
		return false
	}
	sSub := `(SUB|EIG):.*` + gender + `.*`
	sAdj := `(ZAL|PRP:|KON:|ADV:|ADJ:PRD:|(ADJ|PA[12]|PRO:(POS|DEM|IND)):.*` + gender + `).*`
	for i := from + 1; i < to && i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		if tokens[i].MatchesPosTagRegex(sSub) || isPosTagUnknown(tokens[i]) {
			return true
		}
		if tokens[i].HasPosTagStartingWith("ART") || !tokens[i].MatchesPosTagRegex(sAdj) {
			if isArticleWithoutSubMissingComma(gender, tokens, i) {
				return true
			}
			skipTo := skipToSubMissingComma(gender, tokens, i, to)
			if skipTo > 0 {
				i = skipTo
			} else {
				return false
			}
		}
	}
	return to < len(tokens) && isArticleWithoutSubMissingComma(gender, tokens, to)
}

func missedCommaInFront(tokens []*languagetool.AnalyzedTokenReadings, start, end, lastVerb int) int {
	for i := start; i < lastVerb-1 && i < len(tokens); i++ {
		if tokens[i] == nil || tokens[i].IsImmunized() {
			continue
		}
		if !isPronounMissingComma(tokens, i) {
			continue
		}
		// Java: gender != null (getGender never returns null; empty string is allowed)
		gender := getGenderMissingComma(tokens[i])
		if !isAnyVerbMissingComma(tokens, i+1) &&
			matchesGenderMissingComma(gender, tokens, start, i) &&
			!isArticleMissingComma(gender, tokens, i, lastVerb) {
			return i
		}
	}
	return -1
}

// --- getCommaBehind helpers (Java 1:1) ---

func isTwoCombinedVerbsMissingComma(first, second *languagetool.AnalyzedTokenReadings) bool {
	return first != nil && second != nil &&
		first.MatchesPosTagRegex(`(VER:.*INF|.*PA[12]:).*`) &&
		second.HasPosTagStartingWith("VER:")
}

func isThreeCombinedVerbsMissingComma(tokens []*languagetool.AnalyzedTokenReadings, first, last int) bool {
	if first < 0 || last >= len(tokens) || first+1 >= len(tokens) {
		return false
	}
	return tokens[first] != nil && tokens[first+1] != nil && tokens[last] != nil &&
		tokens[first].MatchesPosTagRegex(`VER:(AUX|INF|PA[12]).*`) &&
		tokens[first+1].MatchesPosTagRegex(`VER:(.*INF|PA[12]).*`) &&
		tokens[last].MatchesPosTagRegex(`VER:(MOD|AUX).*`)
}

func isFourCombinedVerbsMissingComma(tokens []*languagetool.AnalyzedTokenReadings, first, last int) bool {
	if first < 0 || last >= len(tokens) || first+2 >= len(tokens) {
		return false
	}
	return tokens[first] != nil && tokens[first+1] != nil && tokens[first+2] != nil && tokens[last] != nil &&
		tokens[first].HasPartialPosTag("KJ2") && tokens[first+1].HasPartialPosTag("PA2") &&
		tokens[first+2].MatchesPosTagRegex(`VER:(.*INF|PA[12]).*`) &&
		tokens[last].MatchesPosTagRegex(`VER:(MOD|AUX).*`)
}

func isParMissingComma(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && (token.HasPosTagStartingWith("PA2:") || token.HasPosTagStartingWith("VER:PA2"))
}

func isInfinitivZuMissingComma(tokens []*languagetool.AnalyzedTokenReadings, last int) bool {
	if last < 1 || last >= len(tokens) || tokens[last] == nil || tokens[last-1] == nil {
		return false
	}
	return tokens[last-1].GetToken() == "zu" && tokens[last].MatchesPosTagRegex(`VER:.*INF.*`)
}

func isTwoPlusCombinedVerbsMissingComma(tokens []*languagetool.AnalyzedTokenReadings, first, last int) bool {
	if first < 0 || last-1 < 0 || last-1 >= len(tokens) || first >= len(tokens) {
		return false
	}
	return tokens[first] != nil && tokens[last-1] != nil &&
		tokens[first].MatchesPosTagRegex(`.*PA[12]:.*`) &&
		tokens[last-1].MatchesPosTagRegex(`VER:.*INF.*`)
}

func isKonAfterVerbMissingComma(tokens []*languagetool.AnalyzedTokenReadings, start, end int) bool {
	if start < 0 || start+1 >= len(tokens) || tokens[start] == nil || tokens[start+1] == nil {
		return false
	}
	if tokens[start].MatchesPosTagRegex(`VER:(MOD|AUX).*`) && tokens[start+1].MatchesPosTagRegex(`(KON|PRP).*`) {
		if start+3 == end {
			return true
		}
		for i := start + 2; i < end && i < len(tokens); i++ {
			if tokens[i] != nil && tokens[i].MatchesPosTagRegex(`(SUB|PRO:PER).*`) {
				return true
			}
		}
	}
	return false
}

func isSpecialPairMissingComma(tokens []*languagetool.AnalyzedTokenReadings, first, second int) bool {
	if first < 0 || second >= len(tokens) || first+1 >= len(tokens) || first+2 >= len(tokens) {
		return false
	}
	if first+3 >= second && tokens[first] != nil && tokens[first].MatchesPosTagRegex(`VER:.*INF.*`) {
		mid := tokens[first+1]
		if mid == nil {
			return false
		}
		tok := mid.GetToken()
		if (tok == "als" || tok == "noch") && tokens[first+2] != nil &&
			tokens[first+2].MatchesPosTagRegex(`VER:.*INF.*`) {
			if first+2 == second {
				return true
			}
			return isTwoCombinedVerbsMissingComma(tokens[second-1], tokens[second])
		}
	}
	return false
}

func isPerfectMissingComma(tokens []*languagetool.AnalyzedTokenReadings, first, second int) bool {
	if first < 0 || second >= len(tokens) || tokens[first] == nil || tokens[second] == nil {
		return false
	}
	return tokens[first].HasPosTagStartingWith("VER:AUX:") &&
		tokens[second].MatchesPosTagRegex(`VER:.*(INF|PA2).*`)
}

func isSpecialInfMissingComma(tokens []*languagetool.AnalyzedTokenReadings, first, second, start int) bool {
	if first < 0 || first >= len(tokens) || tokens[first] == nil || !tokens[first].HasPosTagStartingWith("VER:INF") {
		return false
	}
	for i := first - 1; i > start; i-- {
		if tokens[i] != nil && tokens[i].HasPosTagStartingWith("ART") {
			j := skipSubMissingComma(tokens, i, second)
			return j > 0
		}
	}
	return false
}

func isPerfect3MissingComma(tokens []*languagetool.AnalyzedTokenReadings, first, second, third int) bool {
	if second < 0 || second >= len(tokens) || tokens[second] == nil {
		return false
	}
	return tokens[second].MatchesPosTagRegex(`VER:.*INF.*`) && isPerfectMissingComma(tokens, first, third)
}

func isSeparatorOrInfMissingComma(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	if n < 0 || n >= len(tokens) || tokens[n] == nil {
		return false
	}
	return isSeparatorMissingComma(tokens[n].GetToken()) ||
		tokens[n].HasPosTagStartingWith("VER:INF") ||
		(n+1 < len(tokens) && tokens[n].GetToken() == "zu" && tokens[n+1] != nil &&
			tokens[n+1].MatchesPosTagRegex(`VER:.*INF.*`))
}

// getCommaBehindMissingComma ports Java getCommaBehind.
func getCommaBehindMissingComma(tokens []*languagetool.AnalyzedTokenReadings, verbs []int, start, end int) int {
	if len(verbs) == 0 {
		return -1
	}
	if len(verbs) == 1 {
		v0 := verbs[0]
		if v0+1 < len(tokens) && tokens[v0+1] != nil && isSeparatorMissingComma(tokens[v0+1].GetToken()) {
			return -1
		}
		return v0
	}
	if len(verbs) == 2 {
		v0, v1 := verbs[0], verbs[1]
		if isSpecialPairMissingComma(tokens, v0, v1) {
			if isSeparatorOrInfMissingComma(tokens, v1+1) {
				return -1
			}
			return v1
		} else if v0+1 == v1 {
			if isTwoCombinedVerbsMissingComma(tokens[v0], tokens[v1]) {
				if isSeparatorOrInfMissingComma(tokens, v1+1) || isKonAfterVerbMissingComma(tokens, v1, end) {
					return -1
				}
				return v1
			}
		} else if v0+2 == v1 {
			if isThreeCombinedVerbsMissingComma(tokens, v0, v1) {
				if isSeparatorOrInfMissingComma(tokens, v1+1) {
					return -1
				}
				return v1
			}
		}
		if isParMissingComma(tokens[v0]) || isPerfectMissingComma(tokens, v0, v1) ||
			isInfinitivZuMissingComma(tokens, v1) || isSpecialInfMissingComma(tokens, v0, v1, start) {
			if isSeparatorOrInfMissingComma(tokens, v1+1) {
				return -1
			}
			return v1
		}
	}
	if len(verbs) == 3 {
		v0, v1, v2 := verbs[0], verbs[1], verbs[2]
		if isTwoPlusCombinedVerbsMissingComma(tokens, v0, v2) {
			if isSeparatorOrInfMissingComma(tokens, v2+1) {
				return -1
			}
			return v2
		} else if v0+2 == v2 {
			if v0+1 == v1 && isThreeCombinedVerbsMissingComma(tokens, v0, v2) {
				if isSeparatorOrInfMissingComma(tokens, v2+1) {
					return -1
				}
				return v2
			}
		} else if v0+3 == v2 && isFourCombinedVerbsMissingComma(tokens, v0, v2) {
			if isSeparatorOrInfMissingComma(tokens, v2+1) {
				return -1
			}
			return v2
		} else if tokens[v2] != nil && tokens[v2].HasPosTagStartingWith("VER:MOD:") &&
			isSpecialPairMissingComma(tokens, v0, v1) {
			if isSeparatorOrInfMissingComma(tokens, v2+1) {
				return -1
			}
			return v2
		}
		if isPerfect3MissingComma(tokens, v0, v1, v2) {
			if isSeparatorOrInfMissingComma(tokens, v2+1) {
				return -1
			}
			return v1
		}
	}
	return verbs[0]
}

// missedCommaBehind ports Java missedCommaBehind.
func missedCommaBehind(tokens []*languagetool.AnalyzedTokenReadings, inFront, start, end int) int {
	for i := start; i < end && i < len(tokens); i++ {
		if !isPronounMissingComma(tokens, i) {
			continue
		}
		verbs := verbPosMissingComma(tokens, i, end)
		if len(verbs) == 0 {
			continue
		}
		gender := getGenderMissingComma(tokens[i])
		if !isAnyVerbMissingComma(tokens, i+1) &&
			matchesGenderMissingComma(gender, tokens, inFront, i-1) &&
			!isArticleMissingComma(gender, tokens, i, verbs[len(verbs)-1]) {
			return getCommaBehindMissingComma(tokens, verbs, i, end)
		}
	}
	return -1
}

func getSinOrPluOfProMissingComma(token *languagetool.AnalyzedTokenReadings) string {
	if token == nil {
		return ""
	}
	if !token.HasPartialPosTag("PRO:PER:") && !token.HasPosTagStartingWith("IND:") {
		return ""
	}
	var parts []string
	if token.MatchesPosTagRegex(`.*:SIN.*`) {
		parts = append(parts, "SIN")
	}
	if token.MatchesPosTagRegex(`.*:PLU.*`) {
		parts = append(parts, "PLU")
	}
	if len(parts) > 1 {
		return "(" + strings.Join(parts, "|") + ")"
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return ""
}

func isVerbProPairMissingComma(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	if n+1 >= len(tokens) || tokens[n] == nil || tokens[n+1] == nil {
		return false
	}
	sinOrPlu := getSinOrPluOfProMissingComma(tokens[n+1])
	if sinOrPlu == "" {
		return false
	}
	return tokens[n].MatchesPosTagRegex(`VER:.*` + sinOrPlu + `.*`)
}

func (r *MissingCommaRelativeClauseRule) matchMorphFront(tokens []*languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	subStart := 1
	if subStart < len(tokens) && tokens[subStart] != nil && isSeparatorMissingComma(tokens[subStart].GetToken()) {
		subStart++
	}
	for subStart < len(tokens) {
		subEnd := nextSeparatorMissingComma(tokens, subStart)
		lastVerb := hasPotentialSubclause(tokens, subStart, subEnd)
		if lastVerb > 0 {
			nToken := missedCommaInFront(tokens, subStart, subEnd, lastVerb)
			if nToken > 0 {
				startToken := nToken - 1
				if isPrpMissingComma(tokens[nToken-1]) {
					startToken = nToken - 2
				}
				if startToken < 0 {
					startToken = nToken - 1
				}
				// Java: RuleMatch without shortMessage.
				msg := "Sowohl angehängte als auch eingeschobene Relativsätze werden durch Kommas vom Hauptsatz getrennt."
				rm := rules.NewRuleMatch(r, sentence, tokens[startToken].GetStartPos(), tokens[nToken].GetEndPos(), msg)
				if nToken-startToken > 1 {
					rm.SetSuggestedReplacement(tokens[startToken].GetToken() + ", " + tokens[nToken-1].GetToken() + " " + tokens[nToken].GetToken())
				} else {
					rm.SetSuggestedReplacement(tokens[startToken].GetToken() + ", " + tokens[nToken].GetToken())
				}
				matches = append(matches, rm)
			}
		}
		subStart = subEnd + 1
	}
	return matches
}

// matchMorphBehind ports Java match() when behind==true.
func (r *MissingCommaRelativeClauseRule) matchMorphBehind(tokens []*languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	subStart := 1
	if subStart < len(tokens) && tokens[subStart] != nil && isSeparatorMissingComma(tokens[subStart].GetToken()) {
		subStart++
	}
	subInFront := subStart
	subStart = nextSeparatorMissingComma(tokens, subInFront) + 1
	for subStart < len(tokens) {
		subEnd := nextSeparatorMissingComma(tokens, subStart)
		lastVerb := hasPotentialSubclause(tokens, subStart, subEnd)
		if lastVerb > 0 {
			nToken := missedCommaBehind(tokens, subInFront, subStart, subEnd)
			if nToken > 0 {
				if isVerbProPairMissingComma(tokens, nToken) {
					// Java: RuleMatch without shortMessage.
					msg := "Sollten Sie hier ein Komma einfügen oder zwei?"
					rm := rules.NewRuleMatch(r, sentence, tokens[nToken-1].GetStartPos(), tokens[nToken+1].GetEndPos(), msg)
					rm.SetSuggestedReplacements([]string{
						tokens[nToken-1].GetToken() + ", " + tokens[nToken].GetToken() + " " + tokens[nToken+1].GetToken() + ",",
						tokens[nToken-1].GetToken() + " " + tokens[nToken].GetToken() + " " + tokens[nToken+1].GetToken() + ",",
						tokens[nToken-1].GetToken() + " " + tokens[nToken].GetToken() + ", " + tokens[nToken+1].GetToken(),
					})
					matches = append(matches, rm)
				} else if nToken+1 < len(tokens) && tokens[nToken] != nil && tokens[nToken+1] != nil {
					msg := "Sollten Sie hier ein Komma einfügen?"
					rm := rules.NewRuleMatch(r, sentence, tokens[nToken].GetStartPos(), tokens[nToken+1].GetEndPos(), msg)
					rm.SetSuggestedReplacement(tokens[nToken].GetToken() + ", " + tokens[nToken+1].GetToken())
					matches = append(matches, rm)
				}
			}
		}
		subInFront = subStart
		subStart = subEnd + 1
	}
	return matches
}
