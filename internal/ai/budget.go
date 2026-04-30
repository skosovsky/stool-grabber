package ai

import (
	"encoding/json"
	"unicode/utf8"

	"stool-grabber/internal/ai/contractgen"
)

type BudgetOptions struct {
	MaxMessagesPerUser int
	MaxCharsPerMessage int
	MaxUsersJSONChars  int
}

func DefaultBudgetOptions() BudgetOptions {
	return BudgetOptions{
		MaxMessagesPerUser: 30,
		MaxCharsPerMessage: 900,
		MaxUsersJSONChars:  200_000,
	}
}

func trimRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	// O(n) but deterministic; fine for our sizes.
	out := make([]rune, 0, max)
	for _, r := range s {
		out = append(out, r)
		if len(out) >= max {
			break
		}
	}
	return string(out)
}

// ApplyBudget trims Analyze input deterministically BEFORE prompt rendering.
// It keeps top-N users (already capped by aggregator), then caps messages per user and chars per message.
// Finally it rebuilds UsersJson and truncates it if it still exceeds MaxUsersJSONChars.
func ApplyBudget(in contractgen.AnalyzeCoreInput, opt BudgetOptions) contractgen.AnalyzeCoreInput {
	if opt.MaxMessagesPerUser <= 0 {
		opt.MaxMessagesPerUser = DefaultBudgetOptions().MaxMessagesPerUser
	}
	if opt.MaxCharsPerMessage <= 0 {
		opt.MaxCharsPerMessage = DefaultBudgetOptions().MaxCharsPerMessage
	}
	if opt.MaxUsersJSONChars <= 0 {
		opt.MaxUsersJSONChars = DefaultBudgetOptions().MaxUsersJSONChars
	}

	users := make([]contractgen.AnalyzeCoreInputUsersItem, 0, len(in.Users))
	for _, u := range in.Users {
		msgs := u.Messages
		if len(msgs) > opt.MaxMessagesPerUser {
			msgs = msgs[:opt.MaxMessagesPerUser]
		}
		outMsgs := make([]string, 0, len(msgs))
		for _, m := range msgs {
			outMsgs = append(outMsgs, trimRunes(m, opt.MaxCharsPerMessage))
		}
		users = append(users, contractgen.AnalyzeCoreInputUsersItem{
			UserId:   u.UserId,
			Username: u.Username,
			Messages: outMsgs,
		})
	}
	in.Users = users

	b, err := json.Marshal(users)
	if err == nil {
		in.UsersJson = string(b)
	}
	if len(in.UsersJson) > opt.MaxUsersJSONChars {
		in.UsersJson = in.UsersJson[:opt.MaxUsersJSONChars]
	}
	return in
}

