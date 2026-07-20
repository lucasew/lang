package languagetool

import (
	"os"
	"strings"
)

// UseUpstreamGrammar reports whether official grammar/style/variant XML should
// load (Java Language.getRuleFileNames always does). Default is on.
//
// Opt out for debug / missing-resource environments:
//
//	LANG_USE_UPSTREAM_GRAMMAR=0
//	LANG_USE_UPSTREAM_GRAMMAR=false
//	LANG_USE_UPSTREAM_GRAMMAR=no
//	LANG_USE_UPSTREAM_GRAMMAR=off
//
// Any other non-empty value (including "1" / "true") enables loading.
func UseUpstreamGrammar() bool {
	v := strings.TrimSpace(os.Getenv("LANG_USE_UPSTREAM_GRAMMAR"))
	if v == "" {
		return true
	}
	switch strings.ToLower(v) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}
