package connection

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// CredentialStore provides Get, Set, and Delete for credential storage.
type CredentialStore interface {
	// Get retrieves a credential by connection ID.
	Get(connID string) (string, error)

	// Set stores a credential for a connection ID.
	Set(connID, value string) error

	// Delete removes a credential for a connection ID.
	Delete(connID string) error
}

// ErrCredentialNotFound is returned when no credential exists for the given key.
var ErrCredentialNotFound = fmt.Errorf("credential not found")

// EnvCredentialStore resolves credentials from environment variables.
//
// It checks three patterns in priority order:
//  1. Connection-specific: THREEDOORS_CONN_<LABEL_SLUG>_TOKEN
//  2. Provider-level: THREEDOORS_<PROVIDER>_TOKEN
//  3. Provider aliases: GH_TOKEN / GITHUB_TOKEN for github provider
//
// Set and Delete are no-ops (env vars are read-only from the app's perspective).
type EnvCredentialStore struct {
	// lookupFn allows injecting a custom env lookup for testing.
	// Defaults to os.LookupEnv.
	lookupFn func(string) (string, bool)
}

// NewEnvCredentialStore creates an EnvCredentialStore that reads from os.LookupEnv.
func NewEnvCredentialStore() *EnvCredentialStore {
	return &EnvCredentialStore{lookupFn: os.LookupEnv}
}

// newEnvCredentialStoreWithLookup creates an EnvCredentialStore with a custom lookup
// function for testing.
func newEnvCredentialStoreWithLookup(fn func(string) (string, bool)) *EnvCredentialStore {
	return &EnvCredentialStore{lookupFn: fn}
}

// Get checks environment variables for a credential. It requires the connection
// to be resolved externally—the caller provides connID in the format "provider:label"
// (e.g., "todoist:Personal" or "github:OSS Projects").
//
// Priority: connection-specific env var > provider-level env var > provider aliases.
func (e *EnvCredentialStore) Get(connID string) (string, error) {
	provider, label := parseConnID(connID)

	// 1. Connection-specific: THREEDOORS_CONN_<LABEL_SLUG>_TOKEN
	if label != "" {
		slug := slugify(label)
		key := "THREEDOORS_CONN_" + slug + "_TOKEN"
		if val, ok := e.lookupFn(key); ok && val != "" {
			return val, nil
		}
	}

	// 2. Provider-level: THREEDOORS_<PROVIDER>_TOKEN
	if provider != "" {
		key := "THREEDOORS_" + strings.ToUpper(provider) + "_TOKEN"
		if val, ok := e.lookupFn(key); ok && val != "" {
			return val, nil
		}
	}

	// 3. Provider aliases
	for _, alias := range providerAliases(provider) {
		if val, ok := e.lookupFn(alias); ok && val != "" {
			return val, nil
		}
	}

	return "", fmt.Errorf("env credential %s: %w", connID, ErrCredentialNotFound)
}

// Set is a no-op for environment variables.
func (e *EnvCredentialStore) Set(_, _ string) error {
	return nil
}

// Delete is a no-op for environment variables.
func (e *EnvCredentialStore) Delete(_ string) error {
	return nil
}

// ChainCredentialStore tries multiple CredentialStores in priority order.
// The first store that returns a value wins. Set and Delete operate on the
// primary store (first in the chain that supports writes).
type ChainCredentialStore struct {
	stores       []CredentialStore
	writeStore   CredentialStore
	warnFn       func(string) // called when falling back to a lower-priority store
	fallbackUsed bool
}

// NewChainCredentialStore creates a chain that tries stores in order.
// writeStore is the store used for Set/Delete operations (typically keyring).
func NewChainCredentialStore(writeStore CredentialStore, stores ...CredentialStore) *ChainCredentialStore {
	return &ChainCredentialStore{
		stores:     stores,
		writeStore: writeStore,
	}
}

// SetWarnFunc sets a function called when a fallback store is used for reads.
func (c *ChainCredentialStore) SetWarnFunc(fn func(string)) {
	c.warnFn = fn
}

// Get tries each store in order and returns the first successful result.
func (c *ChainCredentialStore) Get(connID string) (string, error) {
	for i, store := range c.stores {
		val, err := store.Get(connID)
		if err == nil {
			if i > 0 && c.warnFn != nil {
				c.warnFn(fmt.Sprintf("credential for %s resolved from fallback store (index %d)", connID, i))
				c.fallbackUsed = true
			}
			return val, nil
		}
	}
	return "", fmt.Errorf("chain credential %s: %w", connID, ErrCredentialNotFound)
}

// Set stores the credential in the write store.
func (c *ChainCredentialStore) Set(connID, value string) error {
	if c.writeStore == nil {
		return fmt.Errorf("chain credential set: no write store configured")
	}
	return c.writeStore.Set(connID, value)
}

// Delete removes the credential from the write store.
func (c *ChainCredentialStore) Delete(connID string) error {
	if c.writeStore == nil {
		return fmt.Errorf("chain credential delete: no write store configured")
	}
	return c.writeStore.Delete(connID)
}

// MaskCredential replaces a credential value with a masked representation.
// Returns "••••" for any non-empty string. Returns empty string for empty input.
func MaskCredential(value string) string {
	if value == "" {
		return ""
	}
	return "••••"
}

// slugify converts a label to an uppercase slug for env var naming.
// "Work Jira" → "WORK_JIRA", "Personal Todoist" → "PERSONAL_TODOIST"
func slugify(label string) string {
	var b strings.Builder
	prevUnderscore := false
	for _, r := range label {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToUpper(r))
			prevUnderscore = false
		} else if !prevUnderscore && b.Len() > 0 {
			b.WriteByte('_')
			prevUnderscore = true
		}
	}
	// Trim trailing underscore.
	s := b.String()
	return strings.TrimRight(s, "_")
}

// parseConnID splits "provider:label" into its components.
// If no colon is found, the whole string is treated as the connection ID
// and provider/label are returned empty.
func parseConnID(connID string) (provider, label string) {
	idx := strings.IndexByte(connID, ':')
	if idx < 0 {
		return "", ""
	}
	return connID[:idx], connID[idx+1:]
}

// providerAliases returns well-known env var aliases for a provider.
func providerAliases(provider string) []string {
	switch strings.ToLower(provider) {
	case "github":
		return []string{"GH_TOKEN", "GITHUB_TOKEN"}
	case "linear":
		return []string{"LINEAR_API_KEY"}
	default:
		return nil
	}
}

// ConnCredentialKey builds the connID key format used by credential stores.
// Format: "provider:label"
func ConnCredentialKey(c *Connection) string {
	return c.ProviderName + ":" + c.Label
}
