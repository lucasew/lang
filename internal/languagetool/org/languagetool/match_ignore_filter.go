package languagetool

import "strings"

// AddIgnoreWord records a surface form to suppress matches that cover only that token.
func (lt *JLanguageTool) AddIgnoreWord(word string) {
	if lt == nil || word == "" {
		return
	}
	if lt.IgnoreWords == nil {
		lt.IgnoreWords = map[string]struct{}{}
	}
	lt.IgnoreWords[word] = struct{}{}
	lt.IgnoreWords[strings.ToLower(word)] = struct{}{}
}

// AddIgnoreWords records multiple ignore surface forms.
func (lt *JLanguageTool) AddIgnoreWords(words ...string) {
	for _, w := range words {
		lt.AddIgnoreWord(w)
	}
}

func (lt *JLanguageTool) filterMatchesByIgnore(text string, ms []LocalMatch) []LocalMatch {
	if lt == nil || len(ms) == 0 {
		return ms
	}
	if (lt.IgnoreWords == nil || len(lt.IgnoreWords) == 0) &&
		(lt.UserConfig == nil || (len(lt.UserConfig.UserSpecificSpellerWords) == 0 && len(lt.UserConfig.AcceptedPhrases) == 0)) {
		return ms
	}
	// build ignore set from IgnoreWords + user speller words
	ign := map[string]struct{}{}
	for w := range lt.IgnoreWords {
		ign[w] = struct{}{}
		ign[strings.ToLower(w)] = struct{}{}
	}
	if lt.UserConfig != nil {
		for _, w := range lt.UserConfig.UserSpecificSpellerWords {
			ign[w] = struct{}{}
			ign[strings.ToLower(w)] = struct{}{}
		}
	}
	out := make([]LocalMatch, 0, len(ms))
	for _, m := range ms {
		if m.FromPos < 0 || m.ToPos > len(text) || m.FromPos >= m.ToPos {
			out = append(out, m)
			continue
		}
		surface := text[m.FromPos:m.ToPos]
		// drop spelling-like matches on ignored words
		if isSpellRuleID(m.RuleID) {
			if _, ok := ign[surface]; ok {
				continue
			}
			if _, ok := ign[strings.ToLower(surface)]; ok {
				continue
			}
		}
		// drop any match fully covered by an accepted phrase
		if lt.UserConfig != nil && lt.UserConfig.AcceptsPhrase(surface) {
			continue
		}
		if lt.UserConfig != nil && lt.UserConfig.AcceptsPhrase(strings.ToLower(surface)) {
			continue
		}
		out = append(out, m)
	}
	return out
}

func isSpellRuleID(id string) bool {
	if id == "" {
		return false
	}
	u := strings.ToUpper(id)
	return strings.Contains(u, "MORFOLOGIK") || strings.Contains(u, "SPELL") ||
		strings.Contains(u, "HUNSPELL") || strings.HasPrefix(u, "SPELLING")
}

// FilterMatchesByIgnoreWords drops spelling matches on the given surface forms.
func FilterMatchesByIgnoreWords(text string, ms []LocalMatch, words []string) []LocalMatch {
	if len(ms) == 0 || len(words) == 0 {
		return ms
	}
	lt := NewJLanguageTool("en")
	lt.AddIgnoreWords(words...)
	return lt.filterMatchesByIgnore(text, ms)
}
