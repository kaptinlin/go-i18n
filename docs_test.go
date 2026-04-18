package i18n

import (
	"os"
	"strings"
	"testing"
)

func TestDocumentationMentionsCurrentAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "README covers current public entry points",
			path: "README.md",
			want: []string{
				"[AGENTS.md](AGENTS.md)",
				"[SPECS/00-overview.md](SPECS/00-overview.md)",
				"SupportedLocales",
				"Lookup",
				"NewDetector",
				"middleware.HTTPMiddleware",
				"github.com/goccy/go-yaml",
			},
		},
		{
			name: "CLAUDE captures doc maintenance workflow",
			path: "CLAUDE.md",
			want: []string{
				"task markdownlint",
				"reports/<dependency-name>.md",
				"SPECS/00-overview.md",
				"go-i18n-localizing",
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
		})
	}
}

func TestAGENTSSymlinkPointsToCLAUDE(t *testing.T) {
	t.Parallel()

	target, err := os.Readlink("AGENTS.md")
	if err != nil {
		t.Fatalf("Readlink(AGENTS.md): %v", err)
	}

	if target != "CLAUDE.md" {
		t.Fatalf("AGENTS.md points to %q, want %q", target, "CLAUDE.md")
	}
}
