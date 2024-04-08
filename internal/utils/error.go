package utils

import (
	"encoding/json"
	"fmt"
)

const (
	RequestErrorSummary  = "Request Error"
	ClientErrorSummary   = "Client Error"
	InternalErrorSummary = "Internal Error"
)

type ErrorResponse struct {
	Error *string `json:"error"`
}

func BuildRequestErrorMessage(status string, body []byte) (string, error) {
	m := fmt.Sprintf("Did not get expected status code, got status code `%s`", status)

	if len(body) != 0 {
		var e ErrorResponse
		err := json.Unmarshal(body, &e)
		if err != nil {
			return m, err
		}

		if e.Error != nil {
			m += fmt.Sprintf("\nerror `%s`", *e.Error)
		}
	}

	return m, nil
}

func BuildClientErrorMessage(err error) string {
	m := fmt.Sprintf("Unable to make request, got error: %s", err)

	return m
}

func BuildInternalErrorMessage(err error) string {
	m := fmt.Sprintf("An internal error occurred, %s", err)
	return m
}
