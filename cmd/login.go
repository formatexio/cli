package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/formatexio/cli/internal/api"
	"github.com/formatexio/cli/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save your FormaTeX API key",
	Long: `Save your API key to ~/.config/formatex/config.json.

You can also set FORMATEX_API_KEY as an environment variable,
or pass --api-key to any command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Enter your FormaTeX API key: ")
		reader := bufio.NewReader(os.Stdin)
		key, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		baseURL := config.ResolveBaseURL(flagBaseURL)
		client := api.New(key, baseURL)
		fmt.Print("Verifying API key... ")
		info, err := client.WhoAmI()
		if err != nil {
			fmt.Println("failed")
			return fmt.Errorf("invalid API key: %w", err)
		}
		fmt.Println("ok")

		cfg := &config.Config{APIKey: key}
		if flagBaseURL != "" {
			cfg.BaseURL = flagBaseURL
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if email, ok := info["email"].(string); ok && email != "" {
			fmt.Printf("Logged in as %s\n", email)
		} else {
			fmt.Println("API key saved.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
