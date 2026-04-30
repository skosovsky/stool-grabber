package telegram

import (
	"context"
	"fmt"
	"io"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// SessionFunc вызывается внутри одного client.Run после успешной авторизации.
type SessionFunc func(ctx context.Context, api *tg.Client) error

// AuthorizedSessionRun поднимает MTProto-клиент (сессия + NoUpdates), выполняет auth и fn в одном Run.
func AuthorizedSessionRun(ctx context.Context, creds Credentials, in io.Reader, out io.Writer, fn SessionFunc) error {
	if fn == nil {
		return fmt.Errorf("telegram session: fn is nil")
	}

	storage := &AtomicFileStorage{Path: creds.SessionPath}
	client := telegram.NewClient(creds.AppID, creds.AppHash, telegram.Options{
		SessionStorage: storage,
		NoUpdates:      true,
	})
	flow := auth.NewFlow(NewInteractiveAuthenticator(in, out), auth.SendCodeOptions{})

	return client.Run(ctx, func(runCtx context.Context) error {
		if err := client.Auth().IfNecessary(runCtx, flow); err != nil {
			return fmt.Errorf("authenticate: %w", err)
		}
		api := tg.NewClient(client)
		return fn(runCtx, api)
	})
}

// EnsureAuthorizedSession проверяет сессию тривиальным RPC users.getUsers(self).
func EnsureAuthorizedSession(ctx context.Context, creds Credentials, in io.Reader, out io.Writer) error {
	return AuthorizedSessionRun(ctx, creds, in, out, func(ctx context.Context, api *tg.Client) error {
		return PrintAuthorizedUser(ctx, api, out)
	})
}

// PrintAuthorizedUser выполняет users.getUsers(self) и печатает краткую строку об аккаунте.
func PrintAuthorizedUser(ctx context.Context, api *tg.Client, out io.Writer) error {
	users, err := api.UsersGetUsers(ctx, []tg.InputUserClass{
		&tg.InputUserSelf{},
	})
	if err != nil {
		return fmt.Errorf("users.getUsers(self): %w", err)
	}
	if len(users) == 0 {
		return fmt.Errorf("users.getUsers: empty result")
	}
	if out == nil {
		return nil
	}
	switch u := users[0].(type) {
	case *tg.User:
		if u.Username != "" {
			_, _ = fmt.Fprintf(out, "Signed in as @%s (id %d).\n", u.Username, u.ID)
		} else {
			_, _ = fmt.Fprintf(out, "Signed in as user id %d.\n", u.ID)
		}
	default:
		_, _ = fmt.Fprintf(out, "Session OK (authorization verified).\n")
	}
	return nil
}
