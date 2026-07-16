package languagetool

// Tagger ports org.languagetool.tagging.Tagger — language-dependent POS tagging.
// Lives in package languagetool to avoid import cycles with tagging.WordTagger adapters.
type Tagger interface {
	Tag(sentenceTokens []string) ([]*AnalyzedTokenReadings, error)
	CreateNullToken(token string, startPos int) *AnalyzedTokenReadings
	CreateToken(token, posTag string) *AnalyzedToken
}

// FuncTagger adapts functions to Tagger.
type FuncTagger struct {
	TagFn             func(sentenceTokens []string) ([]*AnalyzedTokenReadings, error)
	CreateNullTokenFn func(token string, startPos int) *AnalyzedTokenReadings
	CreateTokenFn     func(token, posTag string) *AnalyzedToken
}

func (f FuncTagger) Tag(sentenceTokens []string) ([]*AnalyzedTokenReadings, error) {
	if f.TagFn == nil {
		return nil, nil
	}
	return f.TagFn(sentenceTokens)
}

func (f FuncTagger) CreateNullToken(token string, startPos int) *AnalyzedTokenReadings {
	if f.CreateNullTokenFn != nil {
		return f.CreateNullTokenFn(token, startPos)
	}
	return NewAnalyzedTokenReadingsAt(NewAnalyzedToken(token, nil, nil), startPos)
}

func (f FuncTagger) CreateToken(token, posTag string) *AnalyzedToken {
	if f.CreateTokenFn != nil {
		return f.CreateTokenFn(token, posTag)
	}
	var p *string
	if posTag != "" {
		p = &posTag
	}
	return NewAnalyzedToken(token, p, nil)
}
