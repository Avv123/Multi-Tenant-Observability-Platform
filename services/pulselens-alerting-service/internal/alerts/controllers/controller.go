package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	alertrequests "github.com/Avv123/pulselens-alerting-service/internal/alerts/requests"
	alertservices "github.com/Avv123/pulselens-alerting-service/internal/alerts/services"
	commonauth "github.com/Avv123/pulselens-common/auth"
	platformresponse "github.com/Avv123/pulselens-platform/response"
)

type Controller struct {
	service *alertservices.Service
}

func NewController(_ context.Context) (*Controller, error) {
	return &Controller{service: alertservices.New()}, nil
}

func (c *Controller) CreateRule(ctx *gin.Context) {
	var request alertrequests.CreateRuleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, wrapError("BAD_REQUEST", err.Error()))
		return
	}
	row, customError := c.service.CreateRule(ctx, claimsFromContext(ctx), &request)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

func (c *Controller) UpdateRule(ctx *gin.Context) {
	var request alertrequests.UpdateRuleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, wrapError("BAD_REQUEST", err.Error()))
		return
	}
	row, customError := c.service.UpdateRule(ctx, claimsFromContext(ctx), ctx.Param("rule_id"), &request)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) ListRules(ctx *gin.Context) {
	rows, customError := c.service.ListRules(ctx, claimsFromContext(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) CreatePolicy(ctx *gin.Context) {
	var request alertrequests.CreateAlertPolicyRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, wrapError("BAD_REQUEST", err.Error()))
		return
	}
	row, customError := c.service.CreatePolicy(ctx, claimsFromContext(ctx), &request)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

func (c *Controller) UpdatePolicy(ctx *gin.Context) {
	var request alertrequests.UpdateAlertPolicyRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, wrapError("BAD_REQUEST", err.Error()))
		return
	}
	row, customError := c.service.UpdatePolicy(ctx, claimsFromContext(ctx), ctx.Param("policy_id"), &request)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) ListPolicies(ctx *gin.Context) {
	rows, customError := c.service.ListPolicies(ctx, claimsFromContext(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) ListIncidents(ctx *gin.Context) {
	rows, customError := c.service.ListIncidents(ctx, claimsFromContext(ctx), alertservices.IncidentFiltersFromQuery(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) GetIncident(ctx *gin.Context) {
	row, customError := c.service.GetIncident(ctx, claimsFromContext(ctx), ctx.Param("incident_id"))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) ListIncidentTimeline(ctx *gin.Context) {
	rows, customError := c.service.ListIncidentTimeline(ctx, claimsFromContext(ctx), ctx.Param("incident_id"))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) ListIncidentDeliveries(ctx *gin.Context) {
	rows, customError := c.service.ListIncidentDeliveries(ctx, claimsFromContext(ctx), ctx.Param("incident_id"))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) AcknowledgeIncident(ctx *gin.Context) {
	row, customError := c.service.AcknowledgeIncident(ctx, claimsFromContext(ctx), ctx.Param("incident_id"))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) ResolveIncident(ctx *gin.Context) {
	row, customError := c.service.ResolveIncident(ctx, claimsFromContext(ctx), ctx.Param("incident_id"))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) AssignIncident(ctx *gin.Context) {
	var request alertrequests.AssignIncidentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, wrapError("BAD_REQUEST", err.Error()))
		return
	}
	row, customError := c.service.AssignIncident(ctx, claimsFromContext(ctx), ctx.Param("incident_id"), &request)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) CreateNotificationChannel(ctx *gin.Context) {
	var request alertrequests.CreateNotificationChannelRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, wrapError("BAD_REQUEST", err.Error()))
		return
	}
	row, customError := c.service.CreateNotificationChannel(ctx, claimsFromContext(ctx), &request)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

func (c *Controller) ListNotificationChannels(ctx *gin.Context) {
	rows, customError := c.service.ListNotificationChannels(ctx, claimsFromContext(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) ListNotificationDeliveries(ctx *gin.Context) {
	rows, customError := c.service.ListNotificationDeliveries(ctx, claimsFromContext(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) AddIncidentComment(ctx *gin.Context) {
	var request alertrequests.CreateIncidentCommentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, wrapError("BAD_REQUEST", err.Error()))
		return
	}
	row, customError := c.service.AddIncidentComment(ctx, claimsFromContext(ctx), ctx.Param("incident_id"), &request)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

func (c *Controller) ListIncidentComments(ctx *gin.Context) {
	rows, customError := c.service.ListIncidentComments(ctx, claimsFromContext(ctx), ctx.Param("incident_id"))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func claimsFromContext(ctx *gin.Context) *commonauth.Claims {
	value, _ := ctx.Get("claims")
	if claims, ok := value.(*commonauth.Claims); ok {
		return claims
	}
	return &commonauth.Claims{}
}
