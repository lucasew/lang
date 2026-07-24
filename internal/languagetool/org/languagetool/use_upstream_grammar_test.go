package languagetool

import (
	"os"
	"testing"
)

func TestUseUpstreamGrammar_DefaultOn(t *testing.T) {
	t.Cleanup(func() { _ = os.Unsetenv("LANG_USE_UPSTREAM_GRAMMAR") })

	_ = os.Unsetenv("LANG_USE_UPSTREAM_GRAMMAR")
	if !UseUpstreamGrammar() {
		t.Fatal("unset env should default on (Java always loads getRuleFileNames)")
	}
	t.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "1")
	if !UseUpstreamGrammar() {
		t.Fatal("1 should enable")
	}
	t.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "true")
	if !UseUpstreamGrammar() {
		t.Fatal("true should enable")
	}
	t.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "0")
	if UseUpstreamGrammar() {
		t.Fatal("0 should disable")
	}
	t.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "false")
	if UseUpstreamGrammar() {
		t.Fatal("false should disable")
	}
	t.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "off")
	if UseUpstreamGrammar() {
		t.Fatal("off should disable")
	}
}
