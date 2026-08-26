package tools

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DictionaryExporter ports org.languagetool.tools.DictionaryExporter surface.
// Binary decompile is deferred; helpers classify dict kind and describe export.
type DictionaryExporter struct {
	*DictionaryBuilder
}

func NewDictionaryExporter(info map[string]string) *DictionaryExporter {
	return &DictionaryExporter{DictionaryBuilder: NewDictionaryBuilder(info)}
}

// IsSpellingDictPath reports hunspell/spelling dict paths (Java uses FSADecompile).
func IsSpellingDictPath(path string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "hunspell") || strings.Contains(p, "spelling")
}

// ExportMode describes which decompiler would be used.
type ExportMode string

const (
	ExportModeFSA  ExportMode = "fsa"
	ExportModeDict ExportMode = "dict"
)

func ExportModeFor(path string) ExportMode {
	if IsSpellingDictPath(path) {
		return ExportModeFSA
	}
	return ExportModeDict
}

// DescribeExport returns a human-readable plan string (no I/O).
func (e *DictionaryExporter) DescribeExport(binaryPath string) string {
	mode := ExportModeFor(binaryPath)
	out := e.GetOutputFilename()
	if out == "" {
		out = filepath.Base(binaryPath) + ".txt"
	}
	return fmt.Sprintf("export %s via %s to %s", binaryPath, mode, out)
}
