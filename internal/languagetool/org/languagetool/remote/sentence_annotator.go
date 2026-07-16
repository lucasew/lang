package remote

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

// AnnotatorConfig ports SentenceAnnotator configuration properties.
type AnnotatorConfig struct {
	RemoteServer   string
	UserName       string
	AnnotatorName  string
	APIKey         string
	InputFilePath  string
	OutputFilePath string
	Language       string
	MotherTongue   string
}

// DefaultAnnotatorConfig returns localhost defaults.
func DefaultAnnotatorConfig() AnnotatorConfig {
	return AnnotatorConfig{
		RemoteServer: "http://localhost:8081",
		Language:     "en-US",
	}
}

// SentenceAnnotator ports org.languagetool.remote.SentenceAnnotator utilities
// for offline annotation workflows (without full interactive loop).
type SentenceAnnotator struct {
	Config AnnotatorConfig
	Client *RemoteLanguageTool
	// Cache of sentence → matches (keyed by MD5 of text+lang).
	Cache map[string][]*RemoteRuleMatch
}

func NewSentenceAnnotator(cfg AnnotatorConfig) *SentenceAnnotator {
	base := cfg.RemoteServer
	for len(base) > 0 && base[len(base)-1] == '/' {
		base = base[:len(base)-1]
	}
	return &SentenceAnnotator{
		Config: cfg,
		Client: NewRemoteLanguageTool(base),
		Cache:  map[string][]*RemoteRuleMatch{},
	}
}

// CacheKey builds a stable cache key for a sentence.
func CacheKey(text, lang string) string {
	sum := md5.Sum([]byte(lang + "\n" + text))
	return hex.EncodeToString(sum[:])
}

// AnnotateSentence checks one sentence, using cache when possible.
func (a *SentenceAnnotator) AnnotateSentence(text string) ([]*RemoteRuleMatch, error) {
	if a == nil {
		return nil, fmt.Errorf("nil annotator")
	}
	lang := a.Config.Language
	if lang == "" {
		lang = "en-US"
	}
	key := CacheKey(text, lang)
	if m, ok := a.Cache[key]; ok {
		return m, nil
	}
	res, err := a.Client.Check(text, lang)
	if err != nil {
		return nil, err
	}
	var matches []*RemoteRuleMatch
	if res != nil {
		matches = res.Matches
	}
	if a.Cache == nil {
		a.Cache = map[string][]*RemoteRuleMatch{}
	}
	a.Cache[key] = matches
	return matches, nil
}

// TimestampPrefix returns YYYY-MM-DD for output filenames.
func TimestampPrefix(t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	return t.Format("2006-01-02")
}
