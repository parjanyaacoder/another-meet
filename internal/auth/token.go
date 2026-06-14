package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/parjanyaacoder/another-meet/internal/config"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const tokenFileName = "token.json"

// SaveToken persists the OAuth2 token to the config directory.
func SaveToken(token *oauth2.Token) error {
	if err := config.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	tokenPath := filepath.Join(config.Dir(), tokenFileName)
	f, err := os.OpenFile(tokenPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// LoadToken reads the stored OAuth2 token from the config directory.
func LoadToken() (*oauth2.Token, error) {
	tokenPath := filepath.Join(config.Dir(), tokenFileName)
	f, err := os.Open(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not authenticated — run 'another-meet auth login' first")
		}
		return nil, fmt.Errorf("failed to read token: %w", err)
	}
	defer f.Close()

	var token oauth2.Token
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &token, nil
}

// DeleteToken removes the stored token file.
func DeleteToken() error {
	tokenPath := filepath.Join(config.Dir(), tokenFileName)
	if err := os.Remove(tokenPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already gone
		}
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// TokenExists checks if a token file exists.
func TokenExists() bool {
	tokenPath := filepath.Join(config.Dir(), tokenFileName)
	_, err := os.Stat(tokenPath)
	return err == nil
}

// GetClient returns an authenticated HTTP client using the stored token.
// It automatically refreshes expired tokens.
func GetClient(ctx context.Context) (*oauth2.Token, *oauth2.Config, error) {
	token, err := LoadToken()
	if err != nil {
		return nil, nil, err
	}

	oauthConfig, err := OAuthConfig("http://localhost/callback")
	if err != nil {
		return nil, nil, err
	}

	// If the token needs refreshing, the TokenSource will handle it automatically
	tokenSource := oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to refresh token: %w — run 'another-meet auth login' again", err)
	}

	// Save the refreshed token if it changed
	if newToken.AccessToken != token.AccessToken {
		if err := SaveToken(newToken); err != nil {
			// Non-fatal: token still works for this session
			fmt.Fprintf(os.Stderr, "Warning: failed to save refreshed token: %v\n", err)
		}
	}

	return newToken, oauthConfig, nil
}

// NewCalendarService creates an authenticated Google Calendar API service.
func NewCalendarService(ctx context.Context) (*calendar.Service, error) {
	_, oauthConfig, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	token, err := LoadToken()
	if err != nil {
		return nil, err
	}

	client := oauthConfig.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Calendar service: %w", err)
	}

	return srv, nil
}

// GetUserEmail retrieves the current user's email using the userinfo API.
func GetUserEmail(ctx context.Context) (string, error) {
	token, oauthConfig, err := GetClient(ctx)
	if err != nil {
		return "", err
	}

	client := oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch user info: status %d", resp.StatusCode)
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("failed to decode user info: %w", err)
	}

	return userInfo.Email, nil
}
