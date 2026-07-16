package language

import (
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// LanguageMetaFromFile extends LanguageMeta with the rules XML path.
type LanguageMetaFromFile struct {
	languagetool.LanguageMeta
	RulesFile  string
	Additional bool
}

// MakeAdditionalLanguage ports LanguageBuilder.makeAdditionalLanguage.
// File name must be rules-<code>-<Name>.xml (e.g. rules-de-German.xml).
func MakeAdditionalLanguage(filename string) (LanguageMetaFromFile, error) {
	base := filepath.Base(filename)
	if !strings.HasSuffix(base, ".xml") {
		return LanguageMetaFromFile{}, NewRuleFilenameException(filename)
	}
	parts := strings.Split(base, "-")
	if len(parts) < 3 || parts[0] != "rules" {
		return LanguageMetaFromFile{}, NewRuleFilenameException(filename)
	}
	code := parts[1]
	// Java: length 2, 3 (ast), or en_US style
	okLen := len(code) == 2 || len(code) == 3 || len(code) == 5 || strings.Contains(code, "_")
	if !okLen {
		return LanguageMetaFromFile{}, NewRuleFilenameException(filename)
	}
	namePart := strings.TrimSuffix(strings.Join(parts[2:], "-"), ".xml")
	meta := LanguageMetaFromFile{
		LanguageMeta: languagetool.LanguageMeta{
			Name: namePart,
			Code: code,
		},
		RulesFile:  filename,
		Additional: true,
	}
	if !languagetool.GlobalLanguages.IsLanguageSupported(code) {
		languagetool.GlobalLanguages.Register(meta.LanguageMeta)
	}
	return meta, nil
}
