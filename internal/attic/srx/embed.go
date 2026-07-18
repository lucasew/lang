package srx

import (
	"bytes"
	_ "embed"
	"sync"
)

//go:embed data/segment.srx
var segmentSRX []byte

var (
	defaultDocOnce sync.Once
	defaultDoc     *Document
	defaultDocErr  error
)

// DefaultDocument returns the LanguageTool segment.srx rules embedded at build
// time (same resource Java loads via /segment.srx on the classpath).
func DefaultDocument() (*Document, error) {
	defaultDocOnce.Do(func() {
		defaultDoc, defaultDocErr = parse(bytes.NewReader(segmentSRX))
	})
	return defaultDoc, defaultDocErr
}
