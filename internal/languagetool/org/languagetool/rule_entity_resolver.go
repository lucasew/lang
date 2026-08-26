package languagetool

import (
	"io"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
)

// RuleEntityResolver ports org.languagetool.RuleEntityResolver.
// Resolves .ent entity systemIds relative to language resource folders.
type RuleEntityResolver struct {
	Broker broker.ResourceDataBroker
}

func NewRuleEntityResolver(b broker.ResourceDataBroker) *RuleEntityResolver {
	return &RuleEntityResolver{Broker: b}
}

var resourceFolderRE = regexp.MustCompile(`.*/resource/`)

// GetPathFromLTResourceFolder ports getPathFromLTResourceFolder.
func (r *RuleEntityResolver) GetPathFromLTResourceFolder(input string) string {
	s := resourceFolderRE.ReplaceAllString(input, "")
	s = strings.ReplaceAll(s, "../", "")
	return s
}

// ResolveEntity returns a reader for .ent files, or nil if not an entity path.
func (r *RuleEntityResolver) ResolveEntity(systemID string) (io.ReadCloser, error) {
	if systemID == "" || !strings.HasSuffix(systemID, ".ent") {
		return nil, nil
	}
	if r.Broker == nil {
		return nil, nil
	}
	path := r.GetPathFromLTResourceFolder(systemID)
	return r.Broker.GetFromResourceDirAsStream(path)
}
