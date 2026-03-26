package connection

import (
	"fmt"
	"sync"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// ProviderBridge adapts existing TaskProvider instances to the ConnectionManager
// lifecycle. It implements HealthChecker and Syncer so that ConnectionService
// can perform health checks and force-syncs through the existing provider API.
type ProviderBridge struct {
	mu        sync.RWMutex
	providers map[string]core.TaskProvider // connection ID → provider
}

// NewProviderBridge creates an empty bridge. Use Register to associate
// providers with connection IDs.
func NewProviderBridge() *ProviderBridge {
	return &ProviderBridge{
		providers: make(map[string]core.TaskProvider),
	}
}

// Register associates a TaskProvider with a connection ID.
func (b *ProviderBridge) Register(connID string, provider core.TaskProvider) {
	b.mu.Lock()
	b.providers[connID] = provider
	b.mu.Unlock()
}

// Unregister removes a provider association.
func (b *ProviderBridge) Unregister(connID string) {
	b.mu.Lock()
	delete(b.providers, connID)
	b.mu.Unlock()
}

// Provider returns the TaskProvider for a connection ID, or nil if not found.
func (b *ProviderBridge) Provider(connID string) core.TaskProvider {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.providers[connID]
}

// Providers returns all registered providers keyed by connection ID.
func (b *ProviderBridge) Providers() map[string]core.TaskProvider {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]core.TaskProvider, len(b.providers))
	for id, p := range b.providers {
		result[id] = p
	}
	return result
}

// CheckHealth implements HealthChecker by delegating to the provider's
// HealthCheck method and mapping core.HealthCheckResult → connection.HealthCheckResult.
func (b *ProviderBridge) CheckHealth(conn *Connection, _ string) (HealthCheckResult, error) {
	b.mu.RLock()
	provider, ok := b.providers[conn.ID]
	b.mu.RUnlock()

	if !ok {
		return HealthCheckResult{}, fmt.Errorf("check health: no provider for connection %s", conn.ID)
	}

	coreResult := provider.HealthCheck()
	return mapHealthCheckResult(coreResult), nil
}

// Sync implements Syncer by calling LoadTasks on the provider and updating
// the connection's task count.
func (b *ProviderBridge) Sync(conn *Connection, _ string) error {
	b.mu.RLock()
	provider, ok := b.providers[conn.ID]
	b.mu.RUnlock()

	if !ok {
		return fmt.Errorf("sync: no provider for connection %s", conn.ID)
	}

	tasks, err := provider.LoadTasks()
	if err != nil {
		return fmt.Errorf("sync connection %s (%s): %w", conn.ID, conn.ProviderName, err)
	}

	conn.TaskCount = len(tasks)
	conn.LastSync = time.Now().UTC()
	return nil
}

// mapHealthCheckResult converts a core.HealthCheckResult (items-based) to a
// connection.HealthCheckResult (bool-flags-based).
//
// Mapping rules:
//   - APIReachable: true if no item has status FAIL (any FAIL means something is unreachable)
//   - TokenValid: true unless an item named "Authentication" or "Auth" has status FAIL
//   - RateLimitOK: always true (core health checks don't track rate limits)
//   - TaskCount: extracted from the "Database" item message if present
//   - Details: all item names and messages
func mapHealthCheckResult(cr core.HealthCheckResult) HealthCheckResult {
	result := HealthCheckResult{
		APIReachable: true,
		TokenValid:   true,
		RateLimitOK:  true,
		Details:      make(map[string]string, len(cr.Items)),
	}

	for _, item := range cr.Items {
		result.Details[item.Name] = fmt.Sprintf("%s: %s", item.Status, item.Message)

		if item.Status == core.HealthFail {
			result.APIReachable = false
		}

		if item.Name == "Authentication" || item.Name == "Auth" {
			if item.Status == core.HealthFail {
				result.TokenValid = false
			}
		}

		// Extract task count from Database item
		if item.Name == "Database" && item.Status == core.HealthOK {
			var count int
			if _, err := fmt.Sscanf(item.Message, "%d tasks loaded", &count); err == nil {
				result.TaskCount = count
			}
		}
	}

	return result
}
