package main

type GuildCountResponse struct {
	Guilds    int64 `json:"guild_count"`
	Shards    int64 `json:"shard_count"`
	Timestamp int64 `json:"timestamp"`
	DryRun    bool  `json:"dry_run"`
}

type GuildCountRequestBody struct {
	Guilds int64 `json:"guild_count" validate:"required,number"`
	Shards int64 `json:"shard_count" validate:"required,number"`
	DryRun bool  `json:"dry_run" validate:"boolean"`
}

type BotListServiceResponse struct {
	ShortName  string `json:"short_name"`
	Url        string `json:"url"`
	GuildCount int64  `json:"guild_count"`
	Error      bool   `json:"error" validate:"omitempty"`
}

type BotListServicesResponse struct {
	Services    []BotListServiceResponse `json:"services"`
	LastUpdated int64                    `json:"last_updated"`
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
