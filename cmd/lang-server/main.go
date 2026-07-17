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
	flag.Parse()

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
	fmt.Fprintf(os.Stderr, "lang-server listening on http://%s (v2/check, v2/languages)\n", addr)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
