default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: generate
generate:
	go generate ./... && \
	go run github.com/google/addlicense -c "Circle Internet Financial, LTD.  All rights reserved." -l "apache" -v -s `find . -name "*.go" -type f -print0 | xargs -0`

.PHONY: vendor
vendor:
	curl --fail --retry 5 --retry-max-time 120 --retry-connrefused -u ${QUICKNODE_SWAGGER_USER}:${QUICKNODE_SWAGGER_PASSWORD} -s https://www.quicknode.com/api-docs/v0/swagger.json -o api/quicknode/openapi.json

.PHONY: validate
validate:
	goreleaser check
