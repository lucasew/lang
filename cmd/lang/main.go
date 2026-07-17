// Command lang — LanguageTool pure-Go port (WIP).
// Production code lives under internal/languagetool (LT-shaped).
package main

import (
	"os"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
)

func main() {
	commandline.VersionString = "languagetool-go (dev)"
	os.Exit(commandline.Run(os.Args[1:], commandline.DefaultCoreHooks()))
}
