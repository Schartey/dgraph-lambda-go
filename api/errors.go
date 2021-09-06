package api

type LambdaResponseStatus int

const (
	OK                  LambdaResponseStatus = 200
	CREATED                                  = 201
	ACCEPTED                                 = 202
	NO_CONTENT                               = 204
	BAD_REQUEST                              = 400
	UNAUTHORIZED                             = 401
	FORBIDDEN                                = 403
	NOT_FOUND                                = 404
	NOT_ALLOWED                              = 405
	PRECONDITION_FAILED                      = 412
	INTERNAL_ERROR                           = 500
	NOT_IMPLEMENTED                          = 501
)

type LambdaError struct {
	Underlying error
	Status     LambdaResponseStatus
}

func (l *LambdaError) Error() string {
	return l.Underlying.Error()
}
