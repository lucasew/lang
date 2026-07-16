package tools

import (
	"fmt"
	"strings"
	"text/template"
)

// I18n formats a message pattern with positional arguments (ports Tools.i18n
// MessageFormat-style "{0}" placeholders approximately via fmt).
func I18n(pattern string, messageArguments ...any) string {
	if len(messageArguments) == 0 {
		return pattern
	}
	// Replace {0}, {1}, ... with %v then Sprintf
	out := pattern
	for i := range messageArguments {
		out = strings.ReplaceAll(out, fmt.Sprintf("{%d}", i), fmt.Sprintf("%%[%d]v", i+1))
	}
	// fallback: if no {n} placeholders, return pattern
	if out == pattern {
		return pattern
	}
	return fmt.Sprintf(out, messageArguments...)
}

// I18nTemplate formats with text/template when pattern uses {{.Args}}.
func I18nTemplate(pattern string, data any) (string, error) {
	t, err := template.New("i18n").Parse(pattern)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

// CorrectListToString joins items for display (ports Tools helpers used in messages).
func CorrectListToString(items []string, lastSep string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " " + lastSep + " " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", " + lastSep + " " + items[len(items)-1]
	}
}
