package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PossessiusRedundantsFilter ports
// org.languagetool.rules.ca.PossessiusRedundantsFilter (1:1 AcceptRuleMatch).
type PossessiusRedundantsFilter struct{}

func NewPossessiusRedundantsFilter() *PossessiusRedundantsFilter {
	return &PossessiusRedundantsFilter{}
}

// PossessiveSuggestionInput holds pre-analyzed span info (unit-test helper surface).
type PossessiveSuggestionInput struct {
	Persona, Number string // from PX postag (chars at 2 and 6)
	// PronounFound: matching dative/reflexive already present
	PronounFound, HasSomePronoun bool
	ApostropheNeeded             bool
	// Verb token at start of verb group
	VerbToken         string
	VerbIsInfOrGerund bool // VMN / VMG
	// Tokens between verb and possessive (exclusive of det before possessive)
	MiddleTokens []string
	// Det before possessive and following noun token
	DetToken, NounToken string
	// Full tokens for non-apostrophe rebuild (det + noun around possessive skip)
	AroundPossessive []string // det and following word(s) excluding possessive
	CasingModel      string
}

// Suggest builds the replacement string; empty means suppress (legacy unit helper).
func (f *PossessiusRedundantsFilter) Suggest(in PossessiveSuggestionInput) string {
	if in.PronounFound {
		if in.ApostropheNeeded {
			return "l'" + in.NounToken
		}
		return "" // delete possessive only — empty suggestion in Java
	}
	if !in.HasSomePronoun {
		var b strings.Builder
		if in.VerbIsInfOrGerund {
			pronounSugg := TransformDarrere(GetDativePronoun(in.Persona+in.Number), in.VerbToken)
			b.WriteString(in.VerbToken)
			b.WriteString(pronounSugg)
		} else {
			pronounSugg := TransformDavant(GetDativePronoun(in.Persona+in.Number), in.VerbToken)
			b.WriteString(tools.PreserveCase(pronounSugg, in.VerbToken))
			b.WriteString(strings.ToLower(in.VerbToken))
		}
		for _, tok := range in.MiddleTokens {
			b.WriteByte(' ')
			b.WriteString(strings.ToLower(tok))
		}
		if in.ApostropheNeeded {
			b.WriteString(" ")
			b.WriteString("l'" + in.NounToken)
		} else {
			for _, tok := range in.AroundPossessive {
				b.WriteByte(' ')
				b.WriteString(tok)
			}
		}
		s := b.String()
		if in.CasingModel != "" {
			s = tools.PreserveCase(s, in.CasingModel)
		}
		return strings.TrimSpace(s)
	}
	return "" // no suggestion
}

// PersonaNumberFromPX extracts persona+number from a PX… postag (Java substrings).
func PersonaNumberFromPX(postag string) (persona, number string) {
	if len(postag) < 7 {
		return "", ""
	}
	return string(postag[2]), string(postag[6])
}

// AcceptRuleMatch ports PossessiusRedundantsFilter.acceptRuleMatch.
func (f *PossessiusRedundantsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = f
	_ = arguments
	_ = patternTokens
	_ = tokenPositions
	if match == nil || match.Sentence == nil {
		return nil
	}
	tokens := match.Sentence.GetTokensWithoutWhitespace()
	if len(tokens) == 0 {
		return nil
	}

	posPossessive := patternTokenPos
	for posPossessive < len(tokens) && !tokens[posPossessive].HasPartialPosTag("PX") {
		posPossessive++
	}
	if posPossessive >= len(tokens) {
		return nil
	}
	pxReading := readingWithTagRegex(tokens[posPossessive], `PX.*`)
	if pxReading == nil || pxReading.GetPOSTag() == nil {
		return nil
	}
	possessivePostag := *pxReading.GetPOSTag()
	if len(possessivePostag) < 7 {
		return nil
	}
	number := possessivePostag[6:7]
	persona := possessivePostag[2:3]

	posVerb := patternTokenPos - 1
	for posVerb > 0 && hasChunkTagGV(tokens[posVerb]) {
		posVerb--
	}
	posVerb++
	if posVerb < 0 || posVerb >= len(tokens) {
		return nil
	}

	pronounFound := false
	hasSomePronoun := false
	// pronom enrere
	posPronoun := posVerb - 1
	for !pronounFound && posPronoun > 0 &&
		(tokens[posPronoun].HasPosTagStartingWith("PP") || tokens[posPronoun].HasPosTagStartingWith("P0")) {
		hasSomePronoun = true
		pr := readingWithTagRegex(tokens[posPronoun], `P.*`)
		if pr != nil && pr.GetPOSTag() != nil {
			pronounPostag := *pr.GetPOSTag()
			if len(pronounPostag) >= 5 {
				pronounFound = pronounPostag[2:3] == persona &&
					(number == "C" || pronounPostag[4:5] == number)
			}
		}
		posPronoun--
	}
	// pronom avant
	posPronoun = patternTokenPos + 1
	for !pronounFound && posPronoun < len(tokens) &&
		(tokens[posPronoun].HasPosTagStartingWith("PP") || tokens[posPronoun].HasPosTagStartingWith("P0")) {
		hasSomePronoun = true
		pr := readingWithTagRegex(tokens[posPronoun], `P.*`)
		if pr != nil && pr.GetPOSTag() != nil {
			pronounPostag := *pr.GetPOSTag()
			if len(pronounPostag) >= 5 {
				pronounFound = pronounPostag[2:3] == persona &&
					(number == "C" || pronounPostag[4:5] == number)
			}
		}
		posPronoun++
	}

	// Cal apostrofar
	if posPossessive-1 < 0 || posPossessive+1 >= len(tokens) {
		return nil
	}
	apostropheNeeded := hasAnyPartialPosTag(tokens[posPossessive-1], "DA0MS0", "DA0FS0") &&
		pApostropheNeeded.MatchString(tokens[posPossessive+1].GetToken())

	if pronounFound {
		if apostropheNeeded {
			match.SetOffsetPosition(tokens[posPossessive-1].GetStartPos(), tokens[posPossessive+1].GetEndPos())
			match.SetSuggestedReplacement("l'" + tokens[posPossessive+1].GetToken())
		} else {
			match.SetOffsetPosition(tokens[posPossessive].GetStartPos(), tokens[posPossessive].GetEndPos())
			match.SetSuggestedReplacement("")
		}
		return match
	}
	if !hasSomePronoun {
		var suggestion strings.Builder
		if hasAnyPartialPosTag(tokens[posVerb], "VMN", "VMG") {
			pronounSugg := TransformDarrere(GetDativePronoun(persona+number), tokens[posVerb].GetToken())
			suggestion.WriteString(tokens[posVerb].GetToken())
			suggestion.WriteString(pronounSugg)
		} else {
			pronounSugg := TransformDavant(GetDativePronoun(persona+number), tokens[posVerb].GetToken())
			suggestion.WriteString(tools.PreserveCase(pronounSugg, tokens[posVerb].GetToken()))
			suggestion.WriteString(strings.ToLower(tokens[posVerb].GetToken()))
		}
		for i := posVerb + 1; i <= posPossessive-2; i++ {
			if i < 0 || i >= len(tokens) {
				continue
			}
			if tokens[i].IsWhitespaceBefore() {
				suggestion.WriteByte(' ')
			}
			suggestion.WriteString(strings.ToLower(tokens[i].GetToken()))
		}
		if apostropheNeeded {
			suggestion.WriteString(" ")
			suggestion.WriteString("l'" + tokens[posPossessive+1].GetToken())
		} else {
			for i := posPossessive - 1; i <= posPossessive+1; i++ {
				if i == posPossessive {
					continue
				}
				if i < 0 || i >= len(tokens) {
					continue
				}
				if tokens[i].IsWhitespaceBefore() {
					suggestion.WriteByte(' ')
				}
				suggestion.WriteString(tokens[i].GetToken())
			}
		}
		match.SetOffsetPosition(tokens[posVerb].GetStartPos(), tokens[posPossessive+1].GetEndPos())
		match.SetSuggestedReplacement(suggestion.String())
		return match
	}
	return nil
}

func hasChunkTagGV(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, c := range tok.GetChunkTags() {
		if c == "GV" {
			return true
		}
	}
	return false
}

// hasAnyPartialPosTag ports AnalyzedTokenReadings.hasAnyPartialPosTag.
func hasAnyPartialPosTag(tok *languagetool.AnalyzedTokenReadings, posTags ...string) bool {
	if tok == nil {
		return false
	}
	for _, t := range posTags {
		if tok.HasPartialPosTag(t) {
			return true
		}
	}
	return false
}
