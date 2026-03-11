package cli

// Exit codes for CLI commands. All CLI commands should use these
// constants instead of raw integer literals.
const (
	ExitSuccess        = 0
	ExitGeneralError   = 1
	ExitNotFound       = 2
	ExitValidation     = 3
	ExitProviderError  = 4
	ExitAmbiguousInput = 5

	// Doctor-specific exit codes (Story 49.10)
	ExitDoctorWarning = 1 // warnings only, no errors
	ExitDoctorError   = 2 // at least one error found
)
