package aggregate

import (
	"encoding/json"
	"testing"

	"stool-grabber/internal/domain"
)

func TestAggregate_MinCommentsAndTopNAndSorting(t *testing.T) {
	t.Parallel()

	params := Params{MinCommentsToAnalyze: 2, MaxUsersToAnalyze: 2, ExcludeAdmins: false}

	raw := &domain.ScrapeResult{
		Threads: []domain.PostThread{
			{
				ChannelMessageID: 1,
				PostText:         "Почему микросервисы умирают",
				Comments: []domain.Comment{
					{SenderUserID: 10, Text: "a"},
					{SenderUserID: 10, Text: "b"},
					{SenderUserID: 20, Text: "c"},
					{SenderUserID: 20, Text: "d"},
					{SenderUserID: 20, Text: "e"},
					{SenderUserID: 30, Text: "x"}, // будет отфильтрован по min_comments
				},
			},
		},
	}

	res, err := Aggregate(params, raw)
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	if res.UsersBeforeTopN != 2 {
		t.Fatalf("UsersBeforeTopN=%d, want 2", res.UsersBeforeTopN)
	}
	if res.UsersAfterTopN != 2 {
		t.Fatalf("UsersAfterTopN=%d, want 2", res.UsersAfterTopN)
	}
	if len(res.Users) != 2 {
		t.Fatalf("len(Users)=%d, want 2", len(res.Users))
	}
	// user 20 (3 msg) должен быть первым.
	if res.Users[0].UserID != "20" {
		t.Fatalf("first user_id=%q, want %q", res.Users[0].UserID, "20")
	}
	if res.Users[1].UserID != "10" {
		t.Fatalf("second user_id=%q, want %q", res.Users[1].UserID, "10")
	}
	if len(res.UsersJSON) == 0 {
		t.Fatalf("UsersJSON empty")
	}

	var decoded []UserMessages
	if err := json.Unmarshal(res.UsersJSON, &decoded); err != nil {
		t.Fatalf("unmarshal JSON: %v; json=%s", err, string(res.UsersJSON))
	}
	if len(decoded) != 2 {
		t.Fatalf("decoded len=%d, want 2", len(decoded))
	}
	if got := decoded[0].Messages[0]; got == "c" {
		t.Fatalf("expected post context prefix, got %q", got)
	}
}

func TestAggregate_ExcludeAdmins(t *testing.T) {
	t.Parallel()

	params := Params{MinCommentsToAnalyze: 1, MaxUsersToAnalyze: 10, ExcludeAdmins: true}

	raw := &domain.ScrapeResult{
		ChannelAdminUserIDs: []int64{42},
		Threads: []domain.PostThread{
			{
				ChannelMessageID: 1,
				Comments: []domain.Comment{
					{SenderUserID: 42, Text: "admin msg"},
					{SenderUserID: 7, Text: "user msg"},
				},
			},
		},
	}

	res, err := Aggregate(params, raw)
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}
	if len(res.Users) != 1 {
		t.Fatalf("len(Users)=%d, want 1", len(res.Users))
	}
	if res.Users[0].UserID != "7" {
		t.Fatalf("user_id=%q, want %q", res.Users[0].UserID, "7")
	}
}

func TestAggregate_IgnoresZeroUserIDAndEmptyText(t *testing.T) {
	t.Parallel()

	params := Params{MinCommentsToAnalyze: 1, MaxUsersToAnalyze: 10, ExcludeAdmins: false}

	raw := &domain.ScrapeResult{
		Threads: []domain.PostThread{
			{
				ChannelMessageID: 1,
				Comments: []domain.Comment{
					{SenderUserID: 0, Text: "ignored"},
					{SenderUserID: 5, Text: ""},
					{SenderUserID: 5, Text: "ok"},
				},
			},
		},
	}

	res, err := Aggregate(params, raw)
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}
	if len(res.Users) != 1 {
		t.Fatalf("len(Users)=%d, want 1", len(res.Users))
	}
	if got := len(res.Users[0].Messages); got != 1 {
		t.Fatalf("messages=%d, want 1", got)
	}
}

func TestAggregate_PopulatesUsernameFromDirectory(t *testing.T) {
	t.Parallel()

	params := Params{MinCommentsToAnalyze: 1, MaxUsersToAnalyze: 10, ExcludeAdmins: false}

	raw := &domain.ScrapeResult{
		Users: map[int64]domain.UserRef{
			7:  {ID: 7, Username: "alice"},
			10: {ID: 10, FirstName: "Bob", LastName: "Builder"},
		},
		Threads: []domain.PostThread{
			{
				ChannelMessageID: 1,
				PostText:         "Тестовый пост",
				Comments: []domain.Comment{
					{SenderUserID: 7, Text: "x"},
					{SenderUserID: 10, Text: "y"},
				},
			},
		},
	}

	res, err := Aggregate(params, raw)
	if err != nil {
		t.Fatalf("Aggregate() error: %v", err)
	}

	byID := make(map[string]UserMessages)
	for _, u := range res.Users {
		byID[u.UserID] = u
	}
	if got := byID["7"].Username; got != "@alice" {
		t.Fatalf("username=%q, want %q", got, "@alice")
	}
	if got := byID["10"].Username; got != "Bob Builder" {
		t.Fatalf("username=%q, want %q", got, "Bob Builder")
	}
}

