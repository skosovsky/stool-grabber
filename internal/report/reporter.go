package report

import (
	"fmt"
	"strconv"
	"strings"

	"stool-grabber/internal/ai/contractgen"
	"stool-grabber/internal/domain"
)

const skipText = "Целевое ядро не выявлено по текущим порогам фильтрации."

func RenderSkipReport() string {
	return "# Отчёт\n\n" + skipText + "\n"
}

func RenderLLMReport(params Params, result *contractgen.AnalyzeCoreOutput, users map[int64]domain.UserRef) string {
	channel := strings.TrimPrefix(strings.TrimSpace(params.ChannelUsername), "@")
	model := strings.TrimSpace(params.Model)

	var b strings.Builder
	b.WriteString("# Отчёт\n\n")
	if channel != "" {
		b.WriteString(fmt.Sprintf("Канал: @%s\n\n", channel))
	}
	if model != "" {
		b.WriteString(fmt.Sprintf("Модель: `%s`\n\n", model))
	}

	if result == nil {
		b.WriteString("Результат анализа отсутствует.\n")
		return b.String()
	}

	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- Agitators: **%d**\n", len(result.Agitators)))
	b.WriteString(fmt.Sprintf("- Hot topics: **%d**\n\n", len(result.HotTopics)))

	b.WriteString("## Agitators\n\n")
	if len(result.Agitators) == 0 {
		b.WriteString("_Нет данных._\n\n")
	} else {
		for _, a := range result.Agitators {
			b.WriteString(fmt.Sprintf("### %s\n\n", formatUserHeader(a.User, users)))
			b.WriteString(fmt.Sprintf("- Messages: %d\n", a.MessageCount))
			b.WriteString(fmt.Sprintf("- Что цепляет: %s\n", sanitizeLLMLine(a.WhatTriggers)))
			b.WriteString(fmt.Sprintf("- Типичная реакция: %s\n", sanitizeLLMLine(a.Style)))
			b.WriteString(fmt.Sprintf("- Мимо не пройдёт: %s\n\n", sanitizeLLMLine(a.WontPassBy)))
		}
	}

	b.WriteString("## Hot topics\n\n")
	if len(result.HotTopics) == 0 {
		b.WriteString("_Нет данных._\n")
	} else {
		for _, t := range result.HotTopics {
			b.WriteString(fmt.Sprintf("- %s\n", sanitizeLLMLine(t)))
		}
		b.WriteString("\n")
	}

	if result.Notes != nil && strings.TrimSpace(*result.Notes) != "" {
		b.WriteString("## Notes\n\n")
		b.WriteString(sanitizeLLMLine(*result.Notes))
		b.WriteString("\n")
	}

	return b.String()
}

func formatUserHeader(user string, users map[int64]domain.UserRef) string {
	user = strings.TrimSpace(user)
	if user == "" {
		return "User"
	}
	id, err := strconv.ParseInt(user, 10, 64)
	if err != nil || id == 0 || users == nil {
		return "User " + user
	}
	u, ok := users[id]
	if !ok {
		return "User " + user
	}
	var parts []string
	parts = append(parts, "User "+user)
	if strings.TrimSpace(u.Username) != "" {
		parts = append(parts, "@"+strings.TrimSpace(u.Username))
	}
	fn := strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
	fn = strings.TrimSpace(fn)
	if fn != "" {
		parts = append(parts, fn)
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[0] + " (" + strings.Join(parts[1:], ", ") + ")"
}

func sanitizeLLMLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Trim common trailing junk: quotes and commas.
	for {
		trimmed := strings.TrimSpace(s)
		next := strings.TrimSuffix(trimmed, "”")
		next = strings.TrimSuffix(next, "\",")
		next = strings.TrimSuffix(next, "\" ,")
		next = strings.TrimSuffix(next, "\", ")
		next = strings.TrimSuffix(next, "\"")
		next = strings.TrimSpace(next)
		if next == s {
			break
		}
		s = next
	}
	// If the whole line is quoted, strip outer quotes.
	if len(s) >= 2 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		s = strings.TrimPrefix(strings.TrimSuffix(s, "\""), "\"")
		s = strings.TrimSpace(s)
	}
	return s
}

