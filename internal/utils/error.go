// Copyright 2024 Circle Internet Financial, LTD.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
