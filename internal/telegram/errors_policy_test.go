package telegram

import (
	"context"
	"io"
	"net"
	"syscall"
	"testing"

	"github.com/gotd/td/tgerr"
)

func TestShouldRetryTransientNetwork(t *testing.T) {
	t.Parallel()
	flood := tgerr.New(420, "FLOOD_WAIT_10")
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"flood_wait", flood, false},
		{"eof", io.EOF, true},
		{"deadline", context.DeadlineExceeded, true},
		{"canceled", context.Canceled, false},
		{"econnreset", syscall.ECONNRESET, true},
		{"timeout_err", &timeoutError{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ShouldRetryTransientNetwork(tt.err); got != tt.want {
				t.Fatalf("ShouldRetryTransientNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return false }

var _ net.Error = (*timeoutError)(nil)
