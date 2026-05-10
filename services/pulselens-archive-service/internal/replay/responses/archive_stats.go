package responses

type ArchiveStatsResponse struct {
	ReplayJobCount     int64 `json:"replay_job_count"`
	ArchivedEvents     int64 `json:"archived_event_count"`
	ArchiveObjectCount int64 `json:"archive_object_count"`
}
