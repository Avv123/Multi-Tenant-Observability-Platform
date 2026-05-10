package responses

type IngestResponse struct {
	Requested           int   `json:"requested"`
	Accepted            int   `json:"accepted"`
	RemainingRate       int64 `json:"remaining_rate"`
	RemainingDailyQuota int64 `json:"remaining_daily_quota"`
}
