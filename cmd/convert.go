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

var convertFormat string

var convertCmd = &cobra.Command{
	Use:   "convert <file.tex>",
	Short: "Convert a LaTeX document to another format",
	Long: `Convert a LaTeX document to DOCX, HTML, ODT, EPUB, Markdown, or plain text.

The output file is saved alongside the source file with the appropriate
extension, or to the path specified with --output.

Examples:
  formatex convert thesis.tex                    # → thesis.docx (default)
  formatex convert thesis.tex --format html      # → thesis.html
  formatex convert thesis.tex -f odt -o out.odt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		if strings.ToLower(filepath.Ext(src)) != ".tex" {
			return fmt.Errorf("convert only supports .tex files")
		}
		latex, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		validFormats := map[string]string{
			"docx":     "docx",
			"html":     "html",
			"odt":      "odt",
			"epub":     "epub",
			"markdown": "md",
			"md":       "md",
			"txt":      "txt",
		}
		ext, ok := validFormats[strings.ToLower(convertFormat)]
		if !ok {
			return fmt.Errorf("unsupported format %q — choose from: docx, html, odt, epub, markdown, txt", convertFormat)
		}

		outPath := compileOutput
		if outPath == "" {
			base := strings.TrimSuffix(src, filepath.Ext(src))
			outPath = base + "." + ext
		}

		apiKey, err := config.ResolveAPIKey(flagAPIKey)
		if err != nil {
			return err
		}
		client := api.New(apiKey, config.ResolveBaseURL(flagBaseURL))

		fmt.Fprintf(os.Stderr, "Converting %s to %s...\n", filepath.Base(src), strings.ToUpper(convertFormat))
		data, err := client.Convert(string(latex), convertFormat)
		if err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}

		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Saved to %s (%.1f KB)\n", outPath, float64(len(data))/1024)
		return nil
	},
}

func init() {
	convertCmd.Flags().StringVarP(&convertFormat, "format", "f", "docx", "Output format: docx, html, odt, epub, markdown, txt")
	convertCmd.Flags().StringVarP(&compileOutput, "output", "o", "", "Output file path")
	rootCmd.AddCommand(convertCmd)
}
