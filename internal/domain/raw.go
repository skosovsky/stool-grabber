// Package domain содержит нейтральные к транспорту типы между этапами пайплайна.
package domain

// ScrapeResult — сырой результат MTProto-скрапинга для Task 05 (агрегация).
type ScrapeResult struct {
	ChannelUsername        string
	LinkedDiscussionChatID int64
	ChannelAdminUserIDs    []int64
	Threads                []PostThread
}

// PostThread — пост канала и собранные ответы из группы обсуждений.
type PostThread struct {
	ChannelMessageID int
	Comments         []Comment
}

// Comment — одно сообщение из треда getReplies (типично супергруппа обсуждений).
type Comment struct {
	MessageID    int
	SenderUserID int64
	Text         string
	DateUnix     int
}
