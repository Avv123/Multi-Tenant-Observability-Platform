package services

import (
	"strings"

	"github.com/gin-gonic/gin"
	alertrepositories "github.com/Avv123/pulselens-alerting-service/internal/alerts/repositories"
)

func IncidentFiltersFromQuery(ctx *gin.Context) alertrepositories.IncidentFilters {
	return alertrepositories.IncidentFilters{
		Status:     strings.TrimSpace(ctx.Query("status")),
		AssignedTo: strings.TrimSpace(ctx.Query("assigned_to")),
		ServiceID:  strings.TrimSpace(ctx.Query("service_id")),
		Severity:   strings.TrimSpace(ctx.Query("severity")),
	}
}
