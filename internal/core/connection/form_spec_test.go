package connection

import "testing"

func TestAuthType_Constants(t *testing.T) {
	t.Parallel()

	// Verify enum values are distinct and in expected order.
	tests := []struct {
		name string
		got  AuthType
		want int
	}{
		{"AuthNone", AuthNone, 0},
		{"AuthAPIToken", AuthAPIToken, 1},
		{"AuthOAuth", AuthOAuth, 2},
		{"AuthLocalPath", AuthLocalPath, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if int(tt.got) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestFormSpec_Defaults(t *testing.T) {
	t.Parallel()

	// A zero-value FormSpec should have AuthNone and empty slices.
	var spec FormSpec
	if spec.AuthType != AuthNone {
		t.Errorf("zero FormSpec.AuthType = %d, want %d (AuthNone)", spec.AuthType, AuthNone)
	}
	if len(spec.AuthFields) != 0 {
		t.Errorf("zero FormSpec.AuthFields has %d elements, want 0", len(spec.AuthFields))
	}
	if len(spec.SyncFields) != 0 {
		t.Errorf("zero FormSpec.SyncFields has %d elements, want 0", len(spec.SyncFields))
	}
}

func TestFormField_Structure(t *testing.T) {
	t.Parallel()

	field := FormField{
		Key:         "api_token",
		Label:       "API Token",
		Description: "Your Todoist API token (Settings > Integrations)",
		Required:    true,
		Default:     "",
		Masked:      true,
		Options:     nil,
	}

	if field.Key != "api_token" {
		t.Errorf("Key = %q, want %q", field.Key, "api_token")
	}
	if !field.Required {
		t.Error("Required = false, want true")
	}
	if !field.Masked {
		t.Error("Masked = false, want true")
	}
	if len(field.Options) != 0 {
		t.Errorf("Options has %d elements, want 0 (text input)", len(field.Options))
	}
}

func TestFormSpec_APITokenProvider(t *testing.T) {
	t.Parallel()

	spec := FormSpec{
		AuthType: AuthAPIToken,
		AuthFields: []FormField{
			{Key: "api_token", Label: "API Token", Required: true, Masked: true},
		},
		TokenHelp:   "Find your token at Settings > Integrations > API token",
		Description: "Todoist task manager",
	}

	if spec.AuthType != AuthAPIToken {
		t.Errorf("AuthType = %d, want %d (AuthAPIToken)", spec.AuthType, AuthAPIToken)
	}
	if len(spec.AuthFields) != 1 {
		t.Fatalf("AuthFields has %d fields, want 1", len(spec.AuthFields))
	}
	if spec.AuthFields[0].Key != "api_token" {
		t.Errorf("AuthFields[0].Key = %q, want %q", spec.AuthFields[0].Key, "api_token")
	}
	if spec.TokenHelp == "" {
		t.Error("TokenHelp is empty")
	}
}

func TestFormSpec_LocalPathProvider(t *testing.T) {
	t.Parallel()

	spec := FormSpec{
		AuthType: AuthLocalPath,
		AuthFields: []FormField{
			{Key: "path", Label: "File path", Required: true},
		},
		Description: "Local YAML task file",
	}

	if spec.AuthType != AuthLocalPath {
		t.Errorf("AuthType = %d, want %d (AuthLocalPath)", spec.AuthType, AuthLocalPath)
	}
	if len(spec.AuthFields) != 1 {
		t.Fatalf("AuthFields has %d fields, want 1", len(spec.AuthFields))
	}
}

func TestFormSpec_OAuthProvider(t *testing.T) {
	t.Parallel()

	spec := FormSpec{
		AuthType:    AuthOAuth,
		AuthFields:  nil, // OAuth has no user-entered fields (device code flow)
		Description: "GitHub Issues",
	}

	if spec.AuthType != AuthOAuth {
		t.Errorf("AuthType = %d, want %d (AuthOAuth)", spec.AuthType, AuthOAuth)
	}
	if len(spec.AuthFields) != 0 {
		t.Errorf("AuthFields has %d fields, want 0 (OAuth uses device code)", len(spec.AuthFields))
	}
}

func TestFormSpec_WithSyncFields(t *testing.T) {
	t.Parallel()

	spec := FormSpec{
		AuthType: AuthAPIToken,
		SyncFields: []FormField{
			{Key: "project_ids", Label: "Project IDs", Description: "Comma-separated project IDs"},
			{Key: "filter", Label: "Filter", Description: "Todoist filter expression"},
		},
	}

	if len(spec.SyncFields) != 2 {
		t.Fatalf("SyncFields has %d fields, want 2", len(spec.SyncFields))
	}
	if spec.SyncFields[0].Key != "project_ids" {
		t.Errorf("SyncFields[0].Key = %q, want %q", spec.SyncFields[0].Key, "project_ids")
	}
	if spec.SyncFields[1].Key != "filter" {
		t.Errorf("SyncFields[1].Key = %q, want %q", spec.SyncFields[1].Key, "filter")
	}
}

// mockFormSpecProvider is a test double implementing FormSpecProvider.
type mockFormSpecProvider struct {
	spec FormSpec
}

func (m *mockFormSpecProvider) FormSpec() FormSpec {
	return m.spec
}

func TestFormSpecProvider_Interface(t *testing.T) {
	t.Parallel()

	provider := &mockFormSpecProvider{
		spec: FormSpec{
			AuthType:    AuthAPIToken,
			Description: "Test provider",
			TokenHelp:   "Use your test token",
		},
	}

	var _ FormSpecProvider = provider // compile-time check

	got := provider.FormSpec()
	if got.AuthType != AuthAPIToken {
		t.Errorf("FormSpec().AuthType = %d, want %d", got.AuthType, AuthAPIToken)
	}
	if got.Description != "Test provider" {
		t.Errorf("FormSpec().Description = %q, want %q", got.Description, "Test provider")
	}
}
