package xx

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// DemoTagger ports org.languagetool.tagging.xx.DemoTagger — assigns null POS tags.
type DemoTagger struct{}

func NewDemoTagger() *DemoTagger { return &DemoTagger{} }

func (DemoTagger) Tag(sentenceTokens []string) ([]*languagetool.AnalyzedTokenReadings, error) {
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	for _, word := range sentenceTokens {
		at := languagetool.NewAnalyzedToken(word, nil, nil)
		out = append(out, languagetool.NewAnalyzedTokenReadings(at))
	}
	return out, nil
}

func (DemoTagger) CreateNullToken(token string, startPos int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, nil, nil), startPos)
}

func (DemoTagger) CreateToken(token, posTag string) *languagetool.AnalyzedToken {
	var p *string
	if posTag != "" {
		p = &posTag
	}
	return languagetool.NewAnalyzedToken(token, p, nil)
}

var _ languagetool.Tagger = DemoTagger{}
