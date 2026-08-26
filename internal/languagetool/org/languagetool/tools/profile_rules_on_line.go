package tools

// ProfileRulesOnLine ports Tools.profileRulesOnLine:
//
//	count = 0
//	for sentence := range lt.sentenceTokenize(contents) {
//	  count += rule.match(lt.getAnalyzedSentence(sentence)).length
//	}
//
// Callers supply tokenize + per-sentence match count to avoid JLanguageTool cycles.
func ProfileRulesOnLine(contents string, sentenceTokenize func(string) []string, matchCount func(sentence string) int) int {
	if sentenceTokenize == nil || matchCount == nil {
		return 0
	}
	count := 0
	for _, sentence := range sentenceTokenize(contents) {
		count += matchCount(sentence)
	}
	return count
}
