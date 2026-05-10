package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	replayrequests "github.com/omniful/pulselens-archive-service/internal/replay/requests"
	replayservices "github.com/omniful/pulselens-archive-service/internal/replay/services"
	commonauth "github.com/omniful/pulselens-common/auth"
	"github.com/omniful/pulselens-platform/errs"
	platformresponse "github.com/omniful/pulselens-platform/response"
)

type Controller struct {
	service *replayservices.Service
}

func NewController(_ context.Context) (*Controller, error) {
	return &Controller{service: replayservices.New()}, nil
}

func (c *Controller) CreateReplayJob(ctx *gin.Context) {
	var request replayrequests.CreateReplayJobRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, errs.New("BAD_REQUEST", err.Error()))
		return
	}
	row, customError := c.service.CreateReplayJob(ctx, claimsFromContext(ctx), &request)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

func (c *Controller) ListReplayJobs(ctx *gin.Context) {
	rows, customError := c.service.ListReplayJobs(ctx, claimsFromContext(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) Stats(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.Stats(ctx, claimsFromContext(ctx)))
}

func claimsFromContext(ctx *gin.Context) *commonauth.Claims {
	value, _ := ctx.Get("claims")
	if claims, ok := value.(*commonauth.Claims); ok {
		return claims
	}
	return &commonauth.Claims{}
}

func statusCode(code errs.Code) int {
	switch code {
	case "BAD_REQUEST":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
