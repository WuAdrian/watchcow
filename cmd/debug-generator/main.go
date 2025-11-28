package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"watchcow/internal/fpkgen"
)

func main() {
	// Flags
	outputDir := flag.String("output", "./debug-output", "Output directory for generated app")
	flag.Parse()

	// Configure logging
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	// Parse key=value arguments
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	// Build config from arguments
	config := &fpkgen.AppConfig{
		// Defaults
		AppName:       "debug.test-app",
		Version:       "1.0.0",
		DisplayName:   "Test App",
		Description:   "Debug generated app",
		Maintainer:    "WatchCow Debug",
		ContainerID:   "debug123456",
		ContainerName: "debug-container",
		Image:         "nginx:latest",
		Protocol:      "http",
		Port:          "8080",
		Path:          "/",
		UIType:        "url",
		Icon:          "",
		RestartPolicy: "unless-stopped",
		Labels:        make(map[string]string),
	}

	// Override with provided key=value pairs
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Invalid argument: %s (expected key=value)\n", arg)
			continue
		}
		key, value := parts[0], parts[1]
		applyConfig(config, key, value)
	}

	// Print config
	fmt.Println("=== Configuration ===")
	fmt.Printf("AppName:      %s\n", config.AppName)
	fmt.Printf("DisplayName:  %s\n", config.DisplayName)
	fmt.Printf("Version:      %s\n", config.Version)
	fmt.Printf("Description:  %s\n", config.Description)
	fmt.Printf("Maintainer:   %s\n", config.Maintainer)
	fmt.Printf("Port:         %s\n", config.Port)
	fmt.Printf("Protocol:     %s\n", config.Protocol)
	fmt.Printf("Path:         %s\n", config.Path)
	fmt.Printf("UIType:       %s\n", config.UIType)
	fmt.Printf("Icon:         %s\n", config.Icon)
	fmt.Printf("Image:        %s\n", config.Image)
	fmt.Printf("Output:       %s\n", *outputDir)
	fmt.Println()

	// Create generator (uses template engine internally)
	generator, err := fpkgen.NewGenerator()
	if err != nil {
		slog.Error("Failed to create generator", "error", err)
		os.Exit(1)
	}
	defer generator.Close()

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		slog.Error("Failed to create output directory", "error", err)
		os.Exit(1)
	}

	// Generate using template engine
	if err := generator.GenerateFromConfig(config, *outputDir); err != nil {
		slog.Error("Failed to generate app", "error", err)
		os.Exit(1)
	}

	fmt.Println("=== Generated Files ===")
	printTree(*outputDir, "")
	fmt.Println()
	fmt.Printf("Output directory: %s\n", *outputDir)
}

func printUsage() {
	fmt.Println("Debug Generator - Generate fnOS app directory from key=value pairs")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  debug-generator [flags] key=value [key=value ...]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -output string   Output directory (default \"./debug-output\")")
	fmt.Println()
	fmt.Println("Supported keys (following fnOS manifest conventions):")
	fmt.Println("  appname        - App identifier (e.g., watchcow.myapp)")
	fmt.Println("  display_name   - Display name")
	fmt.Println("  desc           - Description")
	fmt.Println("  version        - Version (e.g., 1.0.0)")
	fmt.Println("  maintainer     - Maintainer name")
	fmt.Println("  service_port   - Service port")
	fmt.Println("  protocol       - http or https")
	fmt.Println("  path           - URL path")
	fmt.Println("  ui_type        - url or iframe")
	fmt.Println("  icon           - Icon URL")
	fmt.Println("  image          - Docker image name")
	fmt.Println("  container_name - Container name")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  debug-generator appname=watchcow.nginx display_name=\"Nginx Server\" service_port=80")
}

func applyConfig(config *fpkgen.AppConfig, key, value string) {
	switch key {
	case "appname":
		config.AppName = value
	case "display_name":
		config.DisplayName = value
	case "desc":
		config.Description = value
	case "version":
		config.Version = value
	case "maintainer":
		config.Maintainer = value
	case "service_port":
		config.Port = value
	case "protocol":
		config.Protocol = value
	case "path":
		config.Path = value
	case "ui_type":
		config.UIType = value
	case "icon":
		config.Icon = value
	case "image":
		config.Image = value
	case "container_name":
		config.ContainerName = value
	default:
		fmt.Printf("Unknown key: %s\n", key)
	}
}

func printTree(path string, prefix string) {
	entries, _ := os.ReadDir(path)
	for i, entry := range entries {
		connector := "├── "
		if i == len(entries)-1 {
			connector = "└── "
		}
		fmt.Printf("%s%s%s\n", prefix, connector, entry.Name())
		if entry.IsDir() {
			newPrefix := prefix + "│   "
			if i == len(entries)-1 {
				newPrefix = prefix + "    "
			}
			printTree(path+"/"+entry.Name(), newPrefix)
		}
	}
}
