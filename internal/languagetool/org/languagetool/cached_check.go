package languagetool

// CachedCheck runs Check with ResultCache keyed by language + text + disabled rule set.
// Full InputSentence keying deferred; this is a soft demo cache for inject checkers.
func CachedCheck(cache *ResultCache, lt *JLanguageTool, text string) []LocalMatch {
	if lt == nil {
		return nil
	}
	if cache == nil {
		return lt.Check(text)
	}
	// build a lightweight key via Analyze + InputSentence
	sents := lt.Analyze(text)
	if len(sents) == 0 {
		return nil
	}
	// cache per first sentence surface + lang (soft; multi-sentence uses full Check)
	if len(sents) == 1 {
		key := NewInputSentence(sents[0], lt.LanguageCode, "", lt.DisabledRuleIDs, nil, nil, nil, nil, nil, string(lt.Mode), lt.Level, nil, nil)
		if v, ok := cache.GetMatchesIfPresent(key); ok {
			if ms, ok := v.([]LocalMatch); ok {
				return ms
			}
		}
		ms := lt.Check(text)
		cache.PutMatches(key, ms)
		return ms
	}
	return lt.Check(text)
}
