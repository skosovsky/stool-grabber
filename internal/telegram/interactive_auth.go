package telegram

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type stdAuthenticator struct {
	r *bufio.Reader
	w io.Writer
}

// NewInteractiveAuthenticator prompts for phone number, OTP code (and optionally 2FA password) using r/w.
func NewInteractiveAuthenticator(in io.Reader, out io.Writer) auth.UserAuthenticator {
	return &stdAuthenticator{r: bufio.NewReader(in), w: out}
}

func (a *stdAuthenticator) Phone(_ context.Context) (string, error) {
	line, err := readLine(a.r, a.w, "Phone number (international, e.g. +79991234567): ")
	if err != nil {
		return "", err
	}
	if line == "" {
		return "", errors.New("phone number is empty")
	}
	return line, nil
}

func (a *stdAuthenticator) Code(_ context.Context, sent *tg.AuthSentCode) (string, error) {
	_ = sent
	line, err := readLine(a.r, a.w, "Authentication code from Telegram: ")
	if err != nil {
		return "", err
	}
	return line, nil
}

func (a *stdAuthenticator) Password(_ context.Context) (string, error) {
	line, err := readLine(a.r, a.w, "Two-step verification password (2FA): ")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (a *stdAuthenticator) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	if strings.TrimSpace(tos.Text) != "" {
		fmt.Fprintf(a.w, "Telegram terms of service excerpt:\n%s\n", strings.TrimSpace(tos.Text))
	}
	ok, err := readLine(a.r, a.w, "Register this account requires accepting terms. Proceed? [y/N]: ")
	if err != nil {
		return err
	}
	if strings.EqualFold(ok, "y") || strings.EqualFold(ok, "yes") {
		return nil
	}
	return fmt.Errorf("terms of service not accepted")
}

func (a *stdAuthenticator) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New(
		"account sign-up flow is not supported in stool-grabber; create the account using the Telegram app first",
	)
}

func readLine(r *bufio.Reader, w io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(w, prompt); err != nil {
		return "", err
	}
	line, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line != "" {
			return strings.TrimSpace(line), nil
		}
		return "", fmt.Errorf("read input (EOF?): %w", err)
	}
	return strings.TrimSpace(line), nil
}
