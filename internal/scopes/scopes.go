package scopes

import "slices"

// ScopeSlug represents an admin permission scope.
type ScopeSlug string

// Available scopes in the system.
const (
	// ScopeReadDashboard read the dashboard of the system.
	ScopeReadDashboard ScopeSlug = "read_dashboard"

	// ScopeReadChannels read the channels/models of the system.
	ScopeReadChannels ScopeSlug = "read_channels"
	// ScopeWriteChannels manage the channels/models of the system.
	ScopeWriteChannels ScopeSlug = "write_channels"

	// ScopeReadSettings read the system settings.
	ScopeReadSettings ScopeSlug = "read_settings"
	// ScopeWriteSettings manage the system settings.
	ScopeWriteSettings ScopeSlug = "write_settings"

	// ScopeReadAPIKeys read the API keys.
	//nolint:gosec // False positive.
	ScopeReadAPIKeys ScopeSlug = "read_api_keys"
	// ScopeWriteAPIKeys manage the API keys.
	ScopeWriteAPIKeys ScopeSlug = "write_api_keys"

	// ScopeReadRequests read request records.
	ScopeReadRequests ScopeSlug = "read_requests"
	// ScopeWriteRequests manage request records.
	ScopeWriteRequests ScopeSlug = "write_requests"
)

type ScopeLevel string

const (
	// ScopeLevelSystem is the scope level for system-wide operations.
	ScopeLevelSystem ScopeLevel = "system"
)

type Scope struct {
	Slug        ScopeSlug
	Description string
	Levels      []ScopeLevel
}

// scopeConfigs defines all available scopes with their configurations.
var scopeConfigs = []Scope{
	{
		Slug:        ScopeReadDashboard,
		Description: "View dashboard",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
	{
		Slug:        ScopeReadSettings,
		Description: "View system settings",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
	{
		Slug:        ScopeWriteSettings,
		Description: "Manage system settings",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
	{
		Slug:        ScopeReadChannels,
		Description: "View channel information",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
	{
		Slug:        ScopeWriteChannels,
		Description: "Manage channels/models (create, edit, delete)",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
	{
		Slug:        ScopeReadAPIKeys,
		Description: "View API keys",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
	{
		Slug:        ScopeWriteAPIKeys,
		Description: "Manage API keys (create, edit, delete)",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
	{
		Slug:        ScopeReadRequests,
		Description: "View request records",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
	{
		Slug:        ScopeWriteRequests,
		Description: "Manage request records",
		Levels:      []ScopeLevel{ScopeLevelSystem},
	},
}

// AllScopes returns all available scopes, optionally filtered by level.
func AllScopes(level *ScopeLevel) []Scope {
	if level == nil {
		return scopeConfigs
	}

	filtered := make([]Scope, 0)

	for _, scope := range scopeConfigs {
		if slices.Contains(scope.Levels, *level) {
			filtered = append(filtered, scope)
		}
	}

	return filtered
}

// AllScopesAsStrings returns all available scopes as strings.
func AllScopesAsStrings() []string {
	scopes := AllScopes(nil)

	result := make([]string, len(scopes))
	for i, scope := range scopes {
		result[i] = string(scope.Slug)
	}

	return result
}

// IsValidScope checks if a scope is valid.
func IsValidScope(scope string) bool {
	for _, validScope := range AllScopes(nil) {
		if string(validScope.Slug) == scope {
			return true
		}
	}

	return false
}
