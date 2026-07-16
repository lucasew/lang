package language

import "fmt"

// RuleFilenameException ports org.languagetool.language.RuleFilenameException.
type RuleFilenameException struct {
	Filename string
}

func NewRuleFilenameException(filename string) *RuleFilenameException {
	return &RuleFilenameException{Filename: filename}
}

func (e *RuleFilenameException) Error() string {
	return fmt.Sprintf("Rule file must be named rules-<xx>-<lang>.xml (<xx> = language code, "+
		"<lang> = language name),\n"+
		"for example: rules-en-English.xml\n"+
		"Current name: %s", e.Filename)
}
