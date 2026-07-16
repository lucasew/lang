package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Proto-free DTO stand-ins for MLServerProto messages used by GRPCUtils.

// GRPCAnalyzedToken ports MLServerProto.AnalyzedToken fields.
type GRPCAnalyzedToken struct {
	Token  string `json:"token"`
	PosTag string `json:"posTag,omitempty"`
	Lemma  string `json:"lemma,omitempty"`
}

// GRPCAnalyzedTokenReadings ports MLServerProto.AnalyzedTokenReadings.
type GRPCAnalyzedTokenReadings struct {
	StartPos  int                 `json:"startPos"`
	ChunkTags []string            `json:"chunkTags,omitempty"`
	Readings  []GRPCAnalyzedToken `json:"readings"`
}

// GRPCAnalyzedSentence ports MLServerProto.AnalyzedSentence.
type GRPCAnalyzedSentence struct {
	Text   string                      `json:"text"`
	Tokens []GRPCAnalyzedTokenReadings `json:"tokens"`
}

// GRPCMatch ports MLServerProto.Match subset for wire conversion.
type GRPCMatch struct {
	Offset                int      `json:"offset"`
	Length                int      `json:"length"`
	ID                    string   `json:"id"`
	SubID                 string   `json:"subId,omitempty"`
	SuggestedReplacements []string `json:"suggestedReplacements,omitempty"`
	RuleDescription       string   `json:"ruleDescription,omitempty"`
	MatchDescription      string   `json:"matchDescription,omitempty"`
	MatchShortDescription string   `json:"matchShortDescription,omitempty"`
	URL                   string   `json:"url,omitempty"`
	AutoCorrect           bool     `json:"autoCorrect,omitempty"`
	Type                  string   `json:"type,omitempty"`
}

// TokenToGRPC ports GRPCUtils.toGRPC(AnalyzedToken).
func TokenToGRPC(token *languagetool.AnalyzedToken) GRPCAnalyzedToken {
	if token == nil {
		return GRPCAnalyzedToken{}
	}
	out := GRPCAnalyzedToken{Token: token.GetToken()}
	if pt := token.GetPOSTag(); pt != nil {
		out.PosTag = *pt
	}
	if lm := token.GetLemma(); lm != nil {
		out.Lemma = *lm
	}
	return out
}

// TokenFromGRPC ports GRPCUtils.fromGRPC(AnalyzedToken).
func TokenFromGRPC(t GRPCAnalyzedToken) *languagetool.AnalyzedToken {
	var pos, lemma *string
	if t.PosTag != "" {
		p := t.PosTag
		pos = &p
	}
	if t.Lemma != "" {
		l := t.Lemma
		lemma = &l
	}
	return languagetool.NewAnalyzedToken(t.Token, pos, lemma)
}

// ReadingsToGRPC ports GRPCUtils.toGRPC(AnalyzedTokenReadings).
func ReadingsToGRPC(r *languagetool.AnalyzedTokenReadings) GRPCAnalyzedTokenReadings {
	if r == nil {
		return GRPCAnalyzedTokenReadings{}
	}
	out := GRPCAnalyzedTokenReadings{
		StartPos:  r.GetStartPos(),
		ChunkTags: append([]string(nil), r.GetChunkTags()...),
	}
	for _, rd := range r.GetReadings() {
		out.Readings = append(out.Readings, TokenToGRPC(rd))
	}
	return out
}

// ReadingsFromGRPC ports GRPCUtils.fromGRPC(AnalyzedTokenReadings).
func ReadingsFromGRPC(r GRPCAnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	var tokens []*languagetool.AnalyzedToken
	for _, t := range r.Readings {
		tokens = append(tokens, TokenFromGRPC(t))
	}
	if len(tokens) == 0 {
		tokens = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken("", nil, nil)}
	}
	atr := languagetool.NewAnalyzedTokenReadingsList(tokens, r.StartPos)
	if len(r.ChunkTags) > 0 {
		atr.SetChunkTags(r.ChunkTags)
	}
	return atr
}

// SentenceToGRPC ports GRPCUtils.toGRPC(AnalyzedSentence).
func SentenceToGRPC(s *languagetool.AnalyzedSentence) GRPCAnalyzedSentence {
	if s == nil {
		return GRPCAnalyzedSentence{}
	}
	out := GRPCAnalyzedSentence{Text: s.GetText()}
	for _, tok := range s.GetTokens() {
		out.Tokens = append(out.Tokens, ReadingsToGRPC(tok))
	}
	return out
}

// SentenceFromGRPC ports GRPCUtils.fromGRPC(AnalyzedSentence).
func SentenceFromGRPC(s GRPCAnalyzedSentence) *languagetool.AnalyzedSentence {
	var toks []*languagetool.AnalyzedTokenReadings
	for _, t := range s.Tokens {
		toks = append(toks, ReadingsFromGRPC(t))
	}
	return languagetool.NewAnalyzedSentence(toks)
}

// MatchToGRPC converts a RuleMatch to a GRPCMatch DTO.
func MatchToGRPC(m *RuleMatch) GRPCMatch {
	if m == nil {
		return GRPCMatch{}
	}
	id := "UNKNOWN"
	if r, ok := m.Rule.(interface{ GetID() string }); ok {
		id = r.GetID()
	}
	return GRPCMatch{
		Offset:                m.FromPos,
		Length:                m.ToPos - m.FromPos,
		ID:                    id,
		SuggestedReplacements: append([]string(nil), m.SuggestedReplacements...),
		MatchDescription:      m.Message,
		MatchShortDescription: m.ShortMessage,
	}
}

// MatchFromGRPC builds a RuleMatch from GRPCMatch (rule is a FakeRule with ID).
func MatchFromGRPC(m GRPCMatch, sentence *languagetool.AnalyzedSentence) *RuleMatch {
	rule := NewFakeRule(m.ID)
	rm := NewRuleMatch(rule, sentence, m.Offset, m.Offset+m.Length, m.MatchDescription)
	if m.MatchShortDescription != "" {
		rm.ShortMessage = m.MatchShortDescription
	}
	if len(m.SuggestedReplacements) > 0 {
		rm.SetSuggestedReplacements(m.SuggestedReplacements)
	}
	return rm
}

// NormalizeWhitespaceForGRPC ports GRPCRule whitespace normalisation regex idea.
func NormalizeWhitespaceForGRPC(s string) string {
	// replace non-breaking / special spaces with regular space
	replacer := []rune{'\u00a0', '\u202f', '\ufeff', '\ufffd'}
	out := []rune(s)
	for i, r := range out {
		for _, bad := range replacer {
			if r == bad {
				out[i] = ' '
				break
			}
		}
	}
	return string(out)
}

// GRPCUtils is the Java-name twin for conversion helpers.
type GRPCUtils struct{}

func (GRPCUtils) TokenToGRPC(tok *languagetool.AnalyzedToken) GRPCAnalyzedToken {
	return TokenToGRPC(tok)
}
func (GRPCUtils) ReadingsToGRPC(r *languagetool.AnalyzedTokenReadings) GRPCAnalyzedTokenReadings {
	return ReadingsToGRPC(r)
}
func (GRPCUtils) SentenceToGRPC(s *languagetool.AnalyzedSentence) GRPCAnalyzedSentence {
	return SentenceToGRPC(s)
}
