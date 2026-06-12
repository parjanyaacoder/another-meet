package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
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

		fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Authentication Successful | another-meet</title>
  <link href="https://fonts.googleapis.com/css2?family=Product+Sans:wght@400;500;700&family=Inter:wght@400;500&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-color: #0f172a;
      --card-bg: #1e293b;
      --border-color: #334155;
      --text-main: #f8fafc;
      --text-muted: #94a3b8;
      --accent: #10b981;
      --accent-hover: #059669;
    }
    
    body { 
      font-family: 'Inter', -apple-system, sans-serif; 
      display: flex; 
      align-items: center; 
      justify-content: center; 
      min-height: 100vh; 
      margin: 0; 
      background: var(--bg-color); 
      color: var(--text-main);
    }

    .container {
      text-align: center;
      padding: 48px 40px;
      max-width: 480px;
      background: var(--card-bg);
      border: 1px solid var(--border-color);
      border-radius: 16px;
      box-shadow: 0 20px 25px -5px rgba(0,0,0,0.2), 0 10px 10px -5px rgba(0,0,0,0.1);
    }

    .brand {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 12px;
      margin-bottom: 8px;
    }

    .logo-icon {
      width: 32px;
      height: 32px;
    }

    .logo-text {
      font-family: 'Product Sans', 'Inter', sans-serif;
      font-weight: 700;
      font-size: 26px;
      letter-spacing: -0.5px;
      color: var(--text-main);
    }

    .tagline {
      font-size: 15px;
      color: var(--text-muted);
      margin: 0 0 36px 0;
      line-height: 1.5;
    }

    .divider {
      height: 1px;
      background: var(--border-color);
      margin: 0 0 36px 0;
    }

    h1 {
      font-size: 22px;
      font-weight: 500;
      margin: 0 0 32px 0;
      color: var(--accent);
    }

    .close-btn {
      background-color: var(--accent);
      color: #ffffff;
      border: none;
      padding: 12px 28px;
      font-size: 15px;
      font-weight: 500;
      border-radius: 6px;
      cursor: pointer;
      font-family: 'Inter', sans-serif;
      transition: background-color 0.2s;
    }

    .close-btn:hover {
      background-color: var(--accent-hover);
    }

    .fallback-text {
      display: none;
      font-size: 13px;
      color: var(--text-muted);
      margin-top: 16px;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="brand">
      <svg class="logo-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M16 16L22 20V4L16 8V16Z" fill="#34A853"/>
        <path d="M14 6H4C2.9 6 2 6.9 2 8V16C2 17.1 2.9 18 4 18H14C15.1 18 16 17.1 16 16V8C16 6.9 15.1 6 14 6Z" fill="#4285F4"/>
      </svg>
      <div class="logo-text">another-meet</div>
    </div>
    <p class="tagline">Manage Google Meet meetings from your terminal.</p>
    
    <div class="divider"></div>
    
    <h1>✓ Authentication Successful</h1>
    
    <button class="close-btn" onclick="closeWindow()">Close this window</button>
    <p class="fallback-text" id="fallback">You can now safely close this tab.</p>
  </div>

  <script>
    function closeWindow() {
      window.close();
      setTimeout(() => {
        document.getElementById('fallback').style.display = 'block';
      }, 300);
    }
  </script>
</body>
</html>`)

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
