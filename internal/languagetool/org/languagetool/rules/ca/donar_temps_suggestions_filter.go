package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DonarTempsSuggestionsFilter ports
// org.languagetool.rules.ca.DonarTempsSuggestionsFilter (1:1 AcceptRuleMatch).
//
// Synthesize is Synthesizer.synthesize(AnalyzedToken, postag) without POS-regex.
// When nil, Accept returns nil (fail-closed; Java always has a synthesizer).
//
// SynthHaver / SynthTenir remain for unit Suggest helpers (optional).
type DonarTempsSuggestionsFilter struct {
	// Synthesize ports getSynthesizerFromRuleMatch(...).synthesize(token, postag).
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
	// SynthHaver synthesizes "haver" with VA+suffix (legacy Suggest path).
	SynthHaver func(verbPostagSuffix string) string
	// SynthTenir synthesizes "tenir" for the given postag (legacy Suggest path).
	SynthTenir func(postag string) string
}

func NewDonarTempsSuggestionsFilter() *DonarTempsSuggestionsFilter {
	return &DonarTempsSuggestionsFilter{}
}

// DonarTempsInput holds pre-analyzed span pieces (legacy unit helper).
type DonarTempsInput struct {
	PronomGenderNumber string
	AuxTokens          []string
	VerbPostag         string
	CasingModel        string
}

// Suggest returns "hi ha temps" / "tinc temps" style replacements (unit helper).
func (f *DonarTempsSuggestionsFilter) Suggest(in DonarTempsInput) []string {
	if len(in.VerbPostag) < 8 {
		return nil
	}
	var out []string
	if f.SynthHaver != nil {
		suffix := in.VerbPostag[2:8]
		haver := f.SynthHaver(suffix)
		if haver != "" {
			var b strings.Builder
			b.WriteString("hi")
			for _, tok := range in.AuxTokens {
				b.WriteByte(' ')
				b.WriteString(tok)
			}
			b.WriteString(" ")
			b.WriteString(haver)
			b.WriteString(" temps")
			s := strings.ReplaceAll(b.String(), "de haver", "d'haver")
			if in.CasingModel != "" {
				s = tools.PreserveCase(s, in.CasingModel)
			}
			out = append(out, s)
		}
	}
	if f.SynthTenir != nil {
		var s string
		if len(in.AuxTokens) == 0 {
			postag := in.VerbPostag[:4] + in.PronomGenderNumber + in.VerbPostag[6:8]
			tenir := f.SynthTenir(postag)
			if tenir != "" {
				s = tenir + " temps"
			}
		} else {
			tenir := f.SynthTenir(in.VerbPostag)
			if tenir != "" {
				var b strings.Builder
				for i, tok := range in.AuxTokens {
					if i > 0 {
						b.WriteByte(' ')
					}
					b.WriteString(tok)
				}
				b.WriteString(" ")
				b.WriteString(tenir)
				b.WriteString(" temps")
				s = b.String()
			}
		}
		if s != "" {
			if in.CasingModel != "" {
				s = tools.PreserveCase(s, in.CasingModel)
			}
			out = append(out, s)
		}
	}
	return out
}

// PronomGenderNumberFromP extracts person+number from a P… postag.
func PronomGenderNumberFromP(postag string) string {
	if len(postag) < 5 {
		return ""
	}
	return string(postag[2]) + string(postag[4])
}

// AcceptRuleMatch ports DonarTempsSuggestionsFilter.acceptRuleMatch.
func (f *DonarTempsSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = arguments
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	// Without synthesis cannot produce forms (Java always has synthesizer).
	if f.Synthesize == nil && f.SynthHaver == nil && f.SynthTenir == nil {
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

	pr := readingWithTagRegex(tokens[posWord], `P.*`)
	if pr == nil || pr.GetPOSTag() == nil {
		return nil
	}
	pronomPostag := *pr.GetPOSTag()
	if len(pronomPostag) < 5 {
		return nil
	}
	pronomGenderNumber := pronomPostag[2:3] + pronomPostag[4:5]

	indexFirstVerb := posWord + 1
	if indexFirstVerb >= len(tokens) {
		return nil
	}
	indexMainVerb := indexFirstVerb
	for indexMainVerb < len(tokens) && !tokens[indexMainVerb].HasAnyLemma("donar") {
		indexMainVerb++
	}
	if indexMainVerb >= len(tokens) {
		return nil
	}
	// Need temps after donar for end pos (Java uses indexMainVerb+1).
	if indexMainVerb+1 >= len(tokens) {
		return nil
	}

	vr := readingWithTagRegex(tokens[indexMainVerb], `V.*`)
	if vr == nil || vr.GetPOSTag() == nil {
		return nil
	}
	verbPostag := *vr.GetPOSTag()
	if len(verbPostag) < 8 {
		return nil
	}

	var replacements []string

	// haver-hi temps
	if sugg1 := f.buildHaverHi(tokens, indexFirstVerb, indexMainVerb, verbPostag, tokens[posWord].GetToken()); sugg1 != "" {
		replacements = append(replacements, sugg1)
	}
	// tenir temps
	if sugg2 := f.buildTenir(tokens, indexFirstVerb, indexMainVerb, verbPostag, pronomGenderNumber, tokens[posWord].GetToken()); sugg2 != "" {
		replacements = append(replacements, sugg2)
	}
	if len(replacements) == 0 {
		return nil
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[posWord].GetStartPos(), tokens[indexMainVerb+1].GetEndPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacements(replacements)
	return out
}

func (f *DonarTempsSuggestionsFilter) synthLemma(lemma, postag string) []string {
	if f.Synthesize != nil {
		at := languagetool.NewAnalyzedToken("", nil, &lemma)
		return f.Synthesize(at, postag)
	}
	// Legacy hooks used by unit Suggest
	if lemma == "haver" && f.SynthHaver != nil && strings.HasPrefix(postag, "VA") && len(postag) >= 2 {
		if form := f.SynthHaver(postag[2:]); form != "" {
			return []string{form}
		}
	}
	if lemma == "tenir" && f.SynthTenir != nil {
		if form := f.SynthTenir(postag); form != "" {
			return []string{form}
		}
	}
	return nil
}

func (f *DonarTempsSuggestionsFilter) synthToken(tok *languagetool.AnalyzedToken, postag string) []string {
	if f.Synthesize != nil {
		return f.Synthesize(tok, postag)
	}
	return nil
}

func (f *DonarTempsSuggestionsFilter) buildHaverHi(tokens []*languagetool.AnalyzedTokenReadings,
	indexFirstVerb, indexMainVerb int, verbPostag, casingModel string) string {
	forms := f.synthLemma("haver", "VA"+verbPostag[2:8])
	if len(forms) == 0 {
		return ""
	}
	var suggestion1 strings.Builder
	index := indexFirstVerb
	suggestion1.WriteString("hi")
	for index < indexMainVerb {
		if tokens[index].IsWhitespaceBefore() || suggestion1.Len() == 2 {
			suggestion1.WriteByte(' ')
		}
		suggestion1.WriteString(tokens[index].GetToken())
		index++
	}
	suggestion1.WriteString(" " + forms[0] + " temps")
	sugg1 := strings.ReplaceAll(suggestion1.String(), "de haver", "d'haver")
	return tools.PreserveCase(sugg1, casingModel)
}

func (f *DonarTempsSuggestionsFilter) buildTenir(tokens []*languagetool.AnalyzedTokenReadings,
	indexFirstVerb, indexMainVerb int, verbPostag, pronomGenderNumber, casingModel string) string {
	var suggestion2 strings.Builder
	index := indexFirstVerb
	if index == indexMainVerb {
		// direct: tenir with person from pronoun
		forms := f.synthLemma("tenir", verbPostag[:4]+pronomGenderNumber+verbPostag[6:8])
		if len(forms) == 0 {
			return ""
		}
		suggestion2.WriteString(forms[0] + " temps")
	} else {
		// re-inflect first auxiliary to pronom person, keep middles, tenir for main POS
		at2 := tokens[indexFirstVerb].GetAnalyzedToken(0)
		if at2 == nil || at2.GetPOSTag() == nil {
			return ""
		}
		auxPostag := *at2.GetPOSTag()
		if len(auxPostag) < 8 {
			return ""
		}
		forms2 := f.synthToken(at2, auxPostag[:4]+pronomGenderNumber+auxPostag[6:8])
		if len(forms2) == 0 {
			// Legacy path without full token synth: only if no aux re-inflection available
			return ""
		}
		suggestion2.WriteString(forms2[0])
		index++
		for index < indexMainVerb {
			if tokens[index].IsWhitespaceBefore() {
				suggestion2.WriteByte(' ')
			}
			suggestion2.WriteString(tokens[index].GetToken())
			index++
		}
		mainAt := tokens[indexMainVerb].GetAnalyzedToken(0)
		if mainAt == nil || mainAt.GetPOSTag() == nil {
			return ""
		}
		forms3 := f.synthLemma("tenir", *mainAt.GetPOSTag())
		if len(forms3) == 0 {
			return ""
		}
		suggestion2.WriteString(" " + forms3[0] + " temps")
	}
	sugg2 := suggestion2.String()
	if sugg2 == "" {
		return ""
	}
	return tools.PreserveCase(sugg2, casingModel)
}
