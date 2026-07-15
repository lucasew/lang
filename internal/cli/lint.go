package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lucasew/lang/internal/engine"
	"github.com/lucasew/lang/internal/exitcode"
	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/format"
	"github.com/spf13/cobra"
)

func newLintCmd() *cobra.Command {
	var (
		langFlag   string
		formatFlag string
		disable    []string
		enableOnly []string
	)

	cmd := &cobra.Command{
		Use:   "lint [files...]",
		Short: "Lint text files (or stdin) for language issues",
		Long: `Lint files or stdin using the LanguageTool data under the configured data dir.

Examples:
  lang lint --lang en-US README.md
  echo 'This  is wrong' | lang lint --lang en
  lang lint --format json --lang auto file.txt
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmtName, err := format.Parse(formatFlag)
			if err != nil {
				return err
			}

			checker, err := engine.New(dataDirFlag(cmd))
			if err != nil {
				return err
			}

			opt := engine.Options{
				Language:      langFlag,
				DisabledRules: map[string]bool{},
				EnabledOnly:   map[string]bool{},
			}
			for _, id := range disable {
				for _, p := range strings.Split(id, ",") {
					p = strings.TrimSpace(p)
					if p != "" {
						opt.DisabledRules[p] = true
					}
				}
			}
			for _, id := range enableOnly {
				for _, p := range strings.Split(id, ",") {
					p = strings.TrimSpace(p)
					if p != "" {
						opt.EnabledOnly[p] = true
					}
				}
			}

			var all []finding.Finding
			hasErrorSev := false

			inputs, err := collectInputs(args)
			if err != nil {
				return err
			}
			for _, in := range inputs {
				res, err := checker.Check(in.name, in.text, opt)
				if err != nil {
					return err
				}
				for _, f := range res.Findings {
					if exitcode.IsErrorSeverity(f.Severity) {
						hasErrorSev = true
					}
					all = append(all, f)
				}
			}

			if err := format.Write(cmd.OutOrStdout(), fmtName, all); err != nil {
				return err
			}
			if hasErrorSev {
				return exitcode.HasErrorFindings()
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&langFlag, "lang", "l", "auto", "language code or auto")
	cmd.Flags().StringVar(&formatFlag, "format", "text", "output format: text, json, sarif")
	cmd.Flags().StringSliceVar(&disable, "disable", nil, "comma-separated rule IDs to disable")
	cmd.Flags().StringSliceVar(&enableOnly, "only", nil, "if set, only run these rule IDs")
	return cmd
}

type input struct {
	name string
	text string
}

func collectInputs(args []string) ([]input, error) {
	if len(args) == 0 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return []input{{name: "stdin", text: string(b)}}, nil
	}
	var out []input
	for _, a := range args {
		if a == "-" {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("read stdin: %w", err)
			}
			out = append(out, input{name: "stdin", text: string(b)})
			continue
		}
		b, err := os.ReadFile(a)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", a, err)
		}
		out = append(out, input{name: a, text: string(b)})
	}
	return out, nil
}
