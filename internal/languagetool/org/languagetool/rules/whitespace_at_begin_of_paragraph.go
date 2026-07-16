package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WhiteSpaceAtBeginOfParagraph ports org.languagetool.rules.WhiteSpaceAtBeginOfParagraph.
type WhiteSpaceAtBeginOfParagraph struct {
	Messages map[string]string
}

func NewWhiteSpaceAtBeginOfParagraph(messages map[string]string) *WhiteSpaceAtBeginOfParagraph {
	return &WhiteSpaceAtBeginOfParagraph{Messages: messages}
}

func (r *WhiteSpaceAtBeginOfParagraph) GetID() string { return "WHITESPACE_PARAGRAPH_BEGIN" }

func isWhitespaceDel(token *languagetool.AnalyzedTokenReadings) bool {
	return token.IsWhitespace() && token.GetToken() != "\u200B" && !token.IsLinebreak()
}

func (r *WhiteSpaceAtBeginOfParagraph) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	i := 1
	for i < len(tokens) && isWhitespaceDel(tokens[i]) {
		i++
	}
	if i > 1 && i < len(tokens) && !tokens[i].IsLinebreak() {
		msg := r.Messages["whitespace_at_begin_parapgraph_msg"]
		if msg == "" {
			msg = "Don't start a paragraph with whitespace"
		}
		rm := NewRuleMatch(r, sentence, tokens[1].GetStartPos(), tokens[i].GetEndPos(), msg)
		rm.SetSuggestedReplacement(tokens[i].GetToken())
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}
