package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"stool-grabber/internal/aggregate"
	"stool-grabber/internal/ai/contractgen"
)

func normalizeUsername(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "@")
	return strings.TrimSpace(s)
}

// BuildAnalyzeCoreInput собирает типобезопасный вход для Analyze (Task 06).
// Реальный вызов LLM выполняется в Task 07.
func BuildAnalyzeCoreInput(params AnalyzeParams, agg *aggregate.Result) (contractgen.PromptID, contractgen.AnalyzeCoreInput, error) {
	var zero contractgen.AnalyzeCoreInput
	if agg == nil {
		return "", zero, fmt.Errorf("aggregate result is nil")
	}

	items := make([]contractgen.AnalyzeCoreInputUsersItem, 0, len(agg.Users))
	for _, u := range agg.Users {
		items = append(items, contractgen.AnalyzeCoreInputUsersItem{
			UserId:   u.UserID,
			Username: u.Username,
			Messages: u.Messages,
		})
	}

	usersJSON, err := json.Marshal(items)
	if err != nil {
		return "", zero, fmt.Errorf("marshal users json: %w", err)
	}

	in := contractgen.AnalyzeCoreInput{
		ChannelUsername: normalizeUsername(params.ChannelUsername),
		PromptTemplate:  params.PromptTemplate,
		Users:           items,
		UsersJson:       string(usersJSON),
	}

	return contractgen.AnalyzeCore, in, nil
}

