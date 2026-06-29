package controllers

import "github.com/Avv123/pulselens-platform/errs"

func statusCode(code errs.Code) int {
	switch code {
	case "BAD_REQUEST":
		return 400
	case "UNAUTHORIZED":
		return 401
	case "FORBIDDEN":
		return 403
	case "NOT_FOUND":
		return 404
	default:
		return 500
	}
}

func wrapError(code errs.Code, message string) errs.CustomError {
	return errs.New(code, message)
}
