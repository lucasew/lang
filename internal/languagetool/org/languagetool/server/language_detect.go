package server

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
)

// DetectLanguageOfString ports TextChecker.detectLanguageOfString surface.
// When preferred is non-empty and detect returns empty, falls back to preferred[0].
// detect may be nil → commandline.DetectLanguageHeuristic.
func DetectLanguageOfString(text string, preferredVariants []string, detect func(string) string) string {
	fn := detect
	if fn == nil {
		fn = commandline.DetectLanguageHeuristic
	}
	code := fn(text)
	if code == "" && len(preferredVariants) > 0 {
		// preferredVariants like en-US → base en if needed
		code = preferredVariants[0]
	}
	// if preferred has a variant for detected base, promote
	if code != "" && len(preferredVariants) > 0 {
		base := code
		if i := strings.Index(code, "-"); i > 0 {
			base = code[:i]
		}
		for _, v := range preferredVariants {
			if strings.HasPrefix(strings.ToLower(v), strings.ToLower(base)+"-") ||
				strings.EqualFold(v, code) {
				return v
			}
		}
	}
	return code
}
