package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Execute runs the root command.
func Execute() error {
	root := &cobra.Command{
		Use:           "lang",
		Short:         "LanguageTool-compatible grammar/style linter (Go port)",
		Long:          "lang is a pure-Go reimplementation of LanguageTool as a CLI linter.\nSee SPEC.md for parity goals and architecture.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.PersistentFlags().String("data-dir", "", "LanguageTool data root (default: $LANG_DATA or ./inspiration/languagetool)")
	_ = viper.BindPFlag("data-dir", root.PersistentFlags().Lookup("data-dir"))
	_ = viper.BindEnv("data-dir", "LANG_DATA")

	root.AddCommand(newLintCmd())
	root.AddCommand(newLanguagesCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newVersionCmd())

	return root.Execute()
}

// dataDirFlag returns --data-dir if set, else LANG_DATA, else "" (engine uses default path).
func dataDirFlag(cmd *cobra.Command) string {
	if cmd.Flags().Changed("data-dir") {
		v, _ := cmd.Flags().GetString("data-dir")
		return v
	}
	// Persistent flag on root
	if pf := cmd.Root().PersistentFlags(); pf.Changed("data-dir") {
		v, _ := pf.GetString("data-dir")
		return v
	}
	if v := os.Getenv("LANG_DATA"); v != "" {
		return v
	}
	return ""
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "lang 0.0.0-dev")
		},
	}
}
