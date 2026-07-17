// Package cli is a soft Cobra/Viper front-end for the LT-shaped commandline package (SPEC §2).
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Execute runs the Cobra root command.
func Execute() int {
	commandline.VersionString = "languagetool-go (dev)"

	var (
		lang, format, dataDir, failOn, mother, level string
		disable, enable, ruleValues                  string
		disableCats, enableCats                      string
		enabledOnly, recursive                       bool
	)

	root := &cobra.Command{
		Use:   "lang",
		Short: "LanguageTool pure-Go CLI linter",
		Long:  "lang is a pure-Go LanguageTool port. Subcommands map to the LT-shaped commandline engine.",
		// bare `lang` with files → soft lint (SPEC primary product)
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// default product: lint
			runEngine(buildLintArgs(lang, format, dataDir, failOn, mother, level, disable, enable, ruleValues, disableCats, enableCats, enabledOnly, recursive, args))
		},
	}
	root.PersistentFlags().StringVar(&dataDir, "data-dir", "", "soft data root (grammar + false-friends)")
	root.PersistentFlags().StringVarP(&lang, "lang", "l", "", "language code (default auto for lint)")
	_ = viper.BindPFlag("data-dir", root.PersistentFlags().Lookup("data-dir"))
	_ = viper.BindEnv("data-dir", "LANG_DATA_DIR", "LANG_DATA")
	_ = viper.BindPFlag("lang", root.PersistentFlags().Lookup("lang"))
	_ = viper.BindEnv("lang", "LANG_LANG")

	// shared flags for lint-like commands
	addLintFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&format, "format", "text", "output format: text|json|sarif|xml|plaintext")
		c.Flags().StringVar(&failOn, "fail-on", "error", "severity threshold: error|warning|note")
		c.Flags().StringVarP(&mother, "mothertongue", "m", "", "mother tongue for false friends")
		c.Flags().StringVar(&level, "level", "", "DEFAULT or PICKY")
		c.Flags().StringVarP(&disable, "disable", "d", "", "comma-separated disabled rule IDs")
		c.Flags().StringVarP(&enable, "enable", "e", "", "comma-separated enabled rule IDs")
		c.Flags().StringVar(&ruleValues, "ruleValues", "", "RULE_ID:value pairs")
		c.Flags().StringVar(&disableCats, "disablecategories", "", "comma-separated disabled categories")
		c.Flags().StringVar(&enableCats, "enablecategories", "", "comma-separated enabled categories")
		c.Flags().BoolVar(&enabledOnly, "only", false, "only run rules listed in --enable")
		c.Flags().BoolVarP(&recursive, "recursive", "r", false, "recurse into directories")
	}

	lintCmd := &cobra.Command{
		Use:   "lint [files...]",
		Short: "Lint text files or stdin (SPEC primary command)",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// viper default data-dir
			if dataDir == "" {
				dataDir = viper.GetString("data-dir")
			}
			if lang == "" {
				lang = viper.GetString("lang")
			}
			runEngine(buildLintArgs(lang, format, dataDir, failOn, mother, level, disable, enable, ruleValues, disableCats, enableCats, enabledOnly, recursive, args))
		},
	}
	addLintFlags(lintCmd)

	root.AddCommand(lintCmd)
	root.AddCommand(&cobra.Command{
		Use:   "languages",
		Short: "List supported language codes",
		Run:   func(cmd *cobra.Command, args []string) { runEngine([]string{"--list"}) },
	})
	root.AddCommand(&cobra.Command{
		Use:   "rules",
		Short: "List registered rule IDs for --lang",
		Run: func(cmd *cobra.Command, args []string) {
			a := []string{"--list-rules"}
			if lang != "" {
				a = append(a, "-l", lang)
			} else if v := viper.GetString("lang"); v != "" {
				a = append(a, "-l", v)
			}
			runEngine(a)
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "doctor",
		Short: "Environment / self-check diagnostics",
		Run: func(cmd *cobra.Command, args []string) {
			a := []string{"--doctor"}
			if dataDir != "" {
				a = append(a, "--data-dir", dataDir)
			} else if v := viper.GetString("data-dir"); v != "" {
				a = append(a, "--data-dir", v)
			}
			runEngine(a)
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "golden [files...]",
		Short: "Dump SPEC findings JSON (goldens)",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			a := []string{"--golden"}
			if lang != "" {
				a = append(a, "-l", lang)
			}
			if dataDir != "" {
				a = append(a, "--data-dir", dataDir)
			}
			a = append(a, fileArgs(args)...)
			runEngine(a)
		},
	})
	var goldenPath string
	compareCmd := &cobra.Command{
		Use:   "compare GOLDEN.json [files...]",
		Short: "Compare live findings to a golden file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			a := []string{"--compare", args[0]}
			if lang != "" {
				a = append(a, "-l", lang)
			}
			a = append(a, fileArgs(args[1:])...)
			runEngine(a)
		},
	}
	_ = goldenPath
	root.AddCommand(compareCmd)
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run:   func(cmd *cobra.Command, args []string) { runEngine([]string{"--version"}) },
	})

	// Also allow legacy LT-style flags when first arg starts with -
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-") {
		return commandline.Run(os.Args[1:], commandline.DefaultCoreHooks())
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return exitCode
}

// exitCode is set by runEngine (cobra Run cannot return int).
var exitCode int

func runEngine(args []string) {
	exitCode = commandline.Run(args, commandline.DefaultCoreHooks())
	if exitCode != 0 {
		// cobra still returns nil from Run; main uses exitCode
		os.Exit(exitCode)
	}
}

func fileArgs(args []string) []string {
	if len(args) == 0 {
		return []string{"-"}
	}
	return args
}

func buildLintArgs(lang, format, dataDir, failOn, mother, level, disable, enable, ruleValues, disableCats, enableCats string, enabledOnly, recursive bool, files []string) []string {
	a := []string{"--lint"}
	if format != "" && format != "text" && format != "lint" {
		a = []string{"--format", format}
	}
	if lang != "" {
		a = append(a, "-l", lang)
	}
	// empty lang → product lint auto-detect via commandline
	if dataDir != "" {
		a = append(a, "--data-dir", dataDir)
	}
	if failOn != "" && failOn != "error" {
		a = append(a, "--fail-on", failOn)
	}
	if mother != "" {
		a = append(a, "-m", mother)
	}
	if level != "" {
		a = append(a, "--level", level)
	}
	if disable != "" {
		a = append(a, "-d", disable)
	}
	if enable != "" {
		a = append(a, "-e", enable)
	}
	if enabledOnly {
		a = append(a, "--enabledonly")
	}
	if recursive {
		a = append(a, "--recursive")
	}
	if ruleValues != "" {
		a = append(a, "--ruleValues", ruleValues)
	}
	if disableCats != "" {
		a = append(a, "--disablecategories", disableCats)
	}
	if enableCats != "" {
		a = append(a, "--enablecategories", enableCats)
	}
	a = append(a, fileArgs(files)...)
	return a
}
