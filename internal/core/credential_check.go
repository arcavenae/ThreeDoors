package core

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// credentialKeys lists setting key names that indicate credential/secret values.
var credentialKeys = []string{
	"api_token",
	"token",
	"api_key",
	"secret",
	"password",
}

// WarnCredentialExposure checks whether the config file at configPath has
// permissions more permissive than 0o600 AND the loaded config contains
// non-empty credential fields. If both conditions are true, a warning is
// written to w with remediation instructions.
//
// Returns true if a warning was emitted.
func WarnCredentialExposure(w io.Writer, configPath string, cfg *ProviderConfig) bool {
	if cfg == nil {
		return false
	}

	info, err := os.Lstat(configPath)
	if err != nil {
		return false
	}

	if !isPermissive(info.Mode()) {
		return false
	}

	if !configHasCredentials(cfg) {
		return false
	}

	_, _ = fmt.Fprintf(w, "WARNING: %s has permissive permissions (%04o) and contains credentials.\n",
		configPath, info.Mode().Perm())
	_, _ = fmt.Fprintf(w, "Run: chmod 600 %s\n", configPath)
	return true
}

// isPermissive returns true if the file mode grants any access to group or others.
func isPermissive(mode fs.FileMode) bool {
	return mode.Perm()&0o077 != 0
}

// configHasCredentials checks whether any provider entry in the config contains
// non-empty values for known credential keys.
func configHasCredentials(cfg *ProviderConfig) bool {
	for _, p := range cfg.Providers {
		for k, v := range p.Settings {
			if v == "" {
				continue
			}
			lower := strings.ToLower(k)
			for _, ck := range credentialKeys {
				if lower == ck {
					return true
				}
			}
		}
	}
	return false
}
