package github

import (
	"context"
	"fmt"
	"os"

	"github.com/arcavenae/ThreeDoors/internal/core/connection/oauth"
)

// GitHub OAuth device code flow endpoints per
// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#device-flow
const (
	GitHubDeviceCodeEndpoint = "https://github.com/login/device/code"
	GitHubTokenEndpoint      = "https://github.com/login/oauth/access_token"
)

// GitHubOAuthScopes are the minimum scopes required for issue access.
var GitHubOAuthScopes = []string{"repo"}

// GitHubOAuthConfig returns a DeviceCodeConfig for GitHub's OAuth device code flow.
// clientID must be a registered GitHub OAuth App client ID.
func GitHubOAuthConfig(clientID string) oauth.DeviceCodeConfig {
	return oauth.DeviceCodeConfig{
		AuthEndpoint:  GitHubDeviceCodeEndpoint,
		TokenEndpoint: GitHubTokenEndpoint,
		ClientID:      clientID,
		Scopes:        GitHubOAuthScopes,
	}
}

// DetectEnvToken checks for existing GH_TOKEN or GITHUB_TOKEN environment
// variables. Returns the token value and the variable name it was found in,
// or empty strings if neither is set.
func DetectEnvToken() (token, envVar string) {
	if v := os.Getenv("GH_TOKEN"); v != "" {
		return v, "GH_TOKEN"
	}
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		return v, "GITHUB_TOKEN"
	}
	return "", ""
}

// VerifyToken checks whether a GitHub token is valid by calling the
// authenticated user endpoint. Returns the GitHub login name on success.
func VerifyToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("verify github token: token must not be empty")
	}
	client := NewGitHubClient(&GitHubConfig{Token: token})
	login, err := client.GetAuthenticatedUser(ctx)
	if err != nil {
		return "", fmt.Errorf("verify github token: %w", err)
	}
	return login, nil
}
