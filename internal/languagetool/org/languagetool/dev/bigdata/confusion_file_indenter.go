package bigdata

import (
	"strings"
	"unicode"
)

// IndentConfusionFile ports ConfusionFileIndenter.indent — re-aligns comments at column 82.
func IndentConfusionFile(lines []string) string {
	var b strings.Builder
	alreadyDone := map[string]struct{}{}
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && line != "" {
			// strip trailing comment for key
			data := line
			if i := strings.Index(data, "#"); i >= 0 {
				data = strings.TrimSpace(data[:i])
			}
			parts := strings.Split(data, ";")
			if len(parts) >= 2 {
				key := strings.TrimSpace(parts[0]) + ";" + strings.TrimSpace(parts[1])
				if _, ok := alreadyDone[key]; ok {
					continue
				}
				alreadyDone[key] = struct{}{}
			}
		}
		commentPos := strings.LastIndex(line, "#")
		if commentPos <= 0 {
			b.WriteString(line)
			b.WriteByte('\n')
			continue
		}
		endData := commentPos
		for endData > 0 && unicode.IsSpace(rune(line[endData-1])) {
			endData--
		}
		spaces := 82 - endData
		if spaces < 1 {
			spaces = 1
		}
		b.WriteString(line[:endData])
		b.WriteString(strings.Repeat(" ", spaces))
		b.WriteString(line[commentPos:])
		b.WriteByte('\n')
	}
	return b.String()
}
