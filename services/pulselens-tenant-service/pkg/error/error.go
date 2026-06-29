package pulselens_error

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Avv123/pulselens-platform/errs"
	platformresponse "github.com/Avv123/pulselens-platform/response"
)

const (
	BadRequest     errs.Code = "BAD_REQUEST"
	Unauthorized   errs.Code = "UNAUTHORIZED"
	Forbidden      errs.Code = "FORBIDDEN"
	NotFound       errs.Code = "NOT_FOUND"
	Conflict       errs.Code = "CONFLICT"
	InternalServer errs.Code = "INTERNAL_SERVER_ERROR"
	RequestInvalid errs.Code = "REQUEST_INVALID"
	ParseError     errs.Code = "PARSE_ERROR"
)

var codeToStatus = map[errs.Code]int{
	BadRequest:     http.StatusBadRequest,
	Unauthorized:   http.StatusUnauthorized,
	Forbidden:      http.StatusForbidden,
	NotFound:       http.StatusNotFound,
	Conflict:       http.StatusConflict,
	InternalServer: http.StatusInternalServerError,
	RequestInvalid: http.StatusBadRequest,
	ParseError:     http.StatusBadRequest,
}

func NewErrorResponse(ctx *gin.Context, customError errs.CustomError) {
	statusCode, exists := codeToStatus[customError.Code()]
	if !exists {
		statusCode = http.StatusInternalServerError
	}
	platformresponse.Error(ctx, statusCode, customError)
}

func InvalidRequest(_ context.Context, message string) errs.CustomError {
	return errs.New(RequestInvalid, message)
}
