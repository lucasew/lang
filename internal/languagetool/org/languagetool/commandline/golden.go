package commandline

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// CoreGoldenHook writes SPEC findings JSON for text (soft golden dump).
func CoreGoldenHook(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
	if opts != nil && opts.XMLFiltering {
		text = MaybeFilterXML(text, true)
	}
	lang := "en"
	if opts != nil {
		if opts.Language != "" {
			lang = opts.Language
		}
		if opts.IsAutoDetect() {
			lang = ResolveLanguage(text, opts, DetectLanguageHeuristic)
		}
	}
	lt, err := configureCoreLT(lang, opts)
	if err != nil {
		return 0, err
	}
	checker := &CoreRulesChecker{
		Lang:          lang,
		lt:            lt,
		CleanOverlaps: opts != nil && opts.CleanOverlapping,
	}
	if opts != nil {
		checker.DisabledCategories = append([]string(nil), opts.DisabledCategories...)
		checker.EnabledCategories = append([]string(nil), opts.EnabledCategories...)
		checker.UseEnabledOnly = opts.IsUseEnabledOnly()
	}
	ms, err := checker.Check(text)
	if err != nil {
		return 0, err
	}
	if opts != nil && len(opts.GetRuleValues()) > 0 {
		ms = applyCLIRuleValues(lang, text, ms, opts.GetRuleValues())
	}
	fn := ""
	if opts != nil {
		fn = opts.Filename
	}
	findings := MatchesToFindings(ms, text, fn)
	if err := WriteFindingsJSON(w, findings); err != nil {
		return 0, err
	}
	return countFailOnMatches(ms, optsGetFailOn(opts)), nil
}

func optsGetFailOn(opts *CommandLineOptions) string {
	if opts == nil {
		return "error"
	}
	return opts.GetFailOn()
}

// CoreCompareHook loads golden JSON and compares to live findings.
func CoreCompareHook(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
	path := ""
	if opts != nil {
		path = opts.CompareGoldenPath
	}
	if path == "" {
		path = os.Getenv("LANG_GOLDEN_FILE")
	}
	if path == "" {
		return 0, fmt.Errorf("compare requires golden file (compare GOLDEN.json … or LANG_GOLDEN_FILE)")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var want []Finding
	if err := json.Unmarshal(raw, &want); err != nil {
		return 0, fmt.Errorf("golden JSON: %w", err)
	}
	if opts != nil && opts.XMLFiltering {
		text = MaybeFilterXML(text, true)
	}
	lang := "en"
	if opts != nil {
		if opts.Language != "" {
			lang = opts.Language
		}
		if opts.IsAutoDetect() {
			lang = ResolveLanguage(text, opts, DetectLanguageHeuristic)
		}
	}
	lt, err := configureCoreLT(lang, opts)
	if err != nil {
		return 0, err
	}
	checker := &CoreRulesChecker{Lang: lang, lt: lt, CleanOverlaps: opts != nil && opts.CleanOverlapping}
	if opts != nil {
		checker.DisabledCategories = append([]string(nil), opts.DisabledCategories...)
		checker.EnabledCategories = append([]string(nil), opts.EnabledCategories...)
		checker.UseEnabledOnly = opts.IsUseEnabledOnly()
	}
	ms, err := checker.Check(text)
	if err != nil {
		return 0, err
	}
	if opts != nil && len(opts.GetRuleValues()) > 0 {
		ms = applyCLIRuleValues(lang, text, ms, opts.GetRuleValues())
	}
	fn := ""
	if opts != nil {
		fn = opts.Filename
	}
	got := MatchesToFindings(ms, text, fn)
	diff := CompareFindings(got, want)
	if diff == "" {
		_, _ = fmt.Fprintln(w, "OK: findings match golden")
		return 0, nil
	}
	_, _ = io.WriteString(w, diff)
	return 1, nil
}
