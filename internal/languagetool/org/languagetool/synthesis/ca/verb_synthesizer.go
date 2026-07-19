package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// VerbSynthesizer ports org.languagetool.synthesis.ca.VerbSynthesizer (1:1).
//
// Synthesize is language.getSynthesizer().synthesize(token, postag) (exact POS).
// AdaptSuggestion is language.adaptSuggestion (identity if nil).
type VerbSynthesizer struct {
	Tokens            []*languagetool.AnalyzedTokenReadings
	IFirstVerb        int
	ILastVerb         int
	NewLemma          string
	NewPostag         string
	NumPronounsBefore int
	NumPronounsAfter  int
	SearchBackward    bool

	// Synthesize ports language.getSynthesizer().synthesize(token, postag).
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
	// AdaptSuggestion ports language.adaptSuggestion; nil → identity.
	AdaptSuggestion func(s, originalErrorStr string) string
}

// Java public static Pattern fields.
var (
	PVerb           = regexp.MustCompile(`V.*`)
	PInflectedVerb  = regexp.MustCompile(`V.[SIM].*`)
	PImperativeVerb = regexp.MustCompile(`V.M.*`)
	PVerbIS         = regexp.MustCompile(`V.[IS].*`)
	pNonParticiple  = regexp.MustCompile(`V.[^P].*`)
	pParticiple     = regexp.MustCompile(`V.P.*`)
	// PronomsFeblesHelper.pPronomFeble (duplicated to avoid rules↔synthesis import cycle).
	pPronomFeble = regexp.MustCompile(`P0.{6}|PP3CN000|PP3NN000|PP3..A00|PP[123]CP000|PP3CSD00`)
)

// NewVerbSynthesizerAt ports VerbSynthesizer(tokens, startPos, lang[, searchBackward]).
// Language is represented via Synthesize / AdaptSuggestion hooks (set by caller).
func NewVerbSynthesizerAt(tokens []*languagetool.AnalyzedTokenReadings, startPos int, searchBackward bool) *VerbSynthesizer {
	v := &VerbSynthesizer{
		Tokens:            tokens,
		IFirstVerb:        -1,
		ILastVerb:         -1,
		NumPronounsBefore: -1,
		NumPronounsAfter:  -1,
		SearchBackward:    searchBackward,
	}
	if startPos >= 0 && startPos < len(tokens) {
		v.setIndexes(startPos)
	}
	return v
}

// NewVerbSynthesizer ports the no-searchBackward constructor defaulting startPos 0
// and leaving indexes unset until FindVerbGroup / NewVerbSynthesizerAt is used.
// Kept for existing tests that call FindVerbGroup after construction.
func NewVerbSynthesizer(tokens []*languagetool.AnalyzedTokenReadings) *VerbSynthesizer {
	return &VerbSynthesizer{
		Tokens:            tokens,
		IFirstVerb:        -1,
		ILastVerb:         -1,
		NumPronounsBefore: -1,
		NumPronounsAfter:  -1,
	}
}

// FindVerbGroup is a simplified scan used by older tests; prefer NewVerbSynthesizerAt.
func (v *VerbSynthesizer) FindVerbGroup() bool {
	if v == nil {
		return false
	}
	v.IFirstVerb, v.ILastVerb = -1, -1
	for i, tok := range v.Tokens {
		if tok == nil {
			continue
		}
		if vsFullMatchReading(tok, PVerb) != nil {
			if v.IFirstVerb < 0 {
				v.IFirstVerb = i
			}
			v.ILastVerb = i
		} else if v.IFirstVerb >= 0 {
			break
		}
	}
	if v.IFirstVerb >= 0 {
		v.NumPronounsBefore = 0
		v.NumPronounsAfter = 0
		return true
	}
	return false
}

// SetTarget sets the synthesised lemma and POS for the verb group.
func (v *VerbSynthesizer) SetTarget(lemma, postag string) {
	if v == nil {
		return
	}
	v.NewLemma = lemma
	v.NewPostag = postag
}

// HasTarget reports whether lemma and postag were set.
func (v *VerbSynthesizer) HasTarget() bool {
	return v != nil && v.NewLemma != "" && v.NewPostag != ""
}

// SetLemmaAndPostag ports setLemmaAndPostag.
func (v *VerbSynthesizer) SetLemmaAndPostag(lemma, postag string) {
	v.SetTarget(lemma, postag)
}

// SetPostag ports setPostag (lemma from last verb).
func (v *VerbSynthesizer) SetPostag(postag string) {
	if v == nil || v.ILastVerb < 0 || v.ILastVerb >= len(v.Tokens) {
		return
	}
	r := vsFullMatchReading(v.Tokens[v.ILastVerb], PVerb)
	if r != nil && r.GetLemma() != nil {
		v.NewLemma = *r.GetLemma()
	}
	v.NewPostag = postag
}

// SetLemma ports setLemma (postag from first verb).
func (v *VerbSynthesizer) SetLemma(lemma string) {
	if v == nil || v.IFirstVerb < 0 || v.IFirstVerb >= len(v.Tokens) {
		return
	}
	v.NewLemma = lemma
	r := vsFullMatchReading(v.Tokens[v.IFirstVerb], PVerb)
	if r != nil && r.GetPOSTag() != nil {
		v.NewPostag = *r.GetPOSTag()
	}
}

func (v *VerbSynthesizer) setIndexes(startPos int) {
	if v == nil || len(v.Tokens) == 0 {
		return
	}
	j := startPos
	if j < 0 || j >= len(v.Tokens) {
		return
	}
	// single participle
	if vsFullMatchReading(v.Tokens[j], pParticiple) != nil &&
		!v.Tokens[j].HasPosTag("_GV_") && !vsHasChunk(v.Tokens[j], "GV") {
		v.IFirstVerb = j
		v.ILastVerb = j
		v.NumPronounsBefore = 0
		v.NumPronounsAfter = 0
		return
	}
	// If it is not a verb, find the first one
	if v.SearchBackward {
		for j > 0 && !v.isVerb(j) {
			j--
		}
		foundSomeVerb := false
		for j > 0 && v.isVerb(j) {
			foundSomeVerb = true
			j--
		}
		if foundSomeVerb {
			j++
		}
	} else {
		for j < len(v.Tokens) && !v.isVerb(j) {
			j++
		}
	}

	if v.isVerb(j) {
		v.IFirstVerb = j
		v.ILastVerb = j
		// enrere
		i := j - 1
		for v.isMultitokenVerb(i) && !v.IsFirstVerbIS() {
			v.IFirstVerb = i
			i--
		}
		// avant
		i = j + 1
		for v.isMultitokenVerb(i) && !(v.IsFirstVerbIS() && v.isVerbIS(i)) {
			v.ILastVerb = i
			i++
		}
	} else {
		return
	}

	i := 1
	pronounsAfter := 0
	for v.ILastVerb+i < len(v.Tokens) &&
		!v.Tokens[v.ILastVerb+i].IsWhitespaceBefore() &&
		vsFullMatchReading(v.Tokens[v.ILastVerb+i], pPronomFeble) != nil {
		pronounsAfter++
		i++
	}
	v.NumPronounsAfter = pronounsAfter

	i = -1
	pronounsBeforeNoSpaceBefore := 0
	pronounsBefore := 0
	for v.IFirstVerb+i > 0 && vsFullMatchReading(v.Tokens[v.IFirstVerb+i], pPronomFeble) != nil {
		tok := v.Tokens[v.IFirstVerb+i]
		if tok.IsWhitespaceBefore() || v.IFirstVerb+i == 1 ||
			v.Tokens[v.IFirstVerb+i-1].HasPosTagStartingWith("_QM") {
			pronounsBefore = pronounsBefore + pronounsBeforeNoSpaceBefore + 1
			pronounsBeforeNoSpaceBefore = 0
		} else {
			pronounsBeforeNoSpaceBefore++
		}
		i--
	}
	v.NumPronounsBefore = pronounsBefore
}

func (v *VerbSynthesizer) isVerb(i int) bool {
	if v == nil || i < 0 || i > len(v.Tokens)-1 {
		return false
	}
	tok := v.Tokens[i]
	return vsHasChunk(tok, "GV") ||
		vsFullMatchReading(tok, pNonParticiple) != nil ||
		(vsFullMatchReading(tok, pParticiple) != nil && tok.HasPosTag("_GV_"))
}

func (v *VerbSynthesizer) isMultitokenVerb(i int) bool {
	if v == nil || i < 0 || i > len(v.Tokens)-1 {
		return false
	}
	tok := v.Tokens[i]
	return vsHasChunk(tok, "GV") || tok.HasPosTag("_GV_")
}

// SynthesizeForm ports synthesize() (renamed to avoid clash with the hook field).
func (v *VerbSynthesizer) SynthesizeForm() string {
	if v == nil || v.Synthesize == nil || v.IFirstVerb < 0 || v.ILastVerb < 0 {
		return ""
	}
	var result strings.Builder
	firstVerb := vsFullMatchReading(v.Tokens[v.IFirstVerb], PVerb)
	if v.IFirstVerb == v.ILastVerb {
		at := languagetool.NewAnalyzedToken("", &v.NewPostag, &v.NewLemma)
		forms := v.Synthesize(at, v.adjustPostagToLemma(v.NewLemma, v.NewPostag))
		if len(forms) > 0 {
			result.WriteString(forms[0])
		}
	} else {
		for i := v.IFirstVerb; i <= v.ILastVerb; i++ {
			if i == v.IFirstVerb {
				if firstVerb == nil {
					continue
				}
				lemma := ""
				if firstVerb.GetLemma() != nil {
					lemma = *firstVerb.GetLemma()
				}
				forms := v.Synthesize(firstVerb, v.adjustPostagToLemma(lemma, v.NewPostag))
				if len(forms) > 0 {
					result.WriteString(forms[0])
				}
			} else if i == v.ILastVerb {
				if v.Tokens[i].IsWhitespaceBefore() {
					result.WriteByte(' ')
				}
				lastR := vsFullMatchReading(v.Tokens[v.ILastVerb], PVerb)
				if lastR == nil || lastR.GetPOSTag() == nil {
					continue
				}
				postag := *lastR.GetPOSTag()
				at := languagetool.NewAnalyzedToken("", &postag, &v.NewLemma)
				forms := v.Synthesize(at, v.adjustPostagToLemma(v.NewLemma, postag))
				if len(forms) > 0 {
					result.WriteString(forms[0])
				}
			} else {
				if v.Tokens[i].IsWhitespaceBefore() {
					result.WriteByte(' ')
				}
				result.WriteString(v.Tokens[i].GetToken())
			}
		}
	}
	s := result.String()
	if v.AdaptSuggestion != nil {
		return v.AdaptSuggestion(s, "")
	}
	return s
}

func (v *VerbSynthesizer) adjustPostagToLemma(lemma, postag string) string {
	if lemma == "haver" && len(postag) >= 2 {
		postag = "VA" + postag[2:]
	}
	if lemma == "ser" && len(postag) >= 2 {
		postag = "VS" + postag[2:]
	}
	return postag
}

// GetStringFromTo ports getStringFromTo.
func (v *VerbSynthesizer) GetStringFromTo(start, end int) string {
	if v == nil || start > end || start < 0 || end >= len(v.Tokens) {
		return ""
	}
	var sb strings.Builder
	for i := start; i <= end; i++ {
		if i > start && v.Tokens[i].IsWhitespaceBefore() {
			sb.WriteByte(' ')
		}
		sb.WriteString(v.Tokens[i].GetToken())
	}
	return sb.String()
}

func (v *VerbSynthesizer) GetPronounsStrBefore() string {
	return v.GetStringFromTo(v.IFirstVerb-v.NumPronounsBefore, v.IFirstVerb-1)
}

func (v *VerbSynthesizer) GetPronounsStrAfter() string {
	return v.GetStringFromTo(v.ILastVerb+1, v.ILastVerb+v.NumPronounsAfter)
}

func (v *VerbSynthesizer) GetWholeOriginalStr() string {
	return v.GetStringFromTo(v.IFirstVerb-v.NumPronounsBefore, v.ILastVerb+v.NumPronounsAfter)
}

func (v *VerbSynthesizer) GetVerbStr() string {
	return v.GetStringFromTo(v.IFirstVerb, v.ILastVerb)
}

func (v *VerbSynthesizer) GetFirstVerbIndex() int { return v.IFirstVerb }
func (v *VerbSynthesizer) GetLastVerbIndex() int  { return v.ILastVerb }
func (v *VerbSynthesizer) GetLastIndex() int {
	return v.ILastVerb + v.NumPronounsAfter
}
func (v *VerbSynthesizer) GetNumPronounsAfter() int  { return v.NumPronounsAfter }
func (v *VerbSynthesizer) GetNumPronounsBefore() int { return v.NumPronounsBefore }

func (v *VerbSynthesizer) GetFirstVerbPersonaNumber() string {
	if v == nil || v.IFirstVerb < 0 || v.IFirstVerb >= len(v.Tokens) {
		return ""
	}
	r := vsFullMatchReading(v.Tokens[v.IFirstVerb], PInflectedVerb)
	if r != nil && r.GetPOSTag() != nil {
		tag := *r.GetPOSTag()
		if len(tag) >= 6 {
			return tag[4:6]
		}
	}
	return ""
}

func (v *VerbSynthesizer) GetFirstVerbPersonaNumberImperative() string {
	if v == nil || v.IFirstVerb < 0 || v.IFirstVerb >= len(v.Tokens) {
		return ""
	}
	r := vsFullMatchReading(v.Tokens[v.IFirstVerb], PImperativeVerb)
	if r != nil && r.GetPOSTag() != nil {
		tag := *r.GetPOSTag()
		if len(tag) >= 6 {
			return tag[4:6]
		}
	}
	return ""
}

// IsFirstVerbIS ports isFirstVerbIS.
func (v *VerbSynthesizer) IsFirstVerbIS() bool {
	if v == nil || v.IFirstVerb == -1 || v.IFirstVerb >= len(v.Tokens) {
		return false
	}
	return vsFullMatchReading(v.Tokens[v.IFirstVerb], PVerbIS) != nil
}

// GetFirstVerbISPostag ports getFirstVerbISPostag.
func (v *VerbSynthesizer) GetFirstVerbISPostag() string {
	if v == nil || v.IFirstVerb == -1 || v.IFirstVerb >= len(v.Tokens) {
		return ""
	}
	r := vsFullMatchReading(v.Tokens[v.IFirstVerb], PVerbIS)
	if r != nil && r.GetPOSTag() != nil {
		return *r.GetPOSTag()
	}
	return ""
}

func (v *VerbSynthesizer) isVerbIS(i int) bool {
	if v == nil || i < 0 || i >= len(v.Tokens) {
		return false
	}
	return vsFullMatchReading(v.Tokens[i], PVerbIS) != nil
}

func (v *VerbSynthesizer) GetCasingModel() string {
	return v.GetStringFromTo(v.IFirstVerb-v.NumPronounsBefore, v.IFirstVerb)
}

// IsUndefined ports isUndefined.
func (v *VerbSynthesizer) IsUndefined() bool {
	return v == nil || v.IFirstVerb == -1 || v.ILastVerb == -1 ||
		v.NumPronounsAfter == -1 || v.NumPronounsBefore == -1
}

// IsPassatPerifrastic ports isPassatPerifrastic.
func (v *VerbSynthesizer) IsPassatPerifrastic() bool {
	if v == nil || v.IFirstVerb < 1 || v.IFirstVerb+1 > len(v.Tokens)-1 {
		return false
	}
	tok := v.Tokens[v.IFirstVerb]
	next := v.Tokens[v.IFirstVerb+1]
	return tok.HasPosTagStartingWith("VA") && tok.HasAnyLemma("anar") &&
		(next.HasPartialPosTag("VMN") || next.HasPartialPosTag("VSN") || next.HasPartialPosTag("VAN"))
}

// IsPerfet ports isPerfet.
func (v *VerbSynthesizer) IsPerfet() bool {
	if v == nil || v.IFirstVerb < 1 || v.IFirstVerb+1 > len(v.Tokens)-1 {
		return false
	}
	tok := v.Tokens[v.IFirstVerb]
	next := v.Tokens[v.IFirstVerb+1]
	return tok.HasPosTagStartingWith("VA") && tok.HasAnyLemma("haver") &&
		(next.HasPartialPosTag("VMP") || next.HasPartialPosTag("VSP") || next.HasPartialPosTag("VAP"))
}

// vsFullMatch is Java Matcher.matches() (entire string).
func vsFullMatch(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

func vsFullMatchReading(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) *languagetool.AnalyzedToken {
	if tok == nil || re == nil {
		return nil
	}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		posTag := "UNKNOWN"
		if pt := r.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if vsFullMatch(re, posTag) {
			return r
		}
	}
	return nil
}

func vsHasChunk(tok *languagetool.AnalyzedTokenReadings, tag string) bool {
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
