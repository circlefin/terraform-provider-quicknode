default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: generate
generate:
	go generate ./... && \
	go run github.com/google/addlicense -c "Circle Internet Group, Inc.  All rights reserved." -l "apache" -v -s `find . -name "*.go" -type f -print0 | xargs -0`

CURL_FLAGS := --fail --retry 5 --retry-max-time 120 --retry-connrefused -s

.PHONY: vendor
vendor:
	curl $(CURL_FLAGS) https://www.quicknode.com/api-docs/v0/swagger.json  | jq . > api/quicknode/openapi.json
	curl $(CURL_FLAGS) https://api.quicknode.com/streams/rest/openapi.json | jq . > api/streams/streams-openapi.json
	go generate ./api/...

.PHONY: validate
validate:
	goreleaser check
