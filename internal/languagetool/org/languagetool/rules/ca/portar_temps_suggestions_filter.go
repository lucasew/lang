package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PortarTempsSuggestionsFilter ports
// org.languagetool.rules.ca.PortarTempsSuggestionsFilter (1:1 AcceptRuleMatch).
//
// Synthesize ports Synthesizer.synthesize(token, postag, postagRegExp).
// When nil, Accept falls back to SynthFer / SynthInfinitiveToFinite / SynthEstar.
type PortarTempsSuggestionsFilter struct {
	// Synthesize ports getSynthesizerFromRuleMatch(...).synthesize(token, postag, postagRegExp).
	Synthesize func(tok *languagetool.AnalyzedToken, postag string, postagRegExp bool) []string
	// Legacy unit-test hooks (Suggest + Accept fallback).
	SynthFer                func(postagPattern string) string
	SynthInfinitiveToFinite func(lemma, finitePostag string) string
	SynthEstar              func(finitePostag string) string
}

func NewPortarTempsSuggestionsFilter() *PortarTempsSuggestionsFilter {
	return &PortarTempsSuggestionsFilter{}
}

// PortarTempsKind classifies the token after the time span (legacy Suggest).
type PortarTempsKind int

const (
	PortarTempsQue PortarTempsKind = iota
	PortarTempsGerund
	PortarTempsSenseInf
	PortarTempsEstarPred
)

// PortarTempsInput is the surface input for Suggest.
type PortarTempsInput struct {
	PortarPostag             string
	TimeTokens               []string
	Kind                     PortarTempsKind
	NextLemma, PronounsAfter string
	CasingModel              string
}

// Suggest builds "fa una hora que …" replacements (unit helper).
func (f *PortarTempsSuggestionsFilter) Suggest(in PortarTempsInput) string {
	if f == nil || len(in.PortarPostag) < 8 {
		return ""
	}
	pattern := in.PortarPostag[:4] + "[30][S0]." + string(in.PortarPostag[7])
	fer := f.synthFirst("fer", pattern, true)
	if fer == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(fer)
	for _, t := range in.TimeTokens {
		b.WriteByte(' ')
		b.WriteString(t)
	}
	switch in.Kind {
	case PortarTempsQue:
		b.WriteString(" que")
	case PortarTempsGerund:
		finiteTag := "V.I" + in.PortarPostag[3:8]
		fin := f.synthFirst(in.NextLemma, finiteTag, true)
		if fin == "" {
			return ""
		}
		b.WriteString(" que ")
		if in.PronounsAfter != "" {
			b.WriteString(TransformDavant(in.PronounsAfter, fin))
		}
		b.WriteString(fin)
	case PortarTempsSenseInf:
		finiteTag := "V.I" + in.PortarPostag[3:8]
		fin := f.synthFirst(in.NextLemma, finiteTag, false)
		if fin == "" {
			return ""
		}
		b.WriteString(" que no ")
		if in.PronounsAfter != "" {
			b.WriteString(TransformDavant(in.PronounsAfter, fin))
		}
		b.WriteString(fin)
	case PortarTempsEstarPred:
		finiteTag := "V.I" + in.PortarPostag[3:8]
		estar := f.synthFirst("estar", finiteTag, false)
		if estar == "" {
			return ""
		}
		b.WriteString(" que ")
		b.WriteString(estar)
	default:
		return ""
	}
	s := b.String()
	if in.CasingModel != "" {
		s = tools.PreserveCase(s, in.CasingModel)
	}
	return s
}

func (f *PortarTempsSuggestionsFilter) synthFirst(lemma, postag string, postagRegExp bool) string {
	if f.Synthesize != nil {
		at := languagetool.NewAnalyzedToken("", nil, &lemma)
		forms := f.Synthesize(at, postag, postagRegExp)
		if len(forms) > 0 {
			return forms[0]
		}
		return ""
	}
	// Legacy hooks
	if lemma == "fer" && f.SynthFer != nil {
		return f.SynthFer(postag)
	}
	if lemma == "estar" && f.SynthEstar != nil {
		return f.SynthEstar(postag)
	}
	if f.SynthInfinitiveToFinite != nil && lemma != "fer" && lemma != "estar" {
		return f.SynthInfinitiveToFinite(lemma, postag)
	}
	return ""
}

// AcceptRuleMatch ports PortarTempsSuggestionsFilter.acceptRuleMatch.
func (f *PortarTempsSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = arguments
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	if f.Synthesize == nil && f.SynthFer == nil {
		return nil
	}

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	if posWord >= len(tokens) {
		return nil
	}

	vr := readingWithTagRegex(tokens[posWord], `V.*`)
	if vr == nil || vr.GetPOSTag() == nil {
		return nil
	}
	verbPostag := *vr.GetPOSTag()
	if len(verbPostag) < 8 {
		return nil
	}

	// fer with regex postag: verbPostag[0:4]+"[30][S0]."+verbPostag[7:8]
	newPostag := verbPostag[:4] + "[30][S0]." + verbPostag[7:8]
	ferForm := f.synthFirst("fer", newPostag, true)
	if ferForm == "" {
		return nil
	}

	var suggestion strings.Builder
	suggestion.WriteString(ferForm)

	i := posWord + 1
	for i < len(tokens) && hasChunkTag(tokens[i], "PTime") {
		if tokens[i].IsWhitespaceBefore() {
			suggestion.WriteByte(' ')
		}
		suggestion.WriteString(tokens[i].GetToken())
		i++
	}
	lastTokenPos := i
	if lastTokenPos+1 >= len(tokens) {
		return nil
	}
	adjustEndPos := 0
	lastToken := tokens[lastTokenPos]

	switch {
	case lastToken.GetToken() == "que":
		suggestion.WriteString(" que")

	case lastToken.HasPosTagStartingWith("VMG") || lastToken.HasPosTagStartingWith("VSG"):
		suggestion.WriteString(" que ")
		pronoms, nAfter := pronounsStrAfter(tokens, lastTokenPos)
		adjustEndPos += nAfter
		gr := readingWithTagRegex(lastToken, `V.G.*`)
		if gr == nil || gr.GetLemma() == nil {
			return nil
		}
		lemma := *gr.GetLemma()
		finiteTag := "V.I" + verbPostag[3:8]
		fin := f.synthFirst(lemma, finiteTag, true)
		if fin == "" {
			return nil
		}
		if pronoms != "" {
			suggestion.WriteString(TransformDavant(pronoms, fin))
		}
		suggestion.WriteString(fin)

	case lastToken.GetToken() == "sense" &&
		(tokens[lastTokenPos+1].HasPosTagStartingWith("VSN") || tokens[lastTokenPos+1].HasPosTagStartingWith("VMN")):
		suggestion.WriteString(" que no ")
		adjustEndPos++
		pronoms, nAfter := pronounsStrAfter(tokens, lastTokenPos+1)
		adjustEndPos += nAfter
		nr := readingWithTagRegex(tokens[lastTokenPos+1], `V.N.*`)
		if nr == nil || nr.GetLemma() == nil {
			return nil
		}
		lemma := *nr.GetLemma()
		finiteTag := "V.I" + verbPostag[3:8]
		// Java: synthesize without postagRegExp true
		fin := f.synthFirst(lemma, finiteTag, false)
		if fin == "" {
			return nil
		}
		if pronoms != "" {
			suggestion.WriteString(TransformDavant(pronoms, fin))
		}
		suggestion.WriteString(fin)

	case lastToken.GetToken() == "així" || lastToken.GetToken() == "a" || lastToken.GetToken() == "en" ||
		lastToken.GetToken() == "ací" || lastToken.GetToken() == "aquí" || lastToken.GetToken() == "ahí" ||
		lastToken.GetToken() == "allí" || lastToken.GetToken() == "allà" || lastToken.GetToken() == "de" ||
		lastToken.HasPosTagStartingWith("AQ") || lastToken.HasPosTagStartingWith("VMP"):
		finiteTag := "V.I" + verbPostag[3:8]
		estar := f.synthFirst("estar", finiteTag, false)
		if estar == "" {
			return nil
		}
		suggestion.WriteString(" que " + estar)
		adjustEndPos--

	default:
		return nil
	}

	replacement := tools.PreserveCase(suggestion.String(), tokens[posWord].GetToken())
	if replacement == "" {
		return nil
	}
	endIdx := lastTokenPos + adjustEndPos
	if endIdx < 0 || endIdx >= len(tokens) {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[posWord].GetStartPos(), tokens[endIdx].GetEndPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacement(replacement)
	return out
}

func hasChunkTag(tok *languagetool.AnalyzedTokenReadings, tag string) bool {
	if tok == nil {
		return false
	}
	for _, c := range tok.GetChunkTags() {
		if c == tag {
			return true
		}
	}
	return false
}

// pronounsStrAfter ports VerbSynthesizer pronouns-after scan from a verb index
// (clitics attached without whitespace, pPronomFeble).
func pronounsStrAfter(tokens []*languagetool.AnalyzedTokenReadings, iLastVerb int) (string, int) {
	if iLastVerb < 0 || iLastVerb >= len(tokens) {
		return "", 0
	}
	n := 0
	i := 1
	for iLastVerb+i < len(tokens) &&
		!tokens[iLastVerb+i].IsWhitespaceBefore() &&
		readingWithPPronomFeble(tokens[iLastVerb+i]) != nil {
		n++
		i++
	}
	if n == 0 {
		return "", 0
	}
	return joinTokensFromTo(tokens, iLastVerb+1, iLastVerb+n), n
}

func readingWithPPronomFeble(tok *languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedToken {
	// Java PronomsFeblesHelper.pPronomFeble (Matcher.matches = full string).
	return readingWithTagRegex(tok, `P0.{6}|PP3CN000|PP3NN000|PP3..A00|PP[123]CP000|PP3CSD00`)
}

func joinTokensFromTo(tokens []*languagetool.AnalyzedTokenReadings, start, end int) string {
	if start > end || start < 0 || end >= len(tokens) {
		return ""
	}
	var b strings.Builder
	for i := start; i <= end; i++ {
		if i > start && tokens[i].IsWhitespaceBefore() {
			b.WriteByte(' ')
		}
		b.WriteString(tokens[i].GetToken())
	}
	return b.String()
}
