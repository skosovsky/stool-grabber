package telegram

import (
	"errors"
	"testing"
)

func TestNewCredentials(t *testing.T) {
	t.Run("ok with defaults", func(t *testing.T) {
		creds, err := NewCredentials("12345", "abcd", "")
		if err != nil {
			t.Fatal(err)
		}
		if creds.AppID != 12345 || creds.AppHash != "abcd" || creds.SessionPath != DefaultSessionPath {
			t.Fatalf("unexpected creds: %+v", creds)
		}
	})

	t.Run("explicit session path", func(t *testing.T) {
		creds, err := NewCredentials("1", "x", "/tmp/tg_sess.json")
		if err != nil {
			t.Fatal(err)
		}
		if creds.SessionPath != "/tmp/tg_sess.json" {
			t.Fatalf("unexpected session path: %q", creds.SessionPath)
		}
	})

	t.Run("missing app id", func(t *testing.T) {
		_, err := NewCredentials("", "h", "")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrMissingTGAppID) {
			t.Fatalf("expected ErrMissingTGAppID, got %v", err)
		}
	})

	t.Run("invalid app id", func(t *testing.T) {
		_, err := NewCredentials("0", "h", "")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrInvalidTGAppID) {
			t.Fatalf("expected ErrInvalidTGAppID, got %v", err)
		}
	})

	t.Run("non-numeric app id", func(t *testing.T) {
		_, err := NewCredentials("not-int", "h", "")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrInvalidTGAppID) {
			t.Fatalf("expected ErrInvalidTGAppID, got %v", err)
		}
	})

	t.Run("missing hash", func(t *testing.T) {
		_, err := NewCredentials("1", "  ", "")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrMissingTGAppHash) {
			t.Fatalf("expected ErrMissingTGAppHash, got %v", err)
		}
	})
}
