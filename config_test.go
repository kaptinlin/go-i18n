package i18n

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryConfig(t *testing.T) {
	t.Parallel()

	modulePath, moduleGoVersion := mustReadModuleConfig(t, "go.mod")

	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: ".gitignore includes baseline config patterns",
			path: ".gitignore",
			want: []string{
				".vscode/",
				".ralphy/*.json",
				"improve*.md",
				"plan*.md",
				"refactor*.md",
				"TODO*.md",
				"reports/*",
			},
		},
		{
			name: ".golangci uses canonical baseline",
			path: ".golangci.yml",
			want: []string{
				"version: \"2\"",
				"go: \"" + moduleGoVersion + "\"",
				"- errname",
				"- bodyclose",
				"- nolintlint",
				"- " + modulePath,
			},
		},
		{
			name: ".golangci.version pins the configured linter version",
			path: ".golangci.version",
			want: []string{"2.11.4"},
		},
		{
			name: "lefthook excludes research markdown but lints specs",
			path: "lefthook.yml",
			want: []string{
				"exclude:",
				".research/**/*.md",
				"stage_fixed: true",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content := mustReadFile(t, tc.path)
			for _, want := range tc.want {
				if !strings.Contains(content, want) {
					t.Fatalf("%s does not contain %q", tc.path, want)
				}
			}

			if tc.path == "lefthook.yml" && strings.Contains(content, "SPECS/**/*.md") {
				t.Fatalf("%s should lint SPECS markdown", tc.path)
			}
		})
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}

	return string(data)
}

func mustReadModuleConfig(t *testing.T, path string) (string, string) {
	t.Helper()

	var modulePath string
	var goVersion string

	for _, line := range strings.Split(mustReadFile(t, path), "\n") {
		line = strings.TrimSpace(line)

		if value, ok := strings.CutPrefix(line, "module "); ok {
			modulePath = strings.TrimSpace(value)
		}

		if value, ok := strings.CutPrefix(line, "go "); ok {
			goVersion = strings.TrimSpace(value)
		}
	}

	if modulePath == "" {
		t.Fatalf("%s does not contain a module directive", path)
	}

	if goVersion == "" {
		t.Fatalf("%s does not contain a go directive", path)
	}

	return modulePath, goVersion
}
