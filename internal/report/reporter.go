package report

import (
	"fmt"
	"strings"

	"stool-grabber/internal/ai/contractgen"
)

const skipText = "Целевое ядро не выявлено по текущим порогам фильтрации."

func RenderSkipReport() string {
	return "# Отчёт\n\n" + skipText + "\n"
}

func RenderLLMReport(params Params, result *contractgen.AnalyzeCoreOutput) string {
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
			b.WriteString(fmt.Sprintf("### User %s\n\n", a.User))
			b.WriteString(fmt.Sprintf("- Messages: %d\n", a.MessageCount))
			b.WriteString(fmt.Sprintf("- Что цепляет: %s\n", a.WhatTriggers))
			b.WriteString(fmt.Sprintf("- Типичная реакция: %s\n", a.Style))
			b.WriteString(fmt.Sprintf("- Мимо не пройдёт: %s\n\n", a.WontPassBy))
		}
	}

	b.WriteString("## Hot topics\n\n")
	if len(result.HotTopics) == 0 {
		b.WriteString("_Нет данных._\n")
	} else {
		for _, t := range result.HotTopics {
			b.WriteString(fmt.Sprintf("- %s\n", t))
		}
		b.WriteString("\n")
	}

	if result.Notes != nil && strings.TrimSpace(*result.Notes) != "" {
		b.WriteString("## Notes\n\n")
		b.WriteString(strings.TrimSpace(*result.Notes))
		b.WriteString("\n")
	}

	return b.String()
}

