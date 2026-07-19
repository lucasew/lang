package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

// FakeFemList ports TokenAgreementAdjNounRule.FAKE_FEM_LIST.
var FakeFemList = []string{
	"ступінь", "степінь", "продаж", "собака", "дріб", "ярмарок",
	"нежить", "рукопис", "накип", "насип", "путь",
}

var (
	adjInflectionPattern  = regexp.MustCompile(`:([mfnp]):(v_...)(:r(in)?anim)?`)
	nounInflectionPattern = regexp.MustCompile(`((?:[iu]n)?anim):([mfnps]):(v_...)`)
	nounVZnaVarIgnore     = regexp.MustCompile(`v_zna:var`)
)

// CollectPOSTags gathers non-nil POS tags from an AnalyzedTokenReadings.
func CollectPOSTags(tok *languagetool.AnalyzedTokenReadings) []string {
	if tok == nil {
		return nil
	}
	var out []string
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		out = append(out, *r.GetPOSTag())
	}
	return out
}

// HasAdjReading reports whether any reading is adj*.
func HasAdjReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSAdj.Match(p) {
			return true
		}
	}
	return false
}

// HasNounReading reports whether any reading is noun*.
func HasNounReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSNoun.Match(p) {
			return true
		}
	}
	return false
}

// HasNounOrPronSubjectReading treats personal pronouns as subjects for noun–verb agreement.
func HasNounOrPronSubjectReading(tok *languagetool.AnalyzedTokenReadings) bool {
	if HasNounReading(tok) {
		return true
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, "pron:pers") {
			return true
		}
	}
	return false
}

// AdjNounAgree reports whether adj and noun POS tag sets share an inflection.
func AdjNounAgree(adjTags, nounTags []string) bool {
	master := GetAdjCaseInflections(adjTags)
	slave := GetNounInflectionsFromTags(nounTags, nounVZnaVarIgnore)
	if len(master) == 0 || len(slave) == 0 {
		return true // insufficient data — no flag
	}
	return InflectionsIntersect(master, slave)
}

// NumrNounAgree uses numr inflection pattern against nouns.
func NumrNounAgree(numrTags, nounTags []string) bool {
	master := GetNumrCaseInflections(numrTags)
	slave := GetNounCaseInflections(nounTags)
	if len(master) == 0 || len(slave) == 0 {
		return true
	}
	return InflectionsIntersect(master, slave)
}

// tokenAgreementMatch is shared match infrastructure.
// Java TokenAgreement* rules: setCategory(Categories.MISC).
type tokenAgreementMatch struct {
	ruleID      string
	description string
	shortMsg    string
	// Category ports Rule.category (Java MISC).
	category *rules.Category
	// pairChecker returns false when the pair disagrees
	pairChecker func(left, right *languagetool.AnalyzedTokenReadings) bool
	// isLeftToken identifies the "master" token class
	isLeftToken func(tok *languagetool.AnalyzedTokenReadings) bool
	// isRightToken identifies the "slave" token class
	isRightToken func(tok *languagetool.AnalyzedTokenReadings) bool
	// exception when true skips the flag
	exception func(tokens []*languagetool.AnalyzedTokenReadings, leftIdx, rightIdx int) bool
}

func (r *tokenAgreementMatch) GetID() string          { return r.ruleID }
func (r *tokenAgreementMatch) GetDescription() string { return r.description }
func (r *tokenAgreementMatch) GetShort() string       { return r.shortMsg }

// GetCategory ports Rule.getCategory (Java MISC).
func (r *tokenAgreementMatch) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.category
}

// initTokenAgreementMeta applies Java TokenAgreement* constructor metadata (MISC category).
func initTokenAgreementMeta(r *tokenAgreementMatch, messages map[string]string) {
	if r == nil {
		return
	}
	if r.category == nil {
		r.category = rules.CatMisc.GetCategory(messages)
	}
}

func (r *tokenAgreementMatch) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil || r.pairChecker == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	leftIdx := -1
	for i, tok := range tokens {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		if r.isLeftToken != nil && r.isLeftToken(tok) {
			leftIdx = i
			continue
		}
		if leftIdx < 0 {
			continue
		}
		if r.isRightToken != nil && !r.isRightToken(tok) {
			// skip ignorable intermediates (не, і, commas soft)
			if isIgnorableAgreementIntervening(tok) {
				continue
			}
			// non-matching intermediate — reset
			leftIdx = -1
			continue
		}
		if r.exception != nil && r.exception(tokens, leftIdx, i) {
			leftIdx = -1
			continue
		}
		if !r.pairChecker(tokens[leftIdx], tok) {
			msg := r.shortMsg
			if msg == "" {
				msg = r.description
			}
			m := rules.NewRuleMatch(r, sentence, tokens[leftIdx].GetStartPos(), tok.GetEndPos(), msg)
			out = append(out, m)
		}
		leftIdx = -1
	}
	return out
}

// isIgnorableAgreementIntervening allows particle/conj glue between master and slave.
func isIgnorableAgreementIntervening(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	// surface fast path
	switch strings.ToLower(tok.GetToken()) {
	case "не", "й", "і", "та", "чи", "то", "ж", "би", "б":
		return true
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.HasPrefix(p, "part") || strings.HasPrefix(p, "conj") {
			return true
		}
	}
	return false
}

// IsPredicativeAdjException soft-skips predicative adjectives.
func IsPredicativeAdjException(adj *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(adj) {
		if strings.Contains(p, "predic") || strings.HasPrefix(p, "predic") {
			return true
		}
	}
	return false
}

// IsAdjpException soft-skips pure participle adjp without case agreement expectation.
func IsAdjpException(adj *languagetool.AnalyzedTokenReadings) bool {
	tags := CollectPOSTags(adj)
	if len(tags) == 0 {
		return false
	}
	hasAdjp, hasCaseAdj := false, false
	for _, p := range tags {
		if strings.Contains(p, "adjp") {
			hasAdjp = true
		}
		if strings.HasPrefix(p, "adj") && strings.Contains(p, "v_") {
			hasCaseAdj = true
		}
	}
	return hasAdjp && !hasCaseAdj
}

// --- Exception helper stubs (full tables deferred) ---

// IsAdjNounException ports TokenAgreementAdjNounExceptionHelper early arms
// (full 1300-line table still deferred). FAKE_FEM uses Java lemma+partPos.
func IsAdjNounException(tokens []*languagetool.AnalyzedTokenReadings, adjPos, nounPos int) bool {
	if adjPos < 0 || nounPos < 0 || adjPos >= len(tokens) || nounPos >= len(tokens) {
		return true
	}
	// skip if same token
	if adjPos == nounPos {
		return true
	}
	// Java: LemmaHelper.hasLemma(noun, FAKE_FEM_LIST, "noun:inanim:m:")
	if HasLemmaWithPartPos(tokens[nounPos], FakeFemList, "noun:inanim:m:") {
		return true
	}

	adj := tokens[adjPos]
	noun := tokens[nounPos]
	if adj == nil || noun == nil {
		return true
	}

	// схований всередині номера: intervening adv with case government matches noun case
	if nounPos-adjPos > 1 {
		mid := tokens[adjPos+1]
		cases := LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadings(mid, "adv")
		if len(cases) > 0 {
			var list []string
			for c := range cases {
				list = append(list, c)
			}
			if HasVidmPosTag(list, noun) {
				return true
			}
		}
	}

	// Великий + Вітчизняний/Житомирський (capitalized), not війна
	if adjPos > 1 {
		prev := tokens[adjPos-1]
		if prev != nil &&
			IsCapitalized(adj.GetCleanToken()) && IsCapitalized(prev.GetCleanToken()) &&
			(HasLemmaToken(adj, "вітчизняний") || HasLemmaToken(adj, "житомирський")) &&
			HasLemmaToken(prev, "великий") &&
			!HasLemmaToken(noun, "війна") {
			return true
		}
		// Перший Національний (both uppercased first char)
		if HasLemmaToken(adj, "національний") && HasLemmaToken(prev, "перший") {
			at, pt := adj.GetToken(), prev.GetToken()
			if at != "" && pt != "" && isUpperFirst(at) && isUpperFirst(pt) {
				return true
			}
		}
		// (ні)чого доброго
		if CleanTokenLower(adj) == "доброго" {
			if regexp.MustCompile(`^(ні)?чого$`).MatchString(CleanTokenLower(prev)) {
				return true
			}
		}
		// у/в середньому|цілому|основному|подальшому
		if regexp.MustCompile(`(?iu)^(середньому|цілому|основному|подальшому)$`).MatchString(CleanTokenLower(adj)) &&
			regexp.MustCompile(`(?iu)^[ву]$`).MatchString(CleanTokenLower(prev)) {
			return true
		}
		// лава запасних
		if adj.GetToken() == "запасних" && HasLemmaToken(prev, "лава") {
			return true
		}
		// статтю 6-ту / num after стаття
		if HasPosTagPart(adj, "num") && HasLemmaToken(prev, "стаття") {
			return true
		}
	}

	// голому сорочка
	if strings.EqualFold(CleanTokenLower(adj), "голому") && strings.EqualFold(CleanTokenLower(noun), "сорочка") {
		return true
	}
	// бережений бог
	if HasLemmaWithPosRE(adj, []string{"бережений"}, regexp.MustCompile(`^adj:m:v_rod.*$`)) &&
		HasLemmaWithPosRE(noun, []string{"бог"}, regexp.MustCompile(`^noun:anim:m:v_naz.*$`)) {
		return true
	}
	// кожний + mass/quantity noun in instrumental
	if HasLemmaWithPosRE(adj, []string{"кожний"}, regexp.MustCompile(`^adj:f:v_naz.*$`)) &&
		HasLemmaWithPosRE(noun,
			[]string{"вага", "маса", "вартість", "потужність", "тривалість", "чисельність", "номінал", "наклад"},
			regexp.MustCompile(`^noun:inanim:.:v_oru.*$`)) {
		return true
	}
	// Божий / Господній / Христовий capitalized
	if HasLemmaTokenAny(adj, []string{"божий", "господній", "Христовий"}) && isUpperFirst(adj.GetToken()) {
		return true
	}
	// 5-а клас
	if regexp.MustCompile(`^([1-9]|1[0-2])[\x{2018}-][а-д]$`).MatchString(adj.GetToken()) && HasLemmaToken(noun, "клас") {
		return true
	}
	// перший + not FAKE_FEM inanim:m
	if nounPos > 1 && HasLemmaTokenAny(adj, []string{"перший"}) &&
		!HasLemmaWithPartPos(noun, FakeFemList, "noun:inanim:m:") {
		return true
	}
	// старший зміни/групи
	if (CleanTokenLower(noun) == "зміни" || CleanTokenLower(noun) == "групи") && HasLemmaToken(adj, "старший") {
		return true
	}

	return false
}

func isUpperFirst(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsUpper(r)
}

// HasPosTagPart reports whether any POS contains substr (Java PosTagHelper.hasPosTagPart).
func HasPosTagPart(tok *languagetool.AnalyzedTokenReadings, substr string) bool {
	if tok == nil || substr == "" {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, substr) {
			return true
		}
	}
	return false
}

// IsPrepNounException stub.
func IsPrepNounException(tokens []*languagetool.AnalyzedTokenReadings, prepPos, nounPos int) bool {
	return prepPos < 0 || nounPos <= prepPos
}

// IsNumrNounException stub.
func IsNumrNounException(tokens []*languagetool.AnalyzedTokenReadings, numrPos, nounPos int) bool {
	return numrPos < 0 || nounPos <= numrPos
}

// IsNounVerbException stub.
func IsNounVerbException(tokens []*languagetool.AnalyzedTokenReadings, nounPos, verbPos int) bool {
	return nounPos < 0 || verbPos <= nounPos
}

// IsVerbNounException stub.
func IsVerbNounException(tokens []*languagetool.AnalyzedTokenReadings, verbPos, nounPos int) bool {
	return verbPos < 0 || nounPos <= verbPos
}
