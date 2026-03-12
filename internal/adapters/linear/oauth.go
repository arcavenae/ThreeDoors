package linear

import (
	"context"
	"fmt"
	"os"
)

// DetectEnvToken checks for an existing LINEAR_API_KEY environment variable.
// Returns the token value and the variable name it was found in, or empty
// strings if not set.
func DetectEnvToken() (token, envVar string) {
	if v := os.Getenv("LINEAR_API_KEY"); v != "" {
		return v, "LINEAR_API_KEY"
	}
	return "", ""
}

// VerifyToken checks whether a Linear API key is valid by calling the
// authenticated viewer endpoint. Returns the user's display name on success.
// If baseURL is empty, the default Linear GraphQL endpoint is used.
func VerifyToken(ctx context.Context, token, baseURL string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("verify linear token: token must not be empty")
	}
	client := NewLinearClient(token)
	if baseURL != "" {
		client.baseURL = baseURL
	}
	viewer, err := client.QueryViewer(ctx)
	if err != nil {
		return "", fmt.Errorf("verify linear token: %w", err)
	}
	return viewer.Name, nil
}
