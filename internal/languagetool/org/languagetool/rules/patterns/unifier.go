package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Unifier ports org.languagetool.rules.patterns.Unifier.
type Unifier struct {
	equivalenceTypes    map[EquivalenceTypeLocator]*PatternToken
	equivalenceFeatures map[string][]string

	committed []*uniPosition
	current   *uniPosition
	// agreed: set of combination keys (types joined by \x00 in featureOrder)
	agreed map[string]struct{}
	// agreedLists preserves combination as []string for support checks
	agreedLists  map[string][]string
	featureOrder []string
	allFeatsIn   bool

	inUnification bool
	uniMatched    bool
	uniAllMatched bool
}

type uniPosition struct {
	neutral         bool
	tokens          []*languagetool.AnalyzedToken
	matched         []map[string]map[string]struct{} // per reading: feature → set of types
	neutralReadings *languagetool.AnalyzedTokenReadings
}

func NewUnifier(
	equivalenceTypes map[EquivalenceTypeLocator]*PatternToken,
	equivalenceFeatures map[string][]string,
) *Unifier {
	return &Unifier{
		equivalenceTypes:    equivalenceTypes,
		equivalenceFeatures: equivalenceFeatures,
	}
}

// IsSatisfied ports Unifier.isSatisfied.
func (u *Unifier) IsSatisfied(aToken *languagetool.AnalyzedToken, uFeatures map[string][]string) bool {
	if uFeatures == nil {
		panic("uFeatures must not be null")
	}
	u.setFeatures(uFeatures)

	if u.allFeatsIn && (u.agreed == nil || len(u.agreed) == 0) {
		return false
	}

	matched := u.matchedTypes(aToken, uFeatures)
	allFeaturesMatched := matched != nil && noEmptyFeature(matched)

	if !u.allFeatsIn {
		if allFeaturesMatched {
			u.openCurrent(false, nil).add(aToken, matched)
		}
		return allFeaturesMatched
	}

	token := u.openCurrent(false, nil)
	compatible := allFeaturesMatched && supportsAny(matched, u.agreedLists, u.featureOrder)
	if compatible {
		token.add(aToken, matched)
	}
	return compatible
}

// StartUnify ports Unifier.startUnify.
func (u *Unifier) StartUnify() {
	u.commitCurrent()
	u.allFeatsIn = true
}

// StartNextToken ports Unifier.startNextToken.
func (u *Unifier) StartNextToken() {
	u.commitCurrent()
}

// AddNeutralElement ports Unifier.addNeutralElement.
func (u *Unifier) AddNeutralElement(readings *languagetool.AnalyzedTokenReadings) {
	u.commitCurrent()
	u.committed = append(u.committed, &uniPosition{neutral: true, neutralReadings: readings})
}

// GetFinalUnificationValue ports Unifier.getFinalUnificationValue.
func (u *Unifier) GetFinalUnificationValue(uFeatures map[string][]string) bool {
	u.setFeatures(uFeatures)
	shared := u.finalAgreed()
	return len(shared) > 0 && u.hasNonNeutral()
}

// GetUnifiedTokens ports Unifier.getUnifiedTokens.
func (u *Unifier) GetUnifiedTokens() []*languagetool.AnalyzedTokenReadings {
	sequence := u.orderedPositions()
	if len(sequence) == 0 {
		return nil
	}
	shared := u.finalAgreed()
	var result []*languagetool.AnalyzedTokenReadings
	for _, position := range sequence {
		atr := u.unify(position, shared)
		if atr == nil {
			return nil
		}
		result = append(result, atr)
	}
	return result
}

// IsUnified ports Unifier.isUnified with isMatched default true.
func (u *Unifier) IsUnified(matchToken *languagetool.AnalyzedToken, uFeatures map[string][]string, lastReading bool) bool {
	return u.IsUnifiedMatched(matchToken, uFeatures, lastReading, true)
}

// IsUnifiedMatched ports Unifier.isUnified(..., isMatched).
func (u *Unifier) IsUnifiedMatched(matchToken *languagetool.AnalyzedToken, uFeatures map[string][]string, lastReading, isMatched bool) bool {
	if u.inUnification {
		if isMatched {
			u.uniMatched = u.uniMatched || u.IsSatisfied(matchToken, uFeatures)
		}
		u.uniAllMatched = u.uniMatched
		if lastReading {
			u.StartNextToken()
			u.uniMatched = false
		}
		return u.uniAllMatched && u.GetFinalUnificationValue(uFeatures)
	}
	if isMatched {
		u.IsSatisfied(matchToken, uFeatures)
	}
	if lastReading {
		u.inUnification = true
		u.uniMatched = false
		u.StartUnify()
	}
	return true
}

// GetFinalUnified ports Unifier.getFinalUnified.
func (u *Unifier) GetFinalUnified() []*languagetool.AnalyzedTokenReadings {
	if u.inUnification {
		return u.GetUnifiedTokens()
	}
	return nil
}

// Reset ports Unifier.reset.
func (u *Unifier) Reset() {
	u.committed = nil
	u.current = nil
	u.agreed = nil
	u.agreedLists = nil
	u.featureOrder = nil
	u.allFeatsIn = false
	u.inUnification = false
	u.uniMatched = false
	u.uniAllMatched = false
}

func (u *Unifier) setFeatures(uFeatures map[string][]string) {
	if u.featureOrder == nil {
		u.featureOrder = make([]string, 0, len(uFeatures))
		for k := range uFeatures {
			u.featureOrder = append(u.featureOrder, k)
		}
	}
}

func (u *Unifier) matchedTypes(aToken *languagetool.AnalyzedToken, uFeatures map[string][]string) map[string]map[string]struct{} {
	matched := map[string]map[string]struct{}{}
	for feat, types := range uFeatures {
		if len(types) == 0 {
			types = u.equivalenceFeatures[feat]
		}
		matchedForFeat := map[string]struct{}{}
		for _, typeName := range types {
			testElem := u.equivalenceTypes[NewEquivalenceTypeLocator(feat, typeName)]
			if testElem == nil {
				return nil
			}
			if testElem.IsMatched(aToken) {
				matchedForFeat[typeName] = struct{}{}
			}
		}
		matched[feat] = matchedForFeat
	}
	return matched
}

func noEmptyFeature(matched map[string]map[string]struct{}) bool {
	for _, types := range matched {
		if len(types) == 0 {
			return false
		}
	}
	return true
}

func (p *uniPosition) add(token *languagetool.AnalyzedToken, matched map[string]map[string]struct{}) {
	p.tokens = append(p.tokens, token)
	p.matched = append(p.matched, matched)
}

func (p *uniPosition) isEmpty() bool { return len(p.tokens) == 0 }

func (u *Unifier) openCurrent(neutral bool, neutralReadings *languagetool.AnalyzedTokenReadings) *uniPosition {
	if u.current == nil {
		u.current = &uniPosition{neutral: neutral, neutralReadings: neutralReadings}
	}
	return u.current
}

func (u *Unifier) commitCurrent() {
	if u.current == nil {
		return
	}
	u.intersect(u.current)
	u.committed = append(u.committed, u.current)
	u.current = nil
}

func (u *Unifier) intersect(position *uniPosition) {
	if position.neutral {
		return
	}
	combs := u.combinationsOf(position)
	if u.agreed == nil {
		u.agreed = combs
		u.agreedLists = map[string][]string{}
		for k := range combs {
			u.agreedLists[k] = strings.Split(k, "\x00")
		}
		return
	}
	for k := range u.agreed {
		if !u.supportsPosition(position, u.agreedLists[k]) {
			delete(u.agreed, k)
			delete(u.agreedLists, k)
		}
	}
}

func (u *Unifier) finalAgreed() map[string][]string {
	if u.current == nil || u.current.neutral {
		if u.agreedLists == nil {
			return map[string][]string{}
		}
		out := make(map[string][]string, len(u.agreedLists))
		for k, v := range u.agreedLists {
			out[k] = v
		}
		return out
	}
	if u.agreed == nil {
		combs := u.combinationsOf(u.current)
		out := map[string][]string{}
		for k := range combs {
			out[k] = strings.Split(k, "\x00")
		}
		return out
	}
	out := map[string][]string{}
	for k, v := range u.agreedLists {
		if u.supportsPosition(u.current, v) {
			out[k] = v
		}
	}
	return out
}

func (u *Unifier) hasNonNeutral() bool {
	if u.current != nil && !u.current.neutral && !u.current.isEmpty() {
		return true
	}
	for _, position := range u.committed {
		if !position.neutral {
			return true
		}
	}
	return false
}

func (u *Unifier) orderedPositions() []*uniPosition {
	if u.current == nil {
		return u.committed
	}
	return append(append([]*uniPosition{}, u.committed...), u.current)
}

func (u *Unifier) combinationsOf(position *uniPosition) map[string]struct{} {
	combinations := map[string]struct{}{}
	for _, matched := range position.matched {
		u.addCombinations(matched, 0, nil, combinations)
	}
	return combinations
}

func (u *Unifier) addCombinations(matched map[string]map[string]struct{}, featureIdx int, prefix []string, out map[string]struct{}) {
	if featureIdx == len(u.featureOrder) {
		out[strings.Join(prefix, "\x00")] = struct{}{}
		return
	}
	feat := u.featureOrder[featureIdx]
	for typ := range matched[feat] {
		u.addCombinations(matched, featureIdx+1, append(prefix, typ), out)
	}
}

func (u *Unifier) supportsPosition(position *uniPosition, combination []string) bool {
	for _, matched := range position.matched {
		if supportsMatched(matched, combination, u.featureOrder) {
			return true
		}
	}
	return false
}

func supportsMatched(matched map[string]map[string]struct{}, combination, featureOrder []string) bool {
	for i, feat := range featureOrder {
		if i >= len(combination) {
			return false
		}
		if _, ok := matched[feat][combination[i]]; !ok {
			return false
		}
	}
	return true
}

func supportsAny(matched map[string]map[string]struct{}, combinations map[string][]string, featureOrder []string) bool {
	if combinations == nil {
		return false
	}
	for _, combination := range combinations {
		if supportsMatched(matched, combination, featureOrder) {
			return true
		}
	}
	return false
}

func (u *Unifier) unify(position *uniPosition, shared map[string][]string) *languagetool.AnalyzedTokenReadings {
	if position.neutral {
		neutral := position.neutralReadings
		if neutral == nil {
			return nil
		}
		var atr *languagetool.AnalyzedTokenReadings
		for i := 0; i < neutral.GetReadingsLength(); i++ {
			atr = appendReading(atr, neutral.GetAnalyzedToken(i))
		}
		return atr
	}
	var atr *languagetool.AnalyzedTokenReadings
	for i, tok := range position.tokens {
		if supportsAny(position.matched[i], shared, u.featureOrder) {
			atr = appendReading(atr, tok)
		}
	}
	return atr
}

func appendReading(atr *languagetool.AnalyzedTokenReadings, token *languagetool.AnalyzedToken) *languagetool.AnalyzedTokenReadings {
	if atr == nil {
		return languagetool.NewAnalyzedTokenReadings(token)
	}
	atr.AddReading(token, "")
	return atr
}
