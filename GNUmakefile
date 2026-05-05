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
	# The upstream QuickNode spec reuses the `deleteTag` operationId for two
	# distinct paths (`/v0/endpoints/{id}/tags/{tag_id}` and
	# `/v0/endpoints/tags/{id}`), which trips up oapi-codegen. Rename the
	# account-level op locally to `deleteAccountTag` until QuickNode fixes
	# this upstream; remove the jq step once they do.
	curl $(CURL_FLAGS) https://www.quicknode.com/api-docs/v0/swagger.json \
		| jq '.paths."/v0/endpoints/tags/{id}".delete.operationId = "deleteAccountTag"' \
		> api/quicknode/openapi.json
	curl $(CURL_FLAGS) https://api.quicknode.com/streams/rest/openapi.json | jq . > api/streams/streams-openapi.json
	go generate ./api/...

.PHONY: validate
validate:
	goreleaser check
