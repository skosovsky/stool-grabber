package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"stool-grabber/internal/domain"
)

type TechnicalDump struct {
	Channel          string                 `json:"channel"`
	GeneratedAtUnix  int64                  `json:"generated_at_unix"`
	Threads          []TechnicalThread      `json:"threads"`
	Users            map[string]UserSummary `json:"users"`
	DiscussionChatID int64                  `json:"discussion_chat_id,omitempty"`
}

type TechnicalThread struct {
	ChannelMessageID int               `json:"channel_message_id"`
	PostText         string            `json:"post_text"`
	PostDateUnix     int               `json:"post_date_unix"`
	Comments         []domain.Comment  `json:"comments"`
}

type UserSummary struct {
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

func BuildTechnicalDump(scrape *domain.ScrapeResult) (*TechnicalDump, []byte, error) {
	if scrape == nil {
		return nil, nil, fmt.Errorf("scrape is nil")
	}
	threads := make([]TechnicalThread, 0, len(scrape.Threads))
	for _, t := range scrape.Threads {
		threads = append(threads, TechnicalThread{
			ChannelMessageID: t.ChannelMessageID,
			PostText:         t.PostText,
			PostDateUnix:     t.PostDateUnix,
			Comments:         t.Comments,
		})
	}

	users := make(map[string]UserSummary)
	for id, u := range scrape.Users {
		if id == 0 {
			continue
		}
		users[fmt.Sprintf("%d", id)] = UserSummary{
			Username:  u.Username,
			FirstName: u.FirstName,
			LastName:  u.LastName,
		}
	}

	dump := &TechnicalDump{
		Channel:          scrape.ChannelUsername,
		GeneratedAtUnix:  time.Now().Unix(),
		Threads:          threads,
		Users:            users,
		DiscussionChatID: scrape.LinkedDiscussionChatID,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(dump); err != nil {
		return nil, nil, fmt.Errorf("encode technical dump: %w", err)
	}

	return dump, bytes.TrimSpace(buf.Bytes()), nil
}

