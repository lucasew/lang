package languagetool

import "strings"

// RuleFileExists reports whether a rules path is available (pluggable).
type RuleFileExists func(path string) bool

// GetRuleFileNames ports Language.getRuleFileNames for a short code / variant.
// rulesDir is typically the broker rules dir (e.g. "/org/languagetool/rules").
func GetRuleFileNames(shortCode, shortCodeWithVariant, rulesDir string, exists RuleFileExists) []string {
	if rulesDir == "" {
		rulesDir = "/org/languagetool/rules"
	}
	if exists == nil {
		exists = func(string) bool { return true }
	}
	var ruleFiles []string
	// always include grammar.xml path (Java always adds it)
	ruleFiles = append(ruleFiles, rulesDir+"/"+shortCode+"/"+PatternFile)
	style := shortCode + "/" + StyleFile
	if exists(style) {
		ruleFiles = append(ruleFiles, rulesDir+"/"+style)
	}
	custom := shortCode + "/" + CustomPatternFile
	if exists(custom) {
		ruleFiles = append(ruleFiles, rulesDir+"/"+custom)
	}
	variant := shortCodeWithVariant
	if variant == "" {
		variant = shortCode
	}
	// if variant longer than 2 (e.g. en-US), try en/en-US/grammar.xml
	if len(normalizeCode(variant)) > 2 {
		fileName := shortCode + "/" + normalizeCode(variant) + "/" + PatternFile
		if exists(fileName) {
			ruleFiles = append(ruleFiles, rulesDir+"/"+fileName)
		}
	}
	return ruleFiles
}

func normalizeCode(code string) string {
	return strings.ReplaceAll(code, "_", "-")
}
