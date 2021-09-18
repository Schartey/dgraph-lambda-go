package api

type HttpResponseStatus int

type LambdaError struct {
	Underlying error
	Status     HttpResponseStatus
}

func (l *LambdaError) Error() string {
	return l.Underlying.Error()
}
