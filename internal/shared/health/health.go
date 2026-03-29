package health

// DependencyStatus holds the health state of a single dependency.
type DependencyStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthResult is the aggregated health status of the system.
type HealthResult struct {
	Status       string                      `json:"status"`
	Dependencies map[string]DependencyStatus `json:"dependencies"`
}
