package main

type GuildCountResponse struct {
	Count     int64 `json:"guild_count"`
	Timestamp int64 `json:"timestamp"`
}

type BotListServiceResponse struct {
	Id         int64  `json:"id"`
	ShortName  string `json:"short_name"`
	Url        string `json:"url"`
	GuildCount int64  `json:"guild_count"`
}

type BotListServicesResponse struct {
	Services    []*BotListServiceResponse `json:"services"`
	LastUpdated int64                     `json:"last_upated"`
}
