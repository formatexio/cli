package cmd

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/formatexio/cli/internal/api"
	"github.com/formatexio/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	compileEngine   string
	compileOutput   string
	compileRuns     int
	compileTimeout  int
	compileMainFile string
	compileOpen     bool
)

var compileCmd = &cobra.Command{
	Use:   "compile <file.tex|project.zip>",
	Short: "Compile a LaTeX document to PDF",
	Long: `Compile a .tex file or a .zip project archive to PDF using the FormaTeX API.

The output PDF is saved alongside the source file by default,
or to the path specified with --output.

Examples:
  formatex compile thesis.tex
  formatex compile thesis.tex --engine xelatex --output thesis.pdf
  formatex compile project.zip --main main.tex`,
	Args: cobra.ExactArgs(1),
	RunE: runCompile,
}

func init() {
	compileCmd.Flags().StringVarP(&compileEngine, "engine", "e", "pdflatex", "LaTeX engine: pdflatex, xelatex, lualatex, latexmk")
	compileCmd.Flags().StringVarP(&compileOutput, "output", "o", "", "Output PDF path (default: <input>.pdf)")
	compileCmd.Flags().IntVar(&compileRuns, "runs", 0, "Number of compiler passes (1–5)")
	compileCmd.Flags().IntVar(&compileTimeout, "timeout", 0, "Max compile time in seconds")
	compileCmd.Flags().StringVar(&compileMainFile, "main", "", "Entry-point .tex file inside a ZIP archive")
	compileCmd.Flags().BoolVar(&compileOpen, "open", false, "Open the PDF after successful compilation")
	rootCmd.AddCommand(compileCmd)
}

func runCompile(cmd *cobra.Command, args []string) error {
	src := args[0]
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("file not found: %s", src)
	}

	apiKey, err := config.ResolveAPIKey(flagAPIKey)
	if err != nil {
		return err
	}
	client := api.New(apiKey, config.ResolveBaseURL(flagBaseURL))

	outPath := compileOutput
	if outPath == "" {
		ext := filepath.Ext(src)
		outPath = strings.TrimSuffix(src, ext) + ".pdf"
	}

	fmt.Printf("Compiling %s...\n", filepath.Base(src))
	start := time.Now()

	var result *api.CompileResult

	switch strings.ToLower(filepath.Ext(src)) {
	case ".zip":
		zipData, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read ZIP: %w", err)
		}
		result, err = client.CompileZip(zipData, compileEngine, compileMainFile, compileRuns, compileTimeout)
		if err != nil {
			return handleCompileError(err)
		}
	case ".tex":
		latex, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		result, err = client.Compile(api.CompileRequest{
			LaTeX:   string(latex),
			Engine:  compileEngine,
			Runs:    compileRuns,
			Timeout: compileTimeout,
		})
		if err != nil {
			return handleCompileError(err)
		}
	default:
		// Try wrapping the directory as a ZIP if it's a directory
		info, _ := os.Stat(src)
		if info != nil && info.IsDir() {
			zipData, err := zipDirectory(src)
			if err != nil {
				return fmt.Errorf("failed to zip directory: %w", err)
			}
			result, err = client.CompileZip(zipData, compileEngine, compileMainFile, compileRuns, compileTimeout)
			if err != nil {
				return handleCompileError(err)
			}
		} else {
			return fmt.Errorf("unsupported file type: %s (use .tex or .zip)", filepath.Ext(src))
		}
	}

	if err := os.WriteFile(outPath, result.PDF, 0644); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	elapsed := time.Since(start).Round(time.Millisecond)
	fmt.Printf("Done in %s — %s (%.1f KB, engine: %s)\n",
		elapsed,
		outPath,
		float64(len(result.PDF))/1024,
		result.Engine,
	)

	if compileOpen {
		openFile(outPath)
	}

	return nil
}

func handleCompileError(err error) error {
	var cf *api.CompileFailure
	if errors.As(err, &cf) {
		fmt.Fprintln(os.Stderr, "Compilation failed.")
		if len(cf.Errors) > 0 {
			fmt.Fprintln(os.Stderr, "\nErrors:")
			for _, e := range cf.Errors {
				loc := ""
				if e.File != "" {
					loc = e.File
					if e.Line > 0 {
						loc += fmt.Sprintf(":%d", e.Line)
					}
					loc += ": "
				}
				fmt.Fprintf(os.Stderr, "  %s%s\n", loc, e.Message)
			}
		}
		if cf.AIExplanation != "" {
			fmt.Fprintf(os.Stderr, "\nAI: %s\n", cf.AIExplanation)
		}
		if len(cf.Errors) == 0 && cf.Log != "" {
			// Print last 20 lines of log
			lines := strings.Split(cf.Log, "\n")
			if len(lines) > 20 {
				lines = lines[len(lines)-20:]
			}
			fmt.Fprintln(os.Stderr, "\nLog (last 20 lines):")
			for _, l := range lines {
				fmt.Fprintln(os.Stderr, "  "+l)
			}
		}
		return fmt.Errorf("compilation failed")
	}
	return err
}

func zipDirectory(dir string) ([]byte, error) {
	var buf strings.Builder
	_ = buf

	tmpFile, err := os.CreateTemp("", "formatex-*.zip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	zw := zip.NewWriter(tmpFile)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	})
	if err != nil {
		return nil, err
	}
	zw.Close()
	tmpFile.Close()

	return os.ReadFile(tmpFile.Name())
}
