package connection

// AuthType indicates what kind of authentication a provider requires.
type AuthType int

const (
	// AuthNone indicates no authentication is needed.
	AuthNone AuthType = iota
	// AuthAPIToken indicates the provider authenticates via an API token.
	AuthAPIToken
	// AuthOAuth indicates the provider uses OAuth device code flow.
	AuthOAuth
	// AuthLocalPath indicates the provider requires a local file path.
	AuthLocalPath
)

// FormField defines a single form input field for provider configuration.
type FormField struct {
	Key         string   // settings map key
	Label       string   // display label
	Description string   // help text shown below the field
	Required    bool     // whether the field must be non-empty
	Default     string   // default value (pre-filled)
	Masked      bool     // true for secrets/tokens (password input)
	Options     []string // non-empty = select field; empty = text input
}

// FormSpec declares the form fields a provider needs for the setup wizard.
type FormSpec struct {
	AuthType    AuthType    // determines Step 2 layout
	AuthFields  []FormField // Step 2: authentication/config fields
	SyncFields  []FormField // Step 3: provider-specific filter fields
	TokenHelp   string      // where to find the API token (shown in Step 2)
	Description string      // brief provider description for Step 1 list
}

// FormSpecProvider is implemented by providers that support the setup wizard.
// Providers that implement this interface appear in the :connect wizard
// with their custom form fields. Providers that don't implement it are
// still listed but use a generic configuration form.
type FormSpecProvider interface {
	FormSpec() FormSpec
}
