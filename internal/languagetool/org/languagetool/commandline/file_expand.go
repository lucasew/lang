package commandline

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ExpandInputPaths expands files and optional recursive directories into a flat file list.
// Directories are walked when recursive is true; only regular text-like files are included.
func ExpandInputPaths(paths []string, recursive bool) ([]string, error) {
	if len(paths) == 0 {
		return []string{""}, nil // stdin
	}
	var out []string
	seen := map[string]struct{}{}
	add := func(p string) {
		if p == "" || p == "-" {
			if _, ok := seen["-"]; !ok {
				seen["-"] = struct{}{}
				out = append(out, "-")
			}
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	for _, p := range paths {
		if p == "" || p == "-" {
			add("-")
			continue
		}
		st, err := os.Stat(p)
		if err != nil {
			// keep path; load will surface the error
			add(p)
			continue
		}
		if !st.IsDir() {
			add(p)
			continue
		}
		if !recursive {
			// directory without -r: soft-skip (no error) — LT typically requires files
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// skip hidden dirs
				base := d.Name()
				if base != "." && strings.HasPrefix(base, ".") {
					return fs.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return nil
			}
			if !isTextyFilename(path) {
				return nil
			}
			add(path)
			return nil
		})
		if err != nil {
			return out, err
		}
	}
	if len(out) == 0 {
		return []string{""}, nil
	}
	return out, nil
}

func isTextyFilename(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case "", ".txt", ".md", ".markdown", ".rst", ".text", ".html", ".htm", ".xml", ".json", ".csv", ".log", ".go", ".py", ".java", ".js", ".ts", ".css":
		return true
	default:
		// no extension or unknown — include plain names without dots
		base := filepath.Base(path)
		return !strings.Contains(base, ".")
	}
}
