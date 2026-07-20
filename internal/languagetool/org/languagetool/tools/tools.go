package tools

import (
	"fmt"
	"runtime/debug"
	"strings"
	"text/template"
)

// I18n ports Tools.i18n MessageFormat-style substitution for "{0}", "{1}", …
// after ResourceBundle.getString. Supports MessageFormat quote rules:
//   '' → literal apostrophe
//   '…' → quoted literal (placeholders not expanded inside)
// Complex format types ({0,number}) are not used by LT callers of this helper.
func I18n(pattern string, messageArguments ...any) string {
	if len(messageArguments) == 0 {
		return pattern
	}
	var b strings.Builder
	b.Grow(len(pattern))
	for i := 0; i < len(pattern); {
		c := pattern[i]
		if c == '\'' {
			if i+1 < len(pattern) && pattern[i+1] == '\'' {
				b.WriteByte('\'')
				i += 2
				continue
			}
			// quoted section until next unpaired '
			i++
			for i < len(pattern) {
				if pattern[i] == '\'' {
					if i+1 < len(pattern) && pattern[i+1] == '\'' {
						b.WriteByte('\'')
						i += 2
						continue
					}
					i++ // end quote
					break
				}
				b.WriteByte(pattern[i])
				i++
			}
			continue
		}
		if c == '{' {
			j := i + 1
			for j < len(pattern) && pattern[j] >= '0' && pattern[j] <= '9' {
				j++
			}
			if j > i+1 && j < len(pattern) && pattern[j] == '}' {
				idx := 0
				for k := i + 1; k < j; k++ {
					idx = idx*10 + int(pattern[k]-'0')
				}
				if idx >= 0 && idx < len(messageArguments) {
					b.WriteString(fmt.Sprint(messageArguments[idx]))
				}
				i = j + 1
				continue
			}
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// I18nTemplate formats with text/template when pattern uses {{.Args}}.
// Not a Java Tools twin — keep for Go-only call sites.
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

// GetFullStackTrace ports Tools.getFullStackTrace(Throwable):
// message + stack frames (Java printStackTrace). Plain Go errors have no
// capture stack; append the caller's stack via debug.Stack.
func GetFullStackTrace(err error) string {
	if err == nil {
		return ""
	}
	return err.Error() + "\n" + string(debug.Stack())
}

// ConsistencyRulePrefix is the default Language.getConsistencyRulePrefix.
const ConsistencyRulePrefix = "PREFIXFORCONSISTENCYRULES_"
