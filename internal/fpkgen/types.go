package fpkgen

// AppConfig holds all configuration for generating an fnOS app
type AppConfig struct {
	// Identity (matches fnOS manifest fields)
	AppName     string // manifest.appname - Unique app identifier
	Version     string // manifest.version - App version (e.g., "1.0.0")
	DisplayName string // manifest.display_name - Human-readable name
	Description string // manifest.desc - App description
	Maintainer  string // manifest.maintainer - Developer/maintainer name

	// Container Info
	ContainerID   string
	ContainerName string
	Image         string

	// Network / UI
	Protocol string // http or https
	Port     string // service_port
	Path     string // URL path
	UIType   string // "url" (new tab) or "iframe" (desktop window)
	AllUsers bool   // true = all users can access, false = admin only

	// Volumes
	Volumes []VolumeMapping

	// Environment
	Environment []string

	// Metadata
	Icon          string
	RestartPolicy string

	// Labels (original watchcow labels)
	Labels map[string]string
}

// VolumeMapping represents a container volume mount
type VolumeMapping struct {
	Source      string
	Destination string
	ReadOnly    bool
	Type        string // "bind" or "volume"
}
