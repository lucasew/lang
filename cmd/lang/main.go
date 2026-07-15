package main

import (
	"os"

	"github.com/lucasew/lang/internal/cli"
	"github.com/lucasew/lang/internal/exitcode"
)

func main() {
	if err := cli.Execute(); err != nil {
		code := exitcode.FromError(err)
		if code == 0 {
			code = exitcode.ToolFailure
		}
		os.Exit(code)
	}
}
