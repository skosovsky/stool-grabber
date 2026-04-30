package telegram

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	gotdsession "github.com/gotd/td/session"
)

var errNilAtomicStorage = errors.New("nil atomic file storage")

// AtomicFileStorage implements session.Storage with write-to-temp + rename into Path.
type AtomicFileStorage struct {
	Path string
	mux  sync.Mutex
}

func (f *AtomicFileStorage) LoadSession(_ context.Context) ([]byte, error) {
	if f == nil || f.Path == "" {
		return nil, errNilAtomicStorage
	}

	f.mux.Lock()
	defer f.mux.Unlock()

	data, err := os.ReadFile(f.Path)
	if os.IsNotExist(err) {
		return nil, gotdsession.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read session file: %w", err)
	}

	return data, nil
}

func (f *AtomicFileStorage) StoreSession(_ context.Context, data []byte) error {
	if f == nil || f.Path == "" {
		return errNilAtomicStorage
	}

	f.mux.Lock()
	defer f.mux.Unlock()

	dir := filepath.Dir(f.Path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create session directory: %w", err)
		}
	}

	tmp, err := os.CreateTemp(dir, ".session-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp session file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write temp session file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync temp session file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temp session file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o600); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("chmod temp session file: %w", err)
	}

	if err := os.Rename(tmpPath, f.Path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename session file: %w", err)
	}

	return nil
}
