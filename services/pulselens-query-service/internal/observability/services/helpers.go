package services

import (
	"encoding/json"
	"strings"

	"github.com/omniful/pulselens-platform/errs"
	"github.com/omniful/pulselens-platform/idgen"
	observabilitymodels "github.com/omniful/pulselens-query-service/internal/observability/models"
	observabilityrequests "github.com/omniful/pulselens-query-service/internal/observability/requests"
	observabilityresponses "github.com/omniful/pulselens-query-service/internal/observability/responses"
)

func normalizeDashboardRequest(request observabilityrequests.UpdateDashboardRequest) observabilityrequests.UpdateDashboardRequest {
	if strings.TrimSpace(request.DefaultTimeRange) == "" {
		request.DefaultTimeRange = "120m"
	}
	for index := range request.Widgets {
		if strings.TrimSpace(request.Widgets[index].ID) == "" {
			request.Widgets[index].ID = newWidgetID()
		}
		if request.Widgets[index].Filters == nil {
			request.Widgets[index].Filters = map[string]any{}
		}
		if request.Widgets[index].Layout == nil {
			request.Widgets[index].Layout = map[string]any{}
		}
	}
	if request.Layout == nil {
		request.Layout = map[string]any{"columns": 2}
	}
	return request
}

func createToUpdateDashboardRequest(request observabilityrequests.CreateDashboardRequest) observabilityrequests.UpdateDashboardRequest {
	return normalizeDashboardRequest(observabilityrequests.UpdateDashboardRequest{
		Name:             request.Name,
		Description:      request.Description,
		DefaultTimeRange: request.DefaultTimeRange,
		Layout:           request.Layout,
		Widgets:          request.Widgets,
	})
}

func buildDashboardResponse(row observabilitymodels.Dashboard) observabilityresponses.Dashboard {
	response := observabilityresponses.Dashboard{
		ID:               row.ID,
		Name:             row.Name,
		Description:      row.Description,
		DefaultTimeRange: row.DefaultTimeRange,
		Layout:           map[string]any{},
		Widgets:          []observabilityresponses.DashboardWidget{},
		CreatedBy:        row.CreatedBy,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
	_ = json.Unmarshal([]byte(row.Layout), &response.Layout)
	parsedWidgets := make([]observabilitymodels.DashboardWidget, 0)
	_ = json.Unmarshal([]byte(row.Widgets), &parsedWidgets)
	for _, widget := range parsedWidgets {
		response.Widgets = append(response.Widgets, observabilityresponses.DashboardWidget{
			ID:        widget.ID,
			Title:     widget.Title,
			Type:      widget.Type,
			Dataset:   widget.Dataset,
			ChartType: widget.ChartType,
			Metric:    widget.Metric,
			Filters:   widget.Filters,
			Layout:    widget.Layout,
			ValueKey:  widget.ValueKey,
			LabelKey:  widget.LabelKey,
		})
	}
	return response
}

func telemetryUnavailableError(err error) errs.CustomError {
	if err == nil {
		return errs.CustomError{}
	}
	return errs.New("DEPENDENCY_UNAVAILABLE", err.Error())
}

func newWidgetID() string {
	return idgen.New("widget")
}
