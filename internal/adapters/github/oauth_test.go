package github

import (
	"context"
	"testing"
)

func TestGitHubOAuthConfig(t *testing.T) {
	t.Parallel()

	config := GitHubOAuthConfig("test-client-id")

	if config.AuthEndpoint != GitHubDeviceCodeEndpoint {
		t.Errorf("AuthEndpoint = %q, want %q", config.AuthEndpoint, GitHubDeviceCodeEndpoint)
	}
	if config.TokenEndpoint != GitHubTokenEndpoint {
		t.Errorf("TokenEndpoint = %q, want %q", config.TokenEndpoint, GitHubTokenEndpoint)
	}
	if config.ClientID != "test-client-id" {
		t.Errorf("ClientID = %q, want %q", config.ClientID, "test-client-id")
	}
	if len(config.Scopes) != 1 || config.Scopes[0] != "repo" {
		t.Errorf("Scopes = %v, want [repo]", config.Scopes)
	}
}

func TestGitHubOAuthConfig_EndpointValues(t *testing.T) {
	t.Parallel()

	if GitHubDeviceCodeEndpoint != "https://github.com/login/device/code" {
		t.Errorf("GitHubDeviceCodeEndpoint = %q", GitHubDeviceCodeEndpoint)
	}
	if GitHubTokenEndpoint != "https://github.com/login/oauth/access_token" {
		t.Errorf("GitHubTokenEndpoint = %q", GitHubTokenEndpoint)
	}
}

func TestDetectEnvToken_GHToken(t *testing.T) {
	t.Setenv("GH_TOKEN", "gho_abc123")
	t.Setenv("GITHUB_TOKEN", "")

	token, envVar := DetectEnvToken()
	if token != "gho_abc123" {
		t.Errorf("token = %q, want %q", token, "gho_abc123")
	}
	if envVar != "GH_TOKEN" {
		t.Errorf("envVar = %q, want %q", envVar, "GH_TOKEN")
	}
}

func TestDetectEnvToken_GitHubToken(t *testing.T) {
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "ghp_def456")

	token, envVar := DetectEnvToken()
	if token != "ghp_def456" {
		t.Errorf("token = %q, want %q", token, "ghp_def456")
	}
	if envVar != "GITHUB_TOKEN" {
		t.Errorf("envVar = %q, want %q", envVar, "GITHUB_TOKEN")
	}
}

func TestDetectEnvToken_GHTokenPrecedence(t *testing.T) {
	t.Setenv("GH_TOKEN", "gho_first")
	t.Setenv("GITHUB_TOKEN", "ghp_second")

	token, envVar := DetectEnvToken()
	if token != "gho_first" {
		t.Errorf("token = %q, want %q (GH_TOKEN should take precedence)", token, "gho_first")
	}
	if envVar != "GH_TOKEN" {
		t.Errorf("envVar = %q, want %q", envVar, "GH_TOKEN")
	}
}

func TestDetectEnvToken_NeitherSet(t *testing.T) {
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	token, envVar := DetectEnvToken()
	if token != "" {
		t.Errorf("token = %q, want empty", token)
	}
	if envVar != "" {
		t.Errorf("envVar = %q, want empty", envVar)
	}
}

func TestVerifyToken_EmptyToken(t *testing.T) {
	t.Parallel()

	_, err := VerifyToken(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
	if !containsStr(err.Error(), "token must not be empty") {
		t.Errorf("error %q should contain 'token must not be empty'", err.Error())
	}
}

func TestVerifyToken_CancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := VerifyToken(ctx, "some-token")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestGitHubOAuthScopes(t *testing.T) {
	t.Parallel()

	if len(GitHubOAuthScopes) != 1 {
		t.Fatalf("GitHubOAuthScopes length = %d, want 1", len(GitHubOAuthScopes))
	}
	if GitHubOAuthScopes[0] != "repo" {
		t.Errorf("GitHubOAuthScopes[0] = %q, want %q", GitHubOAuthScopes[0], "repo")
	}
}
