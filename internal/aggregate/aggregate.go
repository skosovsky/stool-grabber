package aggregate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"stool-grabber/internal/domain"
)

// UserMessages — компактная структура, оптимизированная под отправку в LLM слой.
type UserMessages struct {
	User     string   `json:"user"`
	Messages []string `json:"messages"`
}

type Result struct {
	Users              []UserMessages
	UsersBeforeTopN    int
	UsersAfterTopN     int
	UsersJSON         []byte
	JSON              []byte // deprecated: use UsersJSON
	FilteredByMinCount int
}

func Aggregate(params Params, raw *domain.ScrapeResult) (*Result, error) {
	if raw == nil {
		return nil, fmt.Errorf("raw is nil")
	}

	admins := make(map[int64]struct{}, len(raw.ChannelAdminUserIDs))
	for _, id := range raw.ChannelAdminUserIDs {
		if id == 0 {
			continue
		}
		admins[id] = struct{}{}
	}

	byUser := make(map[int64][]string)
	for _, thread := range raw.Threads {
		_ = thread
		for _, c := range thread.Comments {
			uid := c.SenderUserID
			if uid == 0 {
				continue
			}
			if params.ExcludeAdmins {
				if _, isAdmin := admins[uid]; isAdmin {
					continue
				}
			}
			if c.Text == "" {
				continue
			}
			byUser[uid] = append(byUser[uid], c.Text)
		}
	}

	type kv struct {
		id       int64
		messages []string
	}
	users := make([]kv, 0, len(byUser))
	filtered := 0
	for uid, msgs := range byUser {
		if len(msgs) < params.MinCommentsToAnalyze {
			filtered++
			continue
		}
		users = append(users, kv{id: uid, messages: msgs})
	}

	sort.Slice(users, func(i, j int) bool {
		li, lj := len(users[i].messages), len(users[j].messages)
		if li != lj {
			return li > lj
		}
		return users[i].id < users[j].id
	})

	beforeTopN := len(users)
	maxN := params.MaxUsersToAnalyze
	if maxN < 0 {
		maxN = 0
	}
	if maxN > 0 && len(users) > maxN {
		users = users[:maxN]
	}

	outUsers := make([]UserMessages, 0, len(users))
	for _, u := range users {
		outUsers = append(outUsers, UserMessages{
			User:     strconv.FormatInt(u.id, 10),
			Messages: u.messages,
		})
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(outUsers); err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	return &Result{
		Users:              outUsers,
		UsersBeforeTopN:    beforeTopN,
		UsersAfterTopN:     len(outUsers),
		UsersJSON:         bytes.TrimSpace(buf.Bytes()),
		JSON:              bytes.TrimSpace(buf.Bytes()),
		FilteredByMinCount: filtered,
	}, nil
}

