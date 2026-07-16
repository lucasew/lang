package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PatternRuleCheck runs loaded pattern/regex rules over text using simple tokenization.
type PatternRuleCheck struct {
	WordTokenizer     tokenizers.Tokenizer
	SentenceTokenizer tokenizers.SentenceTokenizer
	PatternRules      []*PatternRule
	RegexRules        []*RegexPatternRule
	Listener          rules.RuleMatchListener
}

func NewPatternRuleCheck() *PatternRuleCheck {
	return &PatternRuleCheck{
		WordTokenizer:     tokenizers.NewWordTokenizer(),
		SentenceTokenizer: tokenizers.NewSimpleSentenceTokenizer().AsSentenceTokenizer(),
	}
}

// FromHandler copies rules out of a PatternRuleHandler.
func (c *PatternRuleCheck) FromHandler(h *PatternRuleHandler) *PatternRuleCheck {
	if h == nil {
		return c
	}
	c.PatternRules = append(c.PatternRules, h.LoadedPatternRules...)
	c.RegexRules = append(c.RegexRules, h.LoadedRegexRules...)
	return c
}

// Check tokenizes text, analyzes each sentence as surface tokens, and matches rules.
func (c *PatternRuleCheck) Check(text string) ([]*rules.RuleMatch, error) {
	if text == "" {
		return nil, nil
	}
	var all []*rules.RuleMatch
	sents := c.SentenceTokenizer.Tokenize(text)
	if len(sents) == 0 {
		sents = []string{text}
	}
	offset := 0
	for _, sentText := range sents {
		words := c.WordTokenizer.Tokenize(sentText)
		var toks []*languagetool.AnalyzedTokenReadings
		pos := offset
		for _, w := range words {
			// skip pure whitespace tokens for matching (keep positions)
			if isAllSpace(w) {
				pos += len(w) // byte length; approximate for ASCII demos
				continue
			}
			toks = append(toks, languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(w, nil, nil), pos))
			pos += len(w)
		}
		sent := languagetool.NewAnalyzedSentence(toks)
		for _, pr := range c.PatternRules {
			ms, err := pr.Match(sent)
			if err != nil {
				return all, err
			}
			for _, m := range ms {
				rules.NotifyListeners(m, c.Listener)
				all = append(all, m)
			}
		}
		for _, rr := range c.RegexRules {
			ms, err := rr.Match(sent)
			if err != nil {
				return all, err
			}
			for _, m := range ms {
				// adjust if sentence GetText is only non-ws join — positions are absolute via start pos
				rules.NotifyListeners(m, c.Listener)
				all = append(all, m)
			}
		}
		offset += len(sentText)
	}
	return all, nil
}

func isAllSpace(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return s != ""
}
