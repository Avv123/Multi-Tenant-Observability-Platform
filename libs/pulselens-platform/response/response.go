package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Avv123/pulselens-platform/errs"
)

type SuccessPayload struct {
	IsSuccess  bool        `json:"is_success"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
	Meta       interface{} `json:"meta,omitempty"`
}

type ErrorPayload struct {
	IsSuccess  bool   `json:"is_success"`
	StatusCode int    `json:"status_code"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
}

func Success(ctx *gin.Context, data interface{}) {
	WithStatus(ctx, http.StatusOK, data, nil)
}

func WithStatus(ctx *gin.Context, statusCode int, data interface{}, meta interface{}) {
	ctx.AbortWithStatusJSON(statusCode, SuccessPayload{
		IsSuccess:  true,
		StatusCode: statusCode,
		Data:       data,
		Meta:       meta,
	})
}

func Error(ctx *gin.Context, statusCode int, customError errs.CustomError) {
	ctx.AbortWithStatusJSON(statusCode, ErrorPayload{
		IsSuccess:  false,
		StatusCode: statusCode,
		ErrorCode:  string(customError.Code()),
		Message:    customError.Message(),
	})
}
