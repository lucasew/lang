// Command lang — LanguageTool pure-Go port (WIP).
// SPEC product surface uses Cobra; bare LT-style flags still work.
package main

import (
	"os"

	"github.com/lucasew/lang/internal/cli"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
)

func main() {
	commandline.VersionString = "languagetool-go (dev)"
	// Legacy LT-style when first arg is a flag or empty → commandline.Run
	if len(os.Args) <= 1 || (len(os.Args) > 1 && len(os.Args[1]) > 0 && os.Args[1][0] == '-') {
		os.Exit(commandline.Run(os.Args[1:], commandline.DefaultCoreHooks()))
	}
	os.Exit(cli.Execute())
}
