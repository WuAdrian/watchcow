package fpkgen

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// TemplateEngine handles template loading and rendering
type TemplateEngine struct {
	templates map[string]*template.Template
}

// NewTemplateEngine creates a new template engine with embedded templates
func NewTemplateEngine() (*TemplateEngine, error) {
	engine := &TemplateEngine{
		templates: make(map[string]*template.Template),
	}

	// Load all embedded templates
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		content, err := templateFS.ReadFile("templates/" + name)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", name, err)
		}

		tmpl, err := template.New(name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		engine.templates[name] = tmpl
	}

	return engine, nil
}

// Render renders a template with the given data
func (e *TemplateEngine) Render(templateName string, data interface{}) ([]byte, error) {
	tmpl, ok := e.templates[templateName]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.Bytes(), nil
}

// RenderToFile renders a template and writes to file
func (e *TemplateEngine) RenderToFile(templateName, filePath string, data interface{}, perm os.FileMode) error {
	content, err := e.Render(templateName, data)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return os.WriteFile(filePath, content, perm)
}

// ListTemplates returns all available template names
func (e *TemplateEngine) ListTemplates() []string {
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}

// TemplateData holds all data needed for template rendering
type TemplateData struct {
	// Identity
	AppName     string
	Version     string
	DisplayName string
	Description string
	Maintainer  string

	// Container
	ContainerID   string
	ContainerName string
	Image         string

	// Network/UI
	Protocol string
	Port     string
	Path     string
	UIType   string
	AllUsers bool

	// Collections
	Ports       []string
	Volumes     []VolumeMapping
	Environment []string

	// Other
	RestartPolicy string
	Icon          string
}

// NewTemplateData creates TemplateData from AppConfig
func NewTemplateData(config *AppConfig) *TemplateData {
	data := &TemplateData{
		AppName:       config.AppName,
		Version:       config.Version,
		DisplayName:   config.DisplayName,
		Description:   escapeForTemplate(config.Description),
		Maintainer:    config.Maintainer,
		ContainerID:   config.ContainerID,
		ContainerName: config.ContainerName,
		Image:         config.Image,
		Protocol:      config.Protocol,
		Port:          config.Port,
		Path:          config.Path,
		UIType:        config.UIType,
		AllUsers:      config.AllUsers,
		Volumes:       config.Volumes,
		Environment:   config.Environment,
		RestartPolicy: config.RestartPolicy,
		Icon:          config.Icon,
	}

	// Set defaults
	if data.UIType == "" {
		data.UIType = "url"
	}
	if data.Protocol == "" {
		data.Protocol = "http"
	}
	if data.Path == "" {
		data.Path = "/"
	}
	if data.RestartPolicy == "" {
		data.RestartPolicy = "unless-stopped"
	}

	// Build ports list
	if config.Port != "" {
		data.Ports = []string{config.Port + ":" + config.Port}
	}

	return data
}

// escapeForTemplate escapes special characters for template output
func escapeForTemplate(s string) string {
	// For manifest, replace newlines
	s = replaceAll(s, "\n", " ")
	return s
}

func replaceAll(s, old, new string) string {
	for {
		idx := indexOf(s, old)
		if idx < 0 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
	return s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
