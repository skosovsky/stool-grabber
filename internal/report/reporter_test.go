package report_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"stool-grabber/internal/report"
	"stool-grabber/internal/reportfs"
)

func TestWriteMarkdownFile_CreatesDirAndOverwrites(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "report.md")

	if err := reportfs.WriteMarkdownFile(path, "first"); err != nil {
		t.Fatalf("WriteMarkdownFile(first): %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(b) != "first" {
		t.Fatalf("content=%q, want %q", string(b), "first")
	}

	if err := reportfs.WriteMarkdownFile(path, "second"); err != nil {
		t.Fatalf("WriteMarkdownFile(second): %v", err)
	}
	b, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(b) != "second" {
		t.Fatalf("content=%q, want %q", string(b), "second")
	}
}

func TestRenderSkipReport_ContainsDeterministicText(t *testing.T) {
	t.Parallel()

	got := report.RenderSkipReport()
	if got == "" {
		t.Fatalf("empty skip report")
	}
	if want := "Целевое ядро не выявлено по текущим порогам фильтрации."; want != "" && !contains(got, want) {
		t.Fatalf("skip report missing %q: %q", want, got)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

