package main

type UserConfigError struct {
	Err error
}

func (e *UserConfigError) Error() string {
	return "UserConfigError: " + e.Err.Error()
}
