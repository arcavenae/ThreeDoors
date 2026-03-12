package oauth

import (
	"testing"
	"time"
)

func TestTokenMetadata_NeedsRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "zero expiry never needs refresh",
			expiresAt: time.Time{},
			want:      false,
		},
		{
			name:      "far future does not need refresh",
			expiresAt: time.Now().UTC().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "within 5 minutes needs refresh",
			expiresAt: time.Now().UTC().Add(3 * time.Minute),
			want:      true,
		},
		{
			name:      "exactly at threshold needs refresh",
			expiresAt: time.Now().UTC().Add(RefreshThreshold),
			want:      true,
		},
		{
			name:      "already expired needs refresh",
			expiresAt: time.Now().UTC().Add(-1 * time.Minute),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			meta := &TokenMetadata{ExpiresAt: tt.expiresAt}
			if got := meta.NeedsRefresh(); got != tt.want {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenMetadata_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "zero expiry not expired",
			expiresAt: time.Time{},
			want:      false,
		},
		{
			name:      "future not expired",
			expiresAt: time.Now().UTC().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "past is expired",
			expiresAt: time.Now().UTC().Add(-1 * time.Minute),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			meta := &TokenMetadata{ExpiresAt: tt.expiresAt}
			if got := meta.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenMetadata_CanRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		meta TokenMetadata
		want bool
	}{
		{
			name: "all fields present",
			meta: TokenMetadata{
				RefreshToken:  "rt_123",
				TokenEndpoint: "https://example.com/token",
				ClientID:      "client-id",
			},
			want: true,
		},
		{
			name: "missing refresh token",
			meta: TokenMetadata{
				TokenEndpoint: "https://example.com/token",
				ClientID:      "client-id",
			},
			want: false,
		},
		{
			name: "missing token endpoint",
			meta: TokenMetadata{
				RefreshToken: "rt_123",
				ClientID:     "client-id",
			},
			want: false,
		},
		{
			name: "missing client ID",
			meta: TokenMetadata{
				RefreshToken:  "rt_123",
				TokenEndpoint: "https://example.com/token",
			},
			want: false,
		},
		{
			name: "empty metadata",
			meta: TokenMetadata{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.meta.CanRefresh(); got != tt.want {
				t.Errorf("CanRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalUnmarshalCredential(t *testing.T) {
	t.Parallel()

	t.Run("round trip OAuth metadata", func(t *testing.T) {
		t.Parallel()
		original := &TokenMetadata{
			AccessToken:   "at_123",
			RefreshToken:  "rt_456",
			ExpiresAt:     time.Date(2026, 3, 12, 15, 0, 0, 0, time.UTC),
			TokenEndpoint: "https://example.com/token",
			ClientID:      "my-client",
		}

		serialized, err := MarshalCredential(original)
		if err != nil {
			t.Fatalf("MarshalCredential: %v", err)
		}

		got := UnmarshalCredential(serialized)
		if got.AccessToken != original.AccessToken {
			t.Errorf("AccessToken = %q, want %q", got.AccessToken, original.AccessToken)
		}
		if got.RefreshToken != original.RefreshToken {
			t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, original.RefreshToken)
		}
		if !got.ExpiresAt.Equal(original.ExpiresAt) {
			t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, original.ExpiresAt)
		}
		if got.TokenEndpoint != original.TokenEndpoint {
			t.Errorf("TokenEndpoint = %q, want %q", got.TokenEndpoint, original.TokenEndpoint)
		}
		if got.ClientID != original.ClientID {
			t.Errorf("ClientID = %q, want %q", got.ClientID, original.ClientID)
		}
	})

	t.Run("plain API key", func(t *testing.T) {
		t.Parallel()
		got := UnmarshalCredential("sk_my_api_key_123")
		if got.AccessToken != "sk_my_api_key_123" {
			t.Errorf("AccessToken = %q, want %q", got.AccessToken, "sk_my_api_key_123")
		}
		if got.RefreshToken != "" {
			t.Errorf("RefreshToken = %q, want empty", got.RefreshToken)
		}
		if !got.CanRefresh() == false {
			t.Error("plain API key should not be refreshable")
		}
	})

	t.Run("empty credential", func(t *testing.T) {
		t.Parallel()
		got := UnmarshalCredential("")
		if got.AccessToken != "" {
			t.Errorf("AccessToken = %q, want empty", got.AccessToken)
		}
	})

	t.Run("empty JSON object", func(t *testing.T) {
		t.Parallel()
		got := UnmarshalCredential("{}")
		if got.AccessToken != "" {
			t.Errorf("AccessToken = %q, want empty for {}", got.AccessToken)
		}
	})
}

func TestNewTokenMetadataFromResponse(t *testing.T) {
	t.Parallel()

	t.Run("with expiry", func(t *testing.T) {
		t.Parallel()
		before := time.Now().UTC()
		resp := &TokenResponse{
			AccessToken:  "at_new",
			RefreshToken: "rt_new",
			ExpiresIn:    3600,
			TokenType:    "bearer",
		}

		meta := NewTokenMetadataFromResponse(resp, "https://example.com/token", "client-123")

		if meta.AccessToken != "at_new" {
			t.Errorf("AccessToken = %q, want %q", meta.AccessToken, "at_new")
		}
		if meta.RefreshToken != "rt_new" {
			t.Errorf("RefreshToken = %q, want %q", meta.RefreshToken, "rt_new")
		}
		if meta.TokenEndpoint != "https://example.com/token" {
			t.Errorf("TokenEndpoint = %q, want %q", meta.TokenEndpoint, "https://example.com/token")
		}
		if meta.ClientID != "client-123" {
			t.Errorf("ClientID = %q, want %q", meta.ClientID, "client-123")
		}
		expectedExpiry := before.Add(3600 * time.Second)
		if meta.ExpiresAt.Before(expectedExpiry.Add(-2 * time.Second)) {
			t.Errorf("ExpiresAt = %v, expected near %v", meta.ExpiresAt, expectedExpiry)
		}
	})

	t.Run("without expiry", func(t *testing.T) {
		t.Parallel()
		resp := &TokenResponse{
			AccessToken: "at_noexpiry",
			ExpiresIn:   0,
		}

		meta := NewTokenMetadataFromResponse(resp, "https://example.com/token", "client-456")

		if !meta.ExpiresAt.IsZero() {
			t.Errorf("ExpiresAt = %v, want zero", meta.ExpiresAt)
		}
	})
}
