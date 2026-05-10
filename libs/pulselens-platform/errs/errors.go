package errs

type Code string

type CustomError struct {
	code    Code
	message string
}

func New(code Code, message string) CustomError {
	return CustomError{
		code:    code,
		message: message,
	}
}

func (e CustomError) Error() string {
	return e.message
}

func (e CustomError) Message() string {
	return e.message
}

func (e CustomError) Code() Code {
	return e.code
}

func (e CustomError) Exists() bool {
	return e.code != ""
}
