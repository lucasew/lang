package hunspell

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"

// DiscoverLangHunspellWordList delegates to spelling.DiscoverLangHunspellWordList.
func DiscoverLangHunspellWordList(shortCode, name string) string {
	return spelling.DiscoverLangHunspellWordList(shortCode, name)
}

// LoadSpellingWordListFile delegates to spelling.LoadSpellingWordListFile.
func LoadSpellingWordListFile(path string) ([]string, error) {
	return spelling.LoadSpellingWordListFile(path)
}

// ApplyDefaultSpellingWordLists delegates to spelling.ApplyDefaultSpellingWordLists
// (Java SpellingCheckRule.init word lists).
func ApplyDefaultSpellingWordLists(r *spelling.SpellingCheckRule) {
	spelling.ApplyDefaultSpellingWordLists(r)
}
