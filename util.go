package main

import (
	"fmt"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func ErrServiceNotFound(service string) error {
	return fmt.Errorf("Service %s not found", service)
}
