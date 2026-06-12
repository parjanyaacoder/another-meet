package cmd

import (
	"context"
	"fmt"

	"github.com/parjanyaacoder/another-meet/internal/auth"
	"github.com/parjanyaacoder/another-meet/internal/ui"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Google",
	Long:  `Manage authentication with Google for accessing Calendar and Meet.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with your Google account",
	Long: `Log in to your Google account to access Calendar and Meet.
By default, opens your browser for authentication.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if auth.TokenExists() {
			ui.Warn("Already authenticated. Use 'another-meet auth logout' first to re-authenticate.")
			return nil
		}

		ui.Info("Opening browser for Google authorization...")
		t, err := auth.LoginWithBrowser(ctx)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		if err := auth.SaveToken(t); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		// Try to get user email
		email, err := auth.GetUserEmail(ctx)
		if err != nil {
			ui.Success("Authenticated successfully!")
		} else {
			ui.Success("Authenticated as %s", email)
		}

		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !auth.TokenExists() {
			if ui.IsJSON() {
				return ui.PrintJSON(map[string]interface{}{
					"authenticated": false,
					"error":         "not_authenticated",
				})
			}
			ui.Error("Not authenticated — run 'another-meet auth login' to get started.")
			return nil
		}

		ctx := context.Background()
		email, err := auth.GetUserEmail(ctx)
		if err != nil {
			if ui.IsJSON() {
				return ui.PrintJSON(map[string]interface{}{
					"authenticated": true,
					"email":         "",
					"error":         err.Error(),
				})
			}
			ui.Warn("Token found but may be invalid: %v", err)
			ui.Info("Run 'another-meet auth login' to re-authenticate.")
			return nil
		}

		if ui.IsJSON() {
			return ui.PrintJSON(map[string]interface{}{
				"authenticated": true,
				"email":         email,
			})
		}

		ui.Success("Authenticated as %s", email)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !auth.TokenExists() {
			ui.Info("No authentication found.")
			return nil
		}

		if err := auth.DeleteToken(); err != nil {
			return fmt.Errorf("failed to remove authentication: %w", err)
		}

		ui.Success("Logged out successfully.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
