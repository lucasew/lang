package languagetool

// DynamicHunspellLanguage ports org.languagetool.DynamicHunspellLanguage metadata.
// Speller wiring stays pluggable (full HunspellRule needs language stack).
type DynamicHunspellLanguage struct {
	DynamicLanguage
}

func NewDynamicHunspellLanguage(name, code, dictPath string) DynamicHunspellLanguage {
	return DynamicHunspellLanguage{DynamicLanguage: NewDynamicLanguage(name, code, dictPath)}
}

// SpellerRuleID returns e.g. EN-US_SPELLER_RULE from code.
func (d DynamicHunspellLanguage) SpellerRuleID() string {
	return toUpperCode(d.Code) + "_SPELLER_RULE"
}

// DictFilenameInResources strips .dic suffix like Java getDictFilenameInResources.
func (d DynamicHunspellLanguage) DictFilenameInResources() string {
	p := d.DictPath
	if len(p) > 4 && p[len(p)-4:] == ".dic" {
		return p[:len(p)-4]
	}
	return p
}

func toUpperCode(code string) string {
	b := make([]byte, len(code))
	for i := 0; i < len(code); i++ {
		c := code[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
