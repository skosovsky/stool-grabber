package telegram

import (
	"context"
	"errors"
	"io"
	"net"
	"syscall"

	"github.com/gotd/td/tgerr"
)

const maxTransientRPCRetries = 4

// ShouldRetryTransientNetwork возвращает true для ошибок транспорта/временных сетевых сбоев.
// Не используется для FLOOD_WAIT / FLOOD_PREMIUM_WAIT — их обрабатывает tgerr.FloodWait; иначе
// получился бы «двойной» ретрай поверх ожидания Telegram.
//
// routery в проект не подключали: этот предикат — явный тонкий слой повторов только под сырой net.
func ShouldRetryTransientNetwork(err error) bool {
	if err == nil {
		return false
	}
	if _, fw := tgerr.AsFloodWait(err); fw {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return true
		}
		if ne.Temporary() {
			return true
		}
	}
	return false
}
