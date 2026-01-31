package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// CallbackResult contains the result of an OAuth callback.
type CallbackResult struct {
	Code  string
	Error string
}

// StartCallbackServer starts a local HTTP server to receive the OAuth callback.
// Returns a channel that will receive the authorization code or error.
func StartCallbackServer(ctx context.Context, port int) (<-chan CallbackResult, error) {
	resultChan := make(chan CallbackResult, 1)

	// Check if port is available
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server on port %d: %w", port, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errMsg := r.URL.Query().Get("error")
		errDesc := r.URL.Query().Get("error_description")

		if errMsg != "" {
			msg := errMsg
			if errDesc != "" {
				msg = fmt.Sprintf("%s: %s", errMsg, errDesc)
			}
			resultChan <- CallbackResult{Error: msg}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Failed</title></head>
<body>
<h1>Authentication Failed</h1>
<p>Error: %s</p>
<p>You can close this window.</p>
</body>
</html>`, msg)
			return
		}

		if code == "" {
			resultChan <- CallbackResult{Error: "no authorization code received"}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Failed</title></head>
<body>
<h1>Authentication Failed</h1>
<p>No authorization code received.</p>
<p>You can close this window.</p>
</body>
</html>`)
			return
		}

		resultChan <- CallbackResult{Code: code}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Successful</title></head>
<body>
<h1>Authentication Successful!</h1>
<p>You can close this window and return to the terminal.</p>
</body>
</html>`)
	})

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			resultChan <- CallbackResult{Error: fmt.Sprintf("callback server error: %v", err)}
		}
	}()

	// Shutdown server when context is cancelled or result is received
	go func() {
		select {
		case <-ctx.Done():
		case <-resultChan:
		}
		// Give a moment for the response to be sent
		time.Sleep(100 * time.Millisecond)
		_ = server.Close()
	}()

	return resultChan, nil
}
