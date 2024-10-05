package web

type ErrorResponse struct {
	Err error
}

func (e ErrorResponse) Data() (data []byte, contentType string, err error) {
	return []byte(e.Err.Error()), "text/plain", nil
}

func (e ErrorResponse) Error() string {
	return e.Err.Error()
}

func (e ErrorResponse) Unwrap() error {
	return e.Err
}
