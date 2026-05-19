package npmctl

type NPMError struct {
	Message string
}

func (e NPMError) Error() string {
	return e.Message
}

func npmError(message string) error {
	return NPMError{Message: message}
}
