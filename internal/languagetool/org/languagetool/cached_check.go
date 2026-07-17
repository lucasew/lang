package languagetool

// CachedCheck runs Check with ResultCache keyed by language + text + disabled rule set.
// Full InputSentence keying deferred; multi-sentence uses a synthetic whole-text sentence key.
func CachedCheck(cache *ResultCache, lt *JLanguageTool, text string) []LocalMatch {
	if lt == nil {
		return nil
	}
	if cache == nil {
		return lt.Check(text)
	}
	sents := lt.Analyze(text)
	if len(sents) == 0 {
		return nil
	}
	var keySent *AnalyzedSentence
	if len(sents) == 1 {
		keySent = sents[0]
	} else {
		// soft multi-sentence key: whole text as one analyzed surface
		keySent = AnalyzePlain(text)
	}
	key := NewInputSentence(keySent, lt.LanguageCode, "", lt.DisabledRuleIDs, nil, nil, nil, nil, nil, string(lt.Mode), lt.Level, nil, nil)
	if v, ok := cache.GetMatchesIfPresent(key); ok {
		if ms, ok := v.([]LocalMatch); ok {
			return ms
		}
	}
	ms := lt.Check(text)
	cache.PutMatches(key, ms)
	return ms
}
