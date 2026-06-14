package auth

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

//go:embed credentials.json
var credentialsJSON []byte

type installedCredentials struct {
	Installed struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		ProjectID    string   `json:"project_id"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
		RedirectURIs []string `json:"redirect_uris"`
	} `json:"installed"`
}

// OAuthConfig returns the OAuth2 config for the application.
// Priority: environment variables > embedded credentials.json
func OAuthConfig(redirectURL string) (*oauth2.Config, error) {
	clientID := os.Getenv("ANOTHER_MEET_CLIENT_ID")
	clientSecret := os.Getenv("ANOTHER_MEET_CLIENT_SECRET")

	if clientID != "" && clientSecret != "" {
		return &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes: []string{
				calendar.CalendarEventsScope,
				"https://www.googleapis.com/auth/userinfo.email",
				"openid",
			},
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURL,
		}, nil
	}

	// Try embedded credentials
	var creds installedCredentials
	if err := json.Unmarshal(credentialsJSON, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	// Check if this is the placeholder file
	if strings.HasPrefix(creds.Installed.ClientID, "REPLACE_") || creds.Installed.ClientID == "" {
		return nil, fmt.Errorf(
			"OAuth credentials not configured.\n\n" +
				"  Option 1: Set environment variables:\n" +
				"    export ANOTHER_MEET_CLIENT_ID=\"your-client-id\"\n" +
				"    export ANOTHER_MEET_CLIENT_SECRET=\"your-client-secret\"\n\n" +
				"  Option 2: Copy your GCP credentials file:\n" +
				"    cp /path/to/client_secret_*.json internal/auth/credentials.json\n" +
				"    go build -o another-meet ./main.go\n\n" +
				"  See README.md for setup instructions.")
	}

	return &oauth2.Config{
		ClientID:     creds.Installed.ClientID,
		ClientSecret: creds.Installed.ClientSecret,
		Scopes: []string{
			calendar.CalendarEventsScope,
			"https://www.googleapis.com/auth/userinfo.email",
			"openid",
		},
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
	}, nil
}
