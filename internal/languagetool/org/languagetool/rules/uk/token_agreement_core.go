package uk

import (
	"regexp"
	"strings"

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

// IsAdjNounException ports TokenAgreementAdjNounExceptionHelper surface.
func IsAdjNounException(tokens []*languagetool.AnalyzedTokenReadings, adjPos, nounPos int) bool {
	if adjPos < 0 || nounPos < 0 || adjPos >= len(tokens) || nounPos >= len(tokens) {
		return true
	}
	// skip if same token
	if adjPos == nounPos {
		return true
	}
	// fake feminine nouns often mismatch gender with fem adj — treat as exception list presence on noun lemma
	if tokens[nounPos] != nil {
		w := tokens[nounPos].GetToken()
		for _, f := range FakeFemList {
			if w == f {
				return true
			}
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
