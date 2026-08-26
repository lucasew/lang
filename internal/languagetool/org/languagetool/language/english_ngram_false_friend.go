package language

// EnglishHasNGramFalseFriendRule ports English.hasNGramFalseFriendRule.
// Java: motherTongue short codes de, fr, es, nl only — not invent extra L2s.
func EnglishHasNGramFalseFriendRule(motherTongueShortCode string) bool {
	if motherTongueShortCode == "" {
		return false
	}
	// primary subtag
	base := motherTongueShortCode
	for i := 0; i < len(motherTongueShortCode); i++ {
		if motherTongueShortCode[i] == '-' || motherTongueShortCode[i] == '_' {
			base = motherTongueShortCode[:i]
			break
		}
	}
	switch base {
	case "de", "fr", "es", "nl", "DE", "FR", "ES", "NL":
		return true
	}
	// case-fold 2-letter codes
	if len(base) == 2 {
		a, b := base[0], base[1]
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		switch string([]byte{a, b}) {
		case "de", "fr", "es", "nl":
			return true
		}
	}
	return false
}
