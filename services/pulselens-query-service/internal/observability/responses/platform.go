package responses

import (
	platformruntime "github.com/Avv123/pulselens-platform/runtime"
	observabilitymodels "github.com/Avv123/pulselens-query-service/internal/observability/models"
)

type PlatformOverview struct {
	Runtime      []platformruntime.Heartbeat      `json:"runtime"`
	CleanupRuns  []observabilitymodels.CleanupRun `json:"cleanup_runs"`
	Dependencies []DependencyHealthRow            `json:"dependencies,omitempty"`
	KafkaLag     []KafkaLagRow                    `json:"kafka_lag,omitempty"`
}

type DependencyHealthRow struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CheckedAt string `json:"checked_at"`
}

type KafkaLagRow struct {
	GroupID        string `json:"group_id"`
	Topic          string `json:"topic"`
	Partition      int32  `json:"partition"`
	CurrentOffset  int64  `json:"current_offset"`
	LatestOffset   int64  `json:"latest_offset"`
	Lag            int64  `json:"lag"`
	MemberAssigned bool   `json:"member_assigned"`
}
