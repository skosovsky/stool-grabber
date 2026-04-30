package telegram

import "strings"

// NormalizeChannelUsername убирает ведущий @ и пробелы для contacts.resolveUsername.
func NormalizeChannelUsername(username string) string {
	s := strings.TrimSpace(username)
	s = strings.TrimPrefix(s, "@")
	return strings.TrimSpace(s)
}
