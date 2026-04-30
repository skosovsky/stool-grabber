package telegram

type ProgressEvent struct {
	Stage           string
	PostIndex       int
	PostTotal       int
	PostID          int
	CommentsFetched int
	Message         string
}

// ScrapeParams is a narrow DTO for Telegram scraping use-case.
type ScrapeParams struct {
	ChannelUsername string
	ParseDepth      int
	DelayMS         int
	ExcludeAdmins   bool
	Progress        func(event ProgressEvent)
}

// ScrapeOptions is the preferred name for ScrapeParams.
type ScrapeOptions = ScrapeParams

