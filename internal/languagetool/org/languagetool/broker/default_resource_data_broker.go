package broker

import (
	"io/fs"
	"os"
)

// DefaultResourceDataBroker ports org.languagetool.broker.DefaultResourceDataBroker
// as an FSResourceDataBroker with standard resource/rules roots.
// Prefer embedding assets; when fsys is nil, uses os.DirFS(".").
type DefaultResourceDataBroker = FSResourceDataBroker

// NewDefaultResourceDataBroker creates a broker with default dir names.
func NewDefaultResourceDataBroker() *FSResourceDataBroker {
	return NewFSResourceDataBroker(os.DirFS("."), "org/languagetool/resource", "org/languagetool/rules")
}

// NewDefaultResourceDataBrokerFS is DefaultResourceDataBroker with a custom FS.
func NewDefaultResourceDataBrokerFS(fsys fs.FS, resourceDir, rulesDir string) *FSResourceDataBroker {
	if resourceDir == "" {
		resourceDir = "org/languagetool/resource"
	}
	if rulesDir == "" {
		rulesDir = "org/languagetool/rules"
	}
	if fsys == nil {
		fsys = os.DirFS(".")
	}
	return NewFSResourceDataBroker(fsys, resourceDir, rulesDir)
}
