package fpkgen

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Generator handles fnOS application package generation from Docker containers
type Generator struct {
	outputDir      string                // Base output directory for generated apps
	dockerClient   *client.Client        // Docker API client
	templateEngine *TemplateEngine       // Template engine for rendering
	installed      map[string]*AppConfig // map[containerID]AppConfig - installed apps
	mu             sync.RWMutex          // Protects installed map
}

// NewGenerator creates a new application generator
func NewGenerator(outputDir string) (*Generator, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Initialize template engine
	tmplEngine, err := NewTemplateEngine()
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("failed to create template engine: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		cli.Close()
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &Generator{
		outputDir:      outputDir,
		dockerClient:   cli,
		templateEngine: tmplEngine,
		installed:      make(map[string]*AppConfig),
	}, nil
}

// GenerateFromContainer creates fnOS app structure from a running container
func (g *Generator) GenerateFromContainer(ctx context.Context, containerID string) (*AppConfig, string, error) {
	// 1. Inspect container for full details
	container, err := g.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to inspect container: %w", err)
	}

	// 2. Extract configuration from container
	config := g.extractConfig(&container)

	// 3. Create application directory
	appDir := filepath.Join(g.outputDir, config.AppName)

	// Remove existing directory if exists
	if err := os.RemoveAll(appDir); err != nil {
		return nil, "", fmt.Errorf("failed to remove existing directory: %w", err)
	}

	if err := g.createDirectoryStructure(appDir); err != nil {
		return nil, "", fmt.Errorf("failed to create directory structure: %w", err)
	}

	// 4. Generate all files using templates
	slog.Info("Generating fnOS app package", "appName", config.AppName, "container", config.ContainerName)

	data := NewTemplateData(config)

	if err := g.generateFromTemplates(appDir, data); err != nil {
		return nil, "", err
	}

	if err := g.handleIcons(appDir, config); err != nil {
		return nil, "", fmt.Errorf("failed to handle icons: %w", err)
	}

	slog.Info("Successfully generated fnOS app package", "appDir", appDir)

	return config, appDir, nil
}

// GenerateFromConfig creates fnOS app structure from an AppConfig directly
// This is useful for testing/debugging without needing a real Docker container
func (g *Generator) GenerateFromConfig(config *AppConfig, appDir string) error {
	// Remove existing directory if exists
	if err := os.RemoveAll(appDir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	if err := g.createDirectoryStructure(appDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate all files using templates
	slog.Info("Generating fnOS app package from config", "appName", config.AppName)

	data := NewTemplateData(config)

	if err := g.generateFromTemplates(appDir, data); err != nil {
		return err
	}

	if err := g.handleIcons(appDir, config); err != nil {
		return fmt.Errorf("failed to handle icons: %w", err)
	}

	slog.Info("Successfully generated fnOS app package", "appDir", appDir)
	return nil
}

// generateFromTemplates generates all files using template engine
func (g *Generator) generateFromTemplates(appDir string, data *TemplateData) error {
	// Define template -> file mappings
	mappings := []struct {
		template string
		path     string
		perm     os.FileMode
	}{
		{"manifest.tmpl", "manifest", 0644},
		{"cmd_main.tmpl", "cmd/main", 0755},
		{"config_privilege.json.tmpl", "config/privilege", 0644},
		{"config_resource.json.tmpl", "config/resource", 0644},
		{"ui_config.json.tmpl", "app/ui/config", 0644},
		{"LICENSE.tmpl", "LICENSE", 0644},
	}

	for _, m := range mappings {
		filePath := filepath.Join(appDir, m.path)
		if err := g.templateEngine.RenderToFile(m.template, filePath, data, m.perm); err != nil {
			return fmt.Errorf("failed to generate %s: %w", m.path, err)
		}
	}

	// Generate empty cmd scripts
	cmdScripts := []string{"install_init", "install_callback", "uninstall_init", "uninstall_callback",
		"upgrade_init", "upgrade_callback", "config_init", "config_callback"}
	for _, script := range cmdScripts {
		filePath := filepath.Join(appDir, "cmd", script)
		if err := g.templateEngine.RenderToFile("cmd_empty.tmpl", filePath, data, 0755); err != nil {
			return fmt.Errorf("failed to generate cmd/%s: %w", script, err)
		}
	}

	return nil
}

// extractConfig extracts AppConfig from container inspection result
// Label naming follows fnOS manifest conventions:
//
//	watchcow.appname      -> manifest.appname
//	watchcow.display_name -> manifest.display_name
//	watchcow.desc         -> manifest.desc
//	watchcow.version      -> manifest.version
//	watchcow.maintainer   -> manifest.maintainer
//	watchcow.service_port -> manifest.service_port
//	watchcow.protocol     -> UI config (http/https)
//	watchcow.path         -> UI config (url path)
//	watchcow.icon         -> app icon URL
func (g *Generator) extractConfig(container *dockercontainer.InspectResponse) *AppConfig {
	name := strings.TrimPrefix(container.Name, "/")
	labels := container.Config.Labels

	// Generate sanitized app name
	sanitizedName := sanitizeAppName(name)
	appName := getLabel(labels, "watchcow.appname", fmt.Sprintf("watchcow.%s", sanitizedName))

	config := &AppConfig{
		AppName:       appName,
		Version:       getLabel(labels, "watchcow.version", "1.0.0"),
		DisplayName:   getLabel(labels, "watchcow.display_name", prettifyName(name)),
		Description:   getLabel(labels, "watchcow.desc", fmt.Sprintf("Docker container: %s", container.Config.Image)),
		Maintainer:    getLabel(labels, "watchcow.maintainer", "WatchCow"),
		ContainerID:   container.ID[:12],
		ContainerName: name,
		Image:         container.Config.Image,
		Protocol:      getLabel(labels, "watchcow.protocol", "http"),
		Port:          getLabel(labels, "watchcow.service_port", ""),
		Path:          getLabel(labels, "watchcow.path", "/"),
		UIType:        getLabel(labels, "watchcow.ui_type", "url"),
		Icon:          getLabel(labels, "watchcow.icon", guessIcon(container.Config.Image)),
		Environment:   filterEnvironment(container.Config.Env),
		Labels:        labels,
	}

	// Extract port if not specified in label
	if config.Port == "" {
		config.Port = extractFirstPort(container)
	}

	// Extract volumes
	for _, mount := range container.Mounts {
		config.Volumes = append(config.Volumes, VolumeMapping{
			Source:      mount.Source,
			Destination: mount.Destination,
			ReadOnly:    !mount.RW,
			Type:        string(mount.Type),
		})
	}

	// Extract restart policy
	if container.HostConfig.RestartPolicy.Name != "" {
		config.RestartPolicy = string(container.HostConfig.RestartPolicy.Name)
	} else {
		config.RestartPolicy = "unless-stopped"
	}

	return config
}

// createDirectoryStructure creates all required directories
func (g *Generator) createDirectoryStructure(appDir string) error {
	dirs := []string{
		appDir,
		filepath.Join(appDir, "app", "ui", "images"),
		filepath.Join(appDir, "cmd"),
		filepath.Join(appDir, "config"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// IsInstalled checks if a container has already been installed as fnOS app
func (g *Generator) IsInstalled(containerID string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, exists := g.installed[containerID]
	return exists
}

// GetInstalledApp gets the installed app config for a container
func (g *Generator) GetInstalledApp(containerID string) *AppConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.installed[containerID]
}

// MarkInstalled marks a container as installed
func (g *Generator) MarkInstalled(containerID string, config *AppConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.installed[containerID] = config
}

// MarkUninstalled removes a container from the installed list
func (g *Generator) MarkUninstalled(containerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.installed, containerID)
}

// GetAllInstalled returns all installed apps
func (g *Generator) GetAllInstalled() map[string]*AppConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*AppConfig)
	for k, v := range g.installed {
		result[k] = v
	}
	return result
}

// Close closes the Docker client
func (g *Generator) Close() error {
	if g.dockerClient != nil {
		return g.dockerClient.Close()
	}
	return nil
}

// Helper functions

// sanitizeAppName ensures the app name conforms to fnOS requirements
func sanitizeAppName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")
	var result strings.Builder
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// getLabel gets a label value with fallback
func getLabel(labels map[string]string, key, fallback string) string {
	if val, ok := labels[key]; ok && val != "" {
		return val
	}
	return fallback
}

// filterEnvironment removes sensitive/unwanted environment variables
func filterEnvironment(env []string) []string {
	var filtered []string
	blacklist := []string{"PATH=", "HOME=", "USER=", "HOSTNAME=", "PWD=", "SHLVL="}

	for _, e := range env {
		skip := false
		for _, b := range blacklist {
			if strings.HasPrefix(e, b) {
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

// extractFirstPort extracts the first public port from container
func extractFirstPort(container *dockercontainer.InspectResponse) string {
	if container.HostConfig == nil {
		return ""
	}

	for _, bindings := range container.HostConfig.PortBindings {
		for _, binding := range bindings {
			if binding.HostPort != "" {
				return binding.HostPort
			}
		}
	}

	return ""
}

// guessIcon tries to guess an appropriate icon URL based on image name
func guessIcon(image string) string {
	parts := strings.Split(image, "/")
	imageName := parts[len(parts)-1]
	imageName = strings.Split(imageName, ":")[0]

	iconMap := map[string]string{
		"jellyfin": "jellyfin", "portainer": "portainer", "nginx": "nginx",
		"postgres": "postgresql", "postgresql": "postgresql", "mysql": "mysql",
		"mariadb": "mariadb", "redis": "redis", "mongodb": "mongodb", "mongo": "mongodb",
		"plex": "plex", "sonarr": "sonarr", "radarr": "radarr", "traefik": "traefik",
		"grafana": "grafana", "prometheus": "prometheus", "homeassistant": "home-assistant",
		"nextcloud": "nextcloud", "gitea": "gitea", "gitlab": "gitlab", "jenkins": "jenkins",
		"minio": "minio", "rabbitmq": "rabbitmq", "elasticsearch": "elasticsearch",
		"kibana": "kibana", "caddy": "caddy", "apache": "apache", "httpd": "apache",
		"wordpress": "wordpress", "ghost": "ghost", "discourse": "discourse",
		"memos": "memos", "vaultwarden": "vaultwarden", "bitwarden": "bitwarden",
	}

	if iconName, ok := iconMap[strings.ToLower(imageName)]; ok {
		return fmt.Sprintf("https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/png/%s.png", iconName)
	}

	return "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/png/docker.png"
}

// prettifyName converts container name to a nice title
func prettifyName(name string) string {
	name = strings.TrimSuffix(name, "-1")
	name = strings.TrimSuffix(name, "_1")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}
