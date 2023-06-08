package model

// Dependency represents a service dependency.
type Dependency struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

// Health represents the health status of the service and its dependencies.
type Health struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Dependencies []Dependency `json:"dependencies"`
}

// NewHealth creates a new Health instance.
func NewHealth(name, version string, dependencies []Dependency) *Health {
	return &Health{
		Dependencies: dependencies,
		Name:         name,
		Version:      version,
	}
}
