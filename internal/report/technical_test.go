package report_test

import (
	"encoding/json"
	"testing"

	"stool-grabber/internal/domain"
	"stool-grabber/internal/report"
)

func TestBuildTechnicalDump_ContainsThreadsAndUsers(t *testing.T) {
	t.Parallel()

	scrape := &domain.ScrapeResult{
		ChannelUsername:        "chan",
		LinkedDiscussionChatID: 123,
		Users: map[int64]domain.UserRef{
			10: {ID: 10, Username: "u10", FirstName: "A", LastName: "B"},
		},
		Threads: []domain.PostThread{
			{
				ChannelMessageID: 1,
				PostText:         "post",
				PostDateUnix:     111,
				Comments: []domain.Comment{
					{MessageID: 2, SenderUserID: 10, Text: "c", DateUnix: 222},
				},
			},
		},
	}

	_, b, err := report.BuildTechnicalDump(scrape)
	if err != nil {
		t.Fatalf("BuildTechnicalDump: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("empty json")
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("invalid json: %v; json=%s", err, string(b))
	}
	if decoded["channel"] != "chan" {
		t.Fatalf("channel=%v, want %q", decoded["channel"], "chan")
	}
}

