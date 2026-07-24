package dev

import (
	"bufio"
	"io"
	"strings"
)

// RemoveDuplicateLines ports org.languagetool.dev.DuplicateRemover core logic:
// print unique non-comment lines in order; comments always printed.
func RemoveDuplicateLines(r io.Reader, w io.Writer) error {
	seen := map[string]struct{}{}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "#") {
			if _, err := io.WriteString(w, line+"\n"); err != nil {
				return err
			}
			seen[line] = struct{}{}
			continue
		}
		if _, ok := seen[line]; ok {
			seen[line] = struct{}{}
			continue
		}
		seen[line] = struct{}{}
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return err
		}
	}
	return sc.Err()
}
