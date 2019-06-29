package main

import (
	"fmt"
	"net/http"
)

type HttpResponseStatusCodeNotOKError struct {
	HttpResponse *http.Response
}

func (*HttpResponseStatusCodeNotOKError) Error() string {
	return "response status code not 200"
}

type KahlaResponseCodeNotZeroError struct {
	Tag     string
	Code    int32
	Message string
}

func (r *KahlaResponseCodeNotZeroError) Error() string {
	return fmt.Sprintf("kahla response code not 0. %s. %s (%d)", r.Tag, r.Message, r.Code)
}
