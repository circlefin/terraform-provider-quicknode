default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: generate
generate:
	go generate ./... && \
	go run github.com/google/addlicense -c "Circle Internet Financial, LTD.  All rights reserved." -l "apache" -v -s `find . -name "*.go" -type f -print0 | xargs -0`

