// Command lang-server — pure-Go LanguageTool HTTP API (WIP).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/server"
)

func main() {
	port := flag.Int("port", server.DefaultPort, "listen port")
	host := flag.String("host", "127.0.0.1", "listen host")
	public := flag.Bool("public", false, "allow non-loopback clients (disables IP allowlist)")
	allowOrigin := flag.String("allow-origin", "", "CORS Access-Control-Allow-Origin value")
	dataDir := flag.String("data-dir", "", "soft data root (sets LANG_DATA_DIR for grammar/false-friends)")
	grammarDir := flag.String("grammar-dir", "", "soft grammar XML dir (sets LANG_GRAMMAR_DIR)")
	falseFriends := flag.String("falsefriends", "", "soft false-friends XML path (sets LANG_FALSEFRIENDS_FILE)")
	demoSpeller := flag.Bool("demo-speller", false, "enable LANG_DEMO_SPELLER=1 for EN map speller inject")
	flag.Parse()

	if *dataDir != "" {
		_ = os.Setenv("LANG_DATA_DIR", *dataDir)
	}
	if *grammarDir != "" {
		_ = os.Setenv("LANG_GRAMMAR_DIR", *grammarDir)
	}
	if *falseFriends != "" {
		_ = os.Setenv("LANG_FALSEFRIENDS_FILE", *falseFriends)
	}
	if *demoSpeller {
		_ = os.Setenv("LANG_DEMO_SPELLER", "1")
	}

	cfg := server.NewHTTPServerConfig()
	cfg.Port = *port
	cfg.PublicAccess = *public
	if *allowOrigin != "" {
		cfg.AllowOriginURL = *allowOrigin
	}
	var allowed map[string]struct{}
	if !*public {
		allowed = server.DefaultAllowedIPs
	}
	srv := server.NewHTTPServerConfig2(cfg, false, *host, allowed)
	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Fprintf(os.Stderr, "lang-server listening on http://%s (v2/check, v2/languages, v2/info)\n", addr)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
