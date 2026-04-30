package ai

import (
	"encoding/json"
	"strings"
	"testing"

	"stool-grabber/internal/ai/contractgen"
)

func TestApplyBudget_TrimsMessagesPerUserAndChars(t *testing.T) {
	t.Parallel()

	in := contractgen.AnalyzeCoreInput{
		Users: []contractgen.AnalyzeCoreInputUsersItem{
			{UserId: "1", Username: "@u1", Messages: []string{"aaaaa", "bbbbb", "ccccc"}},
		},
	}

	got := ApplyBudget(in, BudgetOptions{
		MaxMessagesPerUser: 2,
		MaxCharsPerMessage: 3,
		MaxUsersJSONChars:  10_000,
	})

	if len(got.Users) != 1 {
		t.Fatalf("users=%d, want 1", len(got.Users))
	}
	if len(got.Users[0].Messages) != 2 {
		t.Fatalf("messages=%d, want 2", len(got.Users[0].Messages))
	}
	if got.Users[0].Messages[0] != "aaa" {
		t.Fatalf("msg0=%q, want %q", got.Users[0].Messages[0], "aaa")
	}
	if got.Users[0].Messages[1] != "bbb" {
		t.Fatalf("msg1=%q, want %q", got.Users[0].Messages[1], "bbb")
	}
}

func TestApplyBudget_RebuildsUsersJsonAndCapsSize(t *testing.T) {
	t.Parallel()

	in := contractgen.AnalyzeCoreInput{
		Users: []contractgen.AnalyzeCoreInputUsersItem{
			{UserId: "1", Username: "@u1", Messages: []string{strings.Repeat("x", 100)}},
			{UserId: "2", Username: "@u2", Messages: []string{strings.Repeat("y", 100)}},
		},
	}

	got := ApplyBudget(in, BudgetOptions{
		MaxMessagesPerUser: 1,
		MaxCharsPerMessage: 100,
		MaxUsersJSONChars:  50,
	})

	if got.UsersJson == "" {
		t.Fatalf("UsersJson is empty")
	}
	if len(got.UsersJson) > 50 {
		t.Fatalf("UsersJson len=%d, want <= 50", len(got.UsersJson))
	}
}

func TestApplyBudget_DeterministicOutput(t *testing.T) {
	t.Parallel()

	in := contractgen.AnalyzeCoreInput{
		Users: []contractgen.AnalyzeCoreInputUsersItem{
			{UserId: "1", Username: "@u1", Messages: []string{"hello", "world"}},
		},
	}
	opt := BudgetOptions{MaxMessagesPerUser: 2, MaxCharsPerMessage: 10, MaxUsersJSONChars: 1000}

	a := ApplyBudget(in, opt)
	b := ApplyBudget(in, opt)

	if a.UsersJson != b.UsersJson {
		t.Fatalf("UsersJson differs:\nA=%q\nB=%q", a.UsersJson, b.UsersJson)
	}
	// Ensure JSON is valid when not truncated by MaxUsersJSONChars.
	var decoded []map[string]any
	if err := json.Unmarshal([]byte(a.UsersJson), &decoded); err != nil {
		t.Fatalf("UsersJson invalid JSON: %v", err)
	}
}

