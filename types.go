package main

type GuildCountResponse struct {
	Count     int64 `json:"guild_count"`
	Timestamp int64 `json:"timestamp"`
}

type GuildCountRequestBody struct {
	Count     int64 `json:"guild_count" validate:"required,number"`
	Timestamp int64 `json:"timestamp" validate:"required,number"`
}

type BotListServiceResponse struct {
	ShortName  string `json:"short_name"`
	Url        string `json:"url"`
	GuildCount int64  `json:"guild_count"`
}

type BotListServicesResponse struct {
	Services    []BotListServiceResponse `json:"services"`
	LastUpdated int64                    `json:"last_upated"`
}

type BotListServiceConfig struct {
	ShortName    string
	LongName     string
	Url          string
	GetStatsUrl  string
	PostStatsUrl string
	Accessor     string
	Key          string
	Enabled      bool
}

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}
