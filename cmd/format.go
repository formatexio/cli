package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/formatexio/cli/internal/api"
	"github.com/formatexio/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	formatOutput  string
	formatInPlace bool
)

var formatCmd = &cobra.Command{
	Use:   "format <file.tex>",
	Short: "Format a LaTeX source file using latexindent",
	Long: `Format a LaTeX source file and print the result to stdout,
or write it back in-place with --write.

Examples:
  formatex format main.tex              # print to stdout
  formatex format main.tex --write      # overwrite the file
  formatex format main.tex -o out.tex   # write to a new file`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		if strings.ToLower(filepath.Ext(src)) != ".tex" {
			return fmt.Errorf("format only supports .tex files")
		}
		latex, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		apiKey, err := config.ResolveAPIKey(flagAPIKey)
		if err != nil {
			return err
		}
		client := api.New(apiKey, config.ResolveBaseURL(flagBaseURL))

		fmt.Fprintf(os.Stderr, "Formatting %s...\n", filepath.Base(src))
		formatted, durMs, err := client.Format(string(latex))
		if err != nil {
			return fmt.Errorf("formatting failed: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Done in %dms\n", durMs)

		switch {
		case formatInPlace:
			if err := os.WriteFile(src, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Written to %s\n", src)
		case formatOutput != "":
			if err := os.WriteFile(formatOutput, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Written to %s\n", formatOutput)
		default:
			fmt.Print(formatted)
		}
		return nil
	},
}

func init() {
	formatCmd.Flags().StringVarP(&formatOutput, "output", "o", "", "Write formatted output to this file")
	formatCmd.Flags().BoolVarP(&formatInPlace, "write", "w", false, "Write formatted output back to the source file")
	rootCmd.AddCommand(formatCmd)
}
