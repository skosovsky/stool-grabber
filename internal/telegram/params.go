package telegram

// ScrapeParams is a narrow DTO for Telegram scraping use-case.
type ScrapeParams struct {
	ChannelUsername string
	ParseDepth      int
	DelayMS         int
	ExcludeAdmins   bool
}

// ScrapeOptions is the preferred name for ScrapeParams.
type ScrapeOptions = ScrapeParams

