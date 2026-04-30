// Package reportfs provides filesystem persistence for rendered reports.
package reportfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteMarkdownFile(path string, content string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("report path is empty")
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o640); err != nil {
		return fmt.Errorf("write report file: %w", err)
	}
	return nil
}

func WriteJSONFile(path string, v any) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("json path is empty")
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o640); err != nil {
		return fmt.Errorf("write json file: %w", err)
	}
	return nil
}

