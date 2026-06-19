package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/formatexio/cli/internal/api"
	"github.com/formatexio/cli/internal/config"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	watchEngine  string
	watchOutput  string
	watchRuns    int
	watchTimeout int
	watchOpen    bool
)

var watchCmd = &cobra.Command{
	Use:   "watch <file.tex>",
	Short: "Watch a .tex file and recompile on save",
	Long: `Watch a LaTeX source file for changes and automatically recompile it
to PDF whenever it is saved.

Press Ctrl+C to stop watching.

Example:
  formatex watch thesis.tex --engine xelatex`,
	Args: cobra.ExactArgs(1),
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringVarP(&watchEngine, "engine", "e", "pdflatex", "LaTeX engine: pdflatex, xelatex, lualatex, latexmk")
	watchCmd.Flags().StringVarP(&watchOutput, "output", "o", "", "Output PDF path (default: <input>.pdf)")
	watchCmd.Flags().IntVar(&watchRuns, "runs", 0, "Number of compiler passes (1–5)")
	watchCmd.Flags().IntVar(&watchTimeout, "timeout", 0, "Max compile time in seconds")
	watchCmd.Flags().BoolVar(&watchOpen, "open", false, "Open the PDF on first successful compilation")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	src := args[0]
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("file not found: %s", src)
	}
	if strings.ToLower(filepath.Ext(src)) != ".tex" {
		return fmt.Errorf("watch only supports .tex files")
	}

	apiKey, err := config.ResolveAPIKey(flagAPIKey)
	if err != nil {
		return err
	}
	client := api.New(apiKey, config.ResolveBaseURL(flagBaseURL))

	outPath := watchOutput
	if outPath == "" {
		outPath = strings.TrimSuffix(src, filepath.Ext(src)) + ".pdf"
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(src); err != nil {
		return fmt.Errorf("failed to watch %s: %w", src, err)
	}

	absSrc, _ := filepath.Abs(src)
	fmt.Printf("Watching %s (engine: %s)\n", absSrc, watchEngine)
	fmt.Println("Press Ctrl+C to stop.")

	// Compile immediately on start.
	compileOnce(client, src, outPath, watchEngine, watchRuns, watchTimeout, watchOpen)
	openedOnce := true
	_ = openedOnce

	// Debounce: ignore rapid successive writes (editor save + swap file).
	var lastEvent time.Time
	const debounce = 300 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if time.Since(lastEvent) < debounce {
					continue
				}
				lastEvent = time.Now()
				compileOnce(client, src, outPath, watchEngine, watchRuns, watchTimeout, false)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
		}
	}
}

func compileOnce(client *api.Client, src, outPath, engine string, runs, timeout int, open bool) {
	fmt.Printf("[%s] Compiling...", time.Now().Format("15:04:05"))
	latex, err := os.ReadFile(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, " failed to read file: %v\n", err)
		return
	}
	start := time.Now()
	result, err := client.Compile(api.CompileRequest{
		LaTeX:   string(latex),
		Engine:  engine,
		Runs:    runs,
		Timeout: timeout,
	})
	if err != nil {
		fmt.Println(" FAILED")
		if err := handleCompileError(err); err != nil {
			// error already printed by handleCompileError
		}
		return
	}
	if err := os.WriteFile(outPath, result.PDF, 0644); err != nil {
		fmt.Fprintf(os.Stderr, " failed to write PDF: %v\n", err)
		return
	}
	elapsed := time.Since(start).Round(time.Millisecond)
	fmt.Printf(" ok (%s, %.1f KB)\n", elapsed, float64(len(result.PDF))/1024)
	if open {
		openFile(outPath)
	}
}
