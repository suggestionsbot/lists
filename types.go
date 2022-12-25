package main

type GuildCountResponse struct {
	Guilds    int64 `json:"guild_count" example:"50000"`
	Shards    int64 `json:"shard_count" example:"50"`
	Timestamp int64 `json:"timestamp" example:"1671940391185"`
	DryRun    bool  `json:"dry_run" example:"false"`
}

type GuildCountRequestBody struct {
	Guilds int64 `json:"guild_count" validate:"required,number" example:"50000"`
	Shards int64 `json:"shard_count" validate:"required,number" example:"50"`
	DryRun bool  `json:"dry_run" validate:"boolean" example:"true"`
}

type BotListServiceResponse struct {
	ShortName  string `json:"short_name" example:"topgg"`
	Url        string `json:"url" example:"https://top.gg"`
	GuildCount int64  `json:"guild_count" example:"50000"`
	Error      bool   `json:"error" validate:"omitempty" example:"false"`
}

type BotListServicesResponse struct {
	Services    []BotListServiceResponse `json:"services"`
	LastUpdated int64                    `json:"last_updated" example:"1671940391185"`
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

type ResponseHTTP struct {
	Data    interface{} `json:"data"`
	Success bool        `json:"success" example:"true"`
	Nonce   int64       `json:"nonce" example:"1671940391185"`
}

type DefaultFiberError struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"missing or malformed API Key"`
}

type InvalidServiceError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"The service 'memelist' is not a valid service."`
}

type ResponseHTTPError struct {
	Data    interface{} `json:"data"`
	Success bool        `json:"success" example:"false"`
	Nonce   int64       `json:"nonce" example:"1671940391185"`
}
