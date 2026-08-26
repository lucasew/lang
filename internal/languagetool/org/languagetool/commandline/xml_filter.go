package commandline

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// MaybeFilterXML applies tools.FilterXML when enabled (ports getFilteredText xmlFiltering branch).
func MaybeFilterXML(contents string, xmlFiltering bool) string {
	if !xmlFiltering {
		return contents
	}
	return tools.FilterXML(contents)
}
