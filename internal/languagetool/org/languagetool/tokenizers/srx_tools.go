package tokenizers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/attic/srx"
)

// createSrxDocument ports SrxTools.createSrxDocument(path):
// Java loads from resource dir via JLanguageTool.getDataBroker().getFromResourceDirAsStream(path).
// Known official resources are embedded/cached; unknown paths try filesystem candidates.
func createSrxDocument(srxInClassPath string) (*srx.Document, error) {
	path := normalizeSrxClasspath(srxInClassPath)
	if path == "/segment.srx" {
		return srx.DefaultDocument()
	}
	if path == "/org/languagetool/tokenizers/segment-simple.srx" {
		return segmentSimpleDocument()
	}
	return loadSrxDocumentFromFS(path)
}

func normalizeSrxClasspath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/segment.srx"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// loadSrxDocumentFromFS tries to resolve a classpath-style SRX path under known
// resource roots (inspiration submodule layout). Used for non-embedded SRX files.
func loadSrxDocumentFromFS(classpath string) (*srx.Document, error) {
	// Java resource dir root: org/languagetool/resource/ + path without leading /
	// Full class path under LT: .../resource + path, e.g. .../resource/segment.srx
	rel := strings.TrimPrefix(classpath, "/")
	candidates := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
			"org", "languagetool", "resource", rel),
		filepath.Join("internal", "languagetool", "org", "languagetool", "tokenizers", "data", filepath.Base(rel)),
	}
	var lastErr error
	for _, p := range candidates {
		doc, err := srx.Load(p)
		if err == nil && doc != nil {
			return doc, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("srx not found for classpath %q", classpath)
	}
	return nil, fmt.Errorf("Could not load SRX rules: %w", lastErr)
}

// srxDocumentCache memoizes createSrxDocument by classpath (Java constructs per tokenizer instance;
// caching is safe because SRX documents are immutable).
var (
	srxDocCacheMu sync.Mutex
	srxDocCache   = map[string]*srx.Document{}
)

func cachedCreateSrxDocument(srxInClassPath string) (*srx.Document, error) {
	path := normalizeSrxClasspath(srxInClassPath)
	srxDocCacheMu.Lock()
	if doc, ok := srxDocCache[path]; ok && doc != nil {
		srxDocCacheMu.Unlock()
		return doc, nil
	}
	srxDocCacheMu.Unlock()

	doc, err := createSrxDocument(path)
	if err != nil || doc == nil {
		return doc, err
	}
	srxDocCacheMu.Lock()
	srxDocCache[path] = doc
	srxDocCacheMu.Unlock()
	return doc, nil
}

// materializeEmbed writes embed bytes to a temp file for srx.Load (attic has no Parse(bytes) export).
func materializeEmbed(prefix string, data []byte) (string, error) {
	f, err := os.CreateTemp("", prefix+"-*.srx")
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return name, nil
}
