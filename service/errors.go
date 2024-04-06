package service

import (
	"errors"
	"fmt"
	"runtime"
)

var (
	ErrContainerNotFound = errors.New("container not found")
	ErrApiError          = errors.New("api error")
)

func getCallerLocation(skip int) string {
	location := "unknown"
	_, file, no, ok := runtime.Caller(skip + 1)
	if ok {
		location = fmt.Sprintf("%s:%d", file, no)
	}

	return location
}

func FnErrContainerNotFound(ctrName string) error {
	location := getCallerLocation(1)
	return fmt.Errorf("[%s] %w:%s", location, ErrContainerNotFound, ctrName)
}

func FnErrApiError(err error) error {
	location := getCallerLocation(1)
	return fmt.Errorf("[%s] %w:%w", location, ErrContainerNotFound, err)
}
