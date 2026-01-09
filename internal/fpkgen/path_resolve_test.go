package fpkgen

import (
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"
)

// **Feature: icon-format-support, Property 3: Relative Path Resolution**
// **Validates: Requirements 3.2, 3.3, 3.5, 3.6**
//
// For any relative path string (not starting with `/`) and a given basePath,
// the resolveFilePath function SHALL return a path equal to filepath.Join(basePath, relativePath).

// **Feature: icon-format-support, Property 4: Compose Working Directory Required for Relative Paths**
// **Validates: Requirements 3.4**
//
// For any relative path (not starting with `/`) when basePath is empty,
// the resolveFilePath function SHALL return an error indicating that compose working directory is required.

// TestGetBasePath_WithLabel tests getBasePath with compose working directory label
func TestGetBasePath_WithLabel(t *testing.T) {
	labels := map[string]string{
		"com.docker.compose.project.working_dir": "/home/user/project",
	}
	got := getBasePath(labels)
	want := "/home/user/project"
	if got != want {
		t.Errorf("getBasePath() = %q, want %q", got, want)
	}
}

// TestGetBasePath_WithoutLabel tests getBasePath without compose working directory label
func TestGetBasePath_WithoutLabel(t *testing.T) {
	labels := map[string]string{
		"other.label": "value",
	}
	got := getBasePath(labels)
	if got != "" {
		t.Errorf("getBasePath() = %q, want empty string", got)
	}
}

// TestGetBasePath_EmptyLabels tests getBasePath with empty labels
func TestGetBasePath_EmptyLabels(t *testing.T) {
	labels := map[string]string{}
	got := getBasePath(labels)
	if got != "" {
		t.Errorf("getBasePath() = %q, want empty string", got)
	}
}

// TestGetBasePath_NilLabels tests getBasePath with nil labels
func TestGetBasePath_NilLabels(t *testing.T) {
	got := getBasePath(nil)
	if got != "" {
		t.Errorf("getBasePath(nil) = %q, want empty string", got)
	}
}

// TestResolveFilePath_AbsolutePath tests resolveFilePath with absolute path
func TestResolveFilePath_AbsolutePath(t *testing.T) {
	got, err := resolveFilePath("file:///absolute/path/icon.png", "")
	if err != nil {
		t.Errorf("resolveFilePath() error = %v, want nil", err)
	}
	want := "/absolute/path/icon.png"
	if got != want {
		t.Errorf("resolveFilePath() = %q, want %q", got, want)
	}
}

// TestResolveFilePath_AbsolutePathWithBasePath tests resolveFilePath with absolute path ignores basePath
func TestResolveFilePath_AbsolutePathWithBasePath(t *testing.T) {
	got, err := resolveFilePath("file:///absolute/path/icon.png", "/some/base")
	if err != nil {
		t.Errorf("resolveFilePath() error = %v, want nil", err)
	}
	want := "/absolute/path/icon.png"
	if got != want {
		t.Errorf("resolveFilePath() = %q, want %q", got, want)
	}
}

// TestResolveFilePath_RelativePath tests resolveFilePath with relative path
func TestResolveFilePath_RelativePath(t *testing.T) {
	got, err := resolveFilePath("file://icon.png", "/home/user/project")
	if err != nil {
		t.Errorf("resolveFilePath() error = %v, want nil", err)
	}
	want := filepath.Join("/home/user/project", "icon.png")
	if got != want {
		t.Errorf("resolveFilePath() = %q, want %q", got, want)
	}
}

// TestResolveFilePath_RelativePathWithDot tests resolveFilePath with ./prefix
func TestResolveFilePath_RelativePathWithDot(t *testing.T) {
	got, err := resolveFilePath("file://./icons/app.png", "/home/user/project")
	if err != nil {
		t.Errorf("resolveFilePath() error = %v, want nil", err)
	}
	want := filepath.Join("/home/user/project", "./icons/app.png")
	if got != want {
		t.Errorf("resolveFilePath() = %q, want %q", got, want)
	}
}

// TestResolveFilePath_RelativePathWithParent tests resolveFilePath with ../prefix
func TestResolveFilePath_RelativePathWithParent(t *testing.T) {
	got, err := resolveFilePath("file://../shared/icon.png", "/home/user/project")
	if err != nil {
		t.Errorf("resolveFilePath() error = %v, want nil", err)
	}
	want := filepath.Join("/home/user/project", "../shared/icon.png")
	if got != want {
		t.Errorf("resolveFilePath() = %q, want %q", got, want)
	}
}

// TestResolveFilePath_RelativePathNoBasePath tests resolveFilePath with relative path but no basePath
func TestResolveFilePath_RelativePathNoBasePath(t *testing.T) {
	_, err := resolveFilePath("file://icon.png", "")
	if err == nil {
		t.Error("resolveFilePath() error = nil, want error for relative path without basePath")
	}
	if !strings.Contains(err.Error(), "relative path requires base path") {
		t.Errorf("resolveFilePath() error = %q, want error containing 'relative path requires base path'", err.Error())
	}
}

// Property test: Relative path resolution
// For any relative path and basePath, resolveFilePath returns filepath.Join(basePath, path)
func TestProperty_RelativePathResolution(t *testing.T) {
	f := func(relativePath, basePath string) bool {
		// Skip empty basePath (covered by Property 4)
		if basePath == "" {
			return true
		}
		// Skip paths that look absolute (start with /)
		if strings.HasPrefix(relativePath, "/") {
			return true
		}
		// Skip empty relative paths
		if relativePath == "" {
			return true
		}

		fileURL := "file://" + relativePath
		got, err := resolveFilePath(fileURL, basePath)
		if err != nil {
			return false
		}

		want := filepath.Join(basePath, relativePath)
		return got == want
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: Compose working directory required for relative paths
// For any relative path with empty basePath, resolveFilePath returns an error
func TestProperty_RelativePathRequiresBasePath(t *testing.T) {
	f := func(relativePath string) bool {
		// Skip paths that look absolute (start with /)
		if strings.HasPrefix(relativePath, "/") {
			return true
		}
		// Skip empty relative paths
		if relativePath == "" {
			return true
		}

		fileURL := "file://" + relativePath
		_, err := resolveFilePath(fileURL, "")
		return err != nil && strings.Contains(err.Error(), "relative path requires base path")
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: Absolute paths ignore basePath
// For any absolute path, resolveFilePath returns the path regardless of basePath
func TestProperty_AbsolutePathIgnoresBasePath(t *testing.T) {
	f := func(absolutePath, basePath string) bool {
		// Only test paths that start with /
		if !strings.HasPrefix(absolutePath, "/") {
			return true
		}

		fileURL := "file://" + absolutePath
		got, err := resolveFilePath(fileURL, basePath)
		if err != nil {
			return false
		}

		return got == absolutePath
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: getBasePath returns the label value when present
func TestProperty_GetBasePathReturnsLabelValue(t *testing.T) {
	f := func(workingDir string) bool {
		labels := map[string]string{
			"com.docker.compose.project.working_dir": workingDir,
		}
		return getBasePath(labels) == workingDir
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Property test: getBasePath returns empty string when label is absent
func TestProperty_GetBasePathReturnsEmptyWithoutLabel(t *testing.T) {
	f := func(otherKey, otherValue string) bool {
		// Ensure we don't accidentally use the compose label key
		if otherKey == "com.docker.compose.project.working_dir" {
			return true
		}
		labels := map[string]string{
			otherKey: otherValue,
		}
		return getBasePath(labels) == ""
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}
