package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

// LoginWithBrowser performs the OAuth2 Authorization Code flow with PKCE.
// It starts a local HTTP server, opens the browser for consent, and returns the token.
func LoginWithBrowser(ctx context.Context) (*oauth2.Token, error) {
	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)

	config, err := OAuthConfig(redirectURL)
	if err != nil {
		listener.Close()
		return nil, err
	}

	// Generate PKCE code verifier and challenge
	verifier, challenge, err := generatePKCE()
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to generate PKCE: %w", err)
	}

	// Generate state parameter for CSRF protection
	state, err := generateRandomString(32)
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Build the authorization URL
	authURL := config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	// Channel to receive the authorization code
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Set up the callback handler
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch: possible CSRF attack")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errCh <- fmt.Errorf("authorization error: %s", errParam)
			http.Error(w, "Authorization failed", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>another-meet</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; 
         display: flex; align-items: center; justify-content: center; height: 100vh; 
         margin: 0; background: #0f172a; color: #e2e8f0; }
  .card { text-align: center; padding: 48px; }
  h1 { color: #34d399; font-size: 24px; margin-bottom: 8px; }
  p { color: #94a3b8; font-size: 16px; }
</style></head>
<body><div class="card">
  <h1>✓ Authentication successful!</h1>
  <p>You can close this tab and return to your terminal.</p>
</div></body></html>`)

		codeCh <- code
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Open the browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Open this URL in your browser:\n\n  %s\n\n", authURL)
	}

	// Wait for the authorization code
	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		server.Close()
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Close()
		return nil, fmt.Errorf("authorization timed out after 5 minutes")
	case <-ctx.Done():
		server.Close()
		return nil, ctx.Err()
	}

	// Shut down the server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)

	// Exchange the code for a token (with PKCE verifier)
	token, err := config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	return token, nil
}

// LoginHeadless performs the device authorization grant flow for headless environments.
func LoginHeadless(ctx context.Context) (*oauth2.Token, error) {
	config, err := OAuthConfig("")
	if err != nil {
		return nil, err
	}

	// Request device code
	deviceResp, err := requestDeviceCode(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}

	fmt.Printf("\n  Visit:  %s\n", deviceResp.VerificationURL)
	fmt.Printf("  Code:   %s\n\n", deviceResp.UserCode)

	// Poll for token
	return pollForToken(ctx, config, deviceResp)
}

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func requestDeviceCode(ctx context.Context, config *oauth2.Config) (*deviceCodeResponse, error) {
	resp, err := http.PostForm("https://oauth2.googleapis.com/device/code", map[string][]string{
		"client_id": {config.ClientID},
		"scope":     {config.Scopes[0]},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result deviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func pollForToken(ctx context.Context, config *oauth2.Config, device *deviceCodeResponse) (*oauth2.Token, error) {
	interval := time.Duration(device.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("device authorization expired")
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		resp, err := http.PostForm("https://oauth2.googleapis.com/token", map[string][]string{
			"client_id":     {config.ClientID},
			"client_secret": {config.ClientSecret},
			"device_code":   {device.DeviceCode},
			"grant_type":    {"urn:ietf:params:oauth:grant-type:device_code"},
		})
		if err != nil {
			continue
		}

		var result struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			TokenType    string `json:"token_type"`
			ExpiresIn    int    `json:"expires_in"`
			Error        string `json:"error"`
		}

		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		switch result.Error {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5 * time.Second
			continue
		case "":
			return &oauth2.Token{
				AccessToken:  result.AccessToken,
				RefreshToken: result.RefreshToken,
				TokenType:    result.TokenType,
				Expiry:       time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
			}, nil
		default:
			return nil, fmt.Errorf("authorization failed: %s", result.Error)
		}
	}
}

// generatePKCE creates a PKCE code verifier and S256 challenge
func generatePKCE() (verifier, challenge string, err error) {
	verifierBytes := make([]byte, 32)
	if _, err = rand.Read(verifierBytes); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])

	return verifier, challenge, nil
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}
