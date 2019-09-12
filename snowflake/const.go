package snowflake

type Code string

const (
	TimeFormat      = "2006-01-02 15:04:05"
	SUCCESS    Code = "SUCCESS"
	ERR500     Code = "ERR500"
	ERR400     Code = "ERR400"
)

type BaseApiResult struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

type SnowflakeApiResult struct {
	BaseApiResult
	Data []int64 `json:"data"`
}

func (r *BaseApiResult) IsSuccess() bool {
	return r.Code == SUCCESS
}
