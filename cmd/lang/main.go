// Command lang — LanguageTool 1:1 Go port (WIP).
// Production code lives under internal/languagetool (LT-shaped).
// Prior product-shaped code is in internal/attic until 1:1 salvage.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "lang: rebuild in progress — LT-shaped port under internal/languagetool (see SPEC / twin audit)")
	fmt.Fprintln(os.Stderr, "prior implementation archived at internal/attic")
	os.Exit(2)
}
