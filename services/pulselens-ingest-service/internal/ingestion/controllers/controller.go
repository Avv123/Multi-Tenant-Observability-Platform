package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	pulsetelemetry "github.com/Avv123/pulselens-common/telemetry"
	"github.com/Avv123/pulselens-ingest-service/internal/ingestion/services"
	pulselens_error "github.com/Avv123/pulselens-ingest-service/pkg/error"
	platformerrs "github.com/Avv123/pulselens-platform/errs"
	platformresponse "github.com/Avv123/pulselens-platform/response"
)

type Controller struct {
	service *services.Service
}

func NewController(_ context.Context) (*Controller, error) {
	return &Controller{service: services.New()}, nil
}

func (c *Controller) Ingest(ctx *gin.Context) {
	var request pulsetelemetry.BatchIngestRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		pulselens_error.NewErrorResponse(ctx, platformerrs.New(pulselens_error.BadRequest, err.Error()))
		return
	}

	response, customError := c.service.Ingest(ctx, ctx.GetHeader("X-API-Key"), &request)
	if customError.Exists() {
		pulselens_error.NewErrorResponse(ctx, customError)
		return
	}

	platformresponse.WithStatus(ctx, http.StatusAccepted, response, nil)
}
