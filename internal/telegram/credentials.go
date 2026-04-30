// Package telegram provides MTProto connectivity (session management, scraping utilities).
//
// Environment variables for Task 03:
//
//	TG_APP_ID      — Telegram app id (positive integer).
//	TG_APP_HASH    — Telegram app hash.
//	TG_SESSION_PATH — Optional path for session JSON (default ./session.json).
//
// Smoke (manual): first run executes interactive login and creates the session file; the second run
// skips prompts if the saved session remains valid (see TD Definition of Done).
package telegram

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// DefaultSessionPath relative to cwd as in TD (“session.json”).
const DefaultSessionPath = "session.json"

// Credentials carries Telegram developer app identifiers from the environment (not job.yaml).
type Credentials struct {
	AppID       int
	AppHash     string
	SessionPath string
}

var (
	ErrMissingTGAppID   = errors.New("TG_APP_ID is required (integer Telegram app ID from https://my.telegram.org)")
	ErrInvalidTGAppID   = errors.New("TG_APP_ID must be a positive integer")
	ErrMissingTGAppHash = errors.New("TG_APP_HASH is required (Telegram app hash)")
)

// NewCredentials validates and constructs Credentials.
// Environment parsing is handled in the composition root (internal/app).
func NewCredentials(appIDRaw string, appHash string, sessionPath string) (Credentials, error) {
	rawID := strings.TrimSpace(appIDRaw)
	if rawID == "" {
		return Credentials{}, ErrMissingTGAppID
	}
	appID, err := strconv.Atoi(rawID)
	if err != nil || appID <= 0 {
		return Credentials{}, fmt.Errorf("%w: %q", ErrInvalidTGAppID, rawID)
	}
	appHash = strings.TrimSpace(appHash)
	if appHash == "" {
		return Credentials{}, ErrMissingTGAppHash
	}
	sessionPath = strings.TrimSpace(sessionPath)
	if sessionPath == "" {
		sessionPath = DefaultSessionPath
	}
	return Credentials{AppID: appID, AppHash: appHash, SessionPath: sessionPath}, nil
}
