package telegram

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	gotdsession "github.com/gotd/td/session"
)

func TestAtomicFileStorage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	s := &AtomicFileStorage{Path: path}

	_, err := s.LoadSession(ctx)
	if err == nil || !errors.Is(err, gotdsession.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	payload := []byte(`{"version":1,"data":{"dc":2}}`)
	if err := s.StoreSession(ctx, payload); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm()&0o077 != 0 {
		t.Fatalf("session file is too permissive: %v", fi.Mode())
	}

	got, err := s.LoadSession(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("want %s, got %s", payload, got)
	}
}
