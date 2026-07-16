package rules

import "fmt"

// ScoredConfusionSet ports org.languagetool.rules.ScoredConfusionSet.
type ScoredConfusionSet struct {
	Words []*ConfusionString
	Score float32
}

// NewScoredConfusionSet requires score > 0.
func NewScoredConfusionSet(score float32, words []*ConfusionString) *ScoredConfusionSet {
	if score <= 0 {
		panic(fmt.Sprintf("factor must be > 0: %v", score))
	}
	return &ScoredConfusionSet{
		Words: append([]*ConfusionString(nil), words...),
		Score: score,
	}
}

func (s *ScoredConfusionSet) GetScore() float32 { return s.Score }

// GetConfusionTokens returns the surface strings of the set.
func (s *ScoredConfusionSet) GetConfusionTokens() []string {
	out := make([]string, len(s.Words))
	for i, w := range s.Words {
		out[i] = w.GetString()
	}
	return out
}

// GetTokenDescriptions returns optional descriptions (nil = absent).
func (s *ScoredConfusionSet) GetTokenDescriptions() []*string {
	out := make([]*string, len(s.Words))
	for i, w := range s.Words {
		out[i] = w.GetDescription()
	}
	return out
}

func (s *ScoredConfusionSet) String() string {
	return fmt.Sprint(s.GetConfusionTokens())
}
