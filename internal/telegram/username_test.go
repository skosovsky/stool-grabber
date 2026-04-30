package telegram

import "testing"

func TestNormalizeChannelUsername(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"@channel", "channel"},
		{"  @name  ", "name"},
		{"plain", "plain"},
		{"", ""},
	}
	for _, tt := range tests {
		name := tt.in
		if name == "" {
			name = "empty_input"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeChannelUsername(tt.in); got != tt.want {
				t.Fatalf("NormalizeChannelUsername(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
