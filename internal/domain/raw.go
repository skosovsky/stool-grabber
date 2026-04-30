// Package domain содержит нейтральные к транспорту типы между этапами пайплайна.
package domain

type UserRef struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
}

// ScrapeResult — сырой результат MTProto-скрапинга для Task 05 (агрегация).
type ScrapeResult struct {
	ChannelUsername        string
	LinkedDiscussionChatID int64
	ChannelAdminUserIDs    []int64
	Users                 map[int64]UserRef
	Threads                []PostThread
}

// PostThread — пост канала и собранные ответы из группы обсуждений.
type PostThread struct {
	ChannelMessageID int
	PostText         string
	PostDateUnix     int
	Comments         []Comment
}

// Comment — одно сообщение из треда getReplies (типично супергруппа обсуждений).
type Comment struct {
	MessageID    int
	SenderUserID int64
	Text         string
	DateUnix     int
}
