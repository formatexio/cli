package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagAPIKey  string
	flagBaseURL string
)

var rootCmd = &cobra.Command{
	Use:   "formatex",
	Short: "FormaTeX CLI — compile LaTeX to PDF from the command line",
	Long: `FormaTeX CLI lets you compile LaTeX documents, watch for changes,
format source files, and convert between formats — all via the FormaTeX API.

Get started:
  formatex login          Save your API key
  formatex compile doc.tex  Compile a document to PDF`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "FormaTeX API key (overrides config and FORMATEX_API_KEY env var)")
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "API base URL (default: https://api.formatex.io)")
}
