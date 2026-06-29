package pulselens_error

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Avv123/pulselens-platform/errs"
	platformresponse "github.com/Avv123/pulselens-platform/response"
)

const (
	BadRequest     errs.Code = "BAD_REQUEST"
	Unauthorized   errs.Code = "UNAUTHORIZED"
	Forbidden      errs.Code = "FORBIDDEN"
	InternalServer errs.Code = "INTERNAL_SERVER_ERROR"
	TooMany        errs.Code = "TOO_MANY_REQUESTS"
)

var codeToStatus = map[errs.Code]int{
	BadRequest:     http.StatusBadRequest,
	Unauthorized:   http.StatusUnauthorized,
	Forbidden:      http.StatusForbidden,
	InternalServer: http.StatusInternalServerError,
	TooMany:        http.StatusTooManyRequests,
}

func NewErrorResponse(ctx *gin.Context, customError errs.CustomError) {
	statusCode, exists := codeToStatus[customError.Code()]
	if !exists {
		statusCode = http.StatusInternalServerError
	}
	platformresponse.Error(ctx, statusCode, customError)
}
