# /validate-adapter — Validate TaskProvider Implementation

Check that TaskProvider implementations are complete and correct. See CLAUDE.md for the adapter and factory patterns used in this project.

## Instructions

1. Read `internal/tasks/provider.go` to get the current `TaskProvider` interface definition and all required methods.

2. Find all types that implement `TaskProvider`:
   ```bash
   grep -rn "func (.*) LoadTasks" internal/tasks/
   ```

3. For each implementing type, verify:
   - All interface methods are implemented (check against the interface definition from step 1)
   - Error wrapping uses `fmt.Errorf("...: %w", err)` pattern (not bare `return err` or `%v`)
   - The type is registered in the provider factory (check `provider_factory.go` or equivalent switch/map)
   - A corresponding `_test.go` file exists with test coverage
   - For file-based providers: no direct file I/O outside the atomic write pattern (write to `.tmp`, sync, rename)

4. Report compliance for each adapter in a table:

   | Adapter | Methods | Error Wrapping | Factory | Tests | Atomic Writes |
   |---------|---------|---------------|---------|-------|--------------|
   | TextFileProvider | OK/MISSING | OK/ISSUES | OK/MISSING | OK/MISSING | OK/N/A |

5. Flag any specific issues found with file:line references.

6. If all adapters pass, say "All TaskProvider implementations are compliant."
   If issues are found, list them with suggested fixes.
