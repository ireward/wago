### Variables ###
COLOR := "\e[1;36m%s\e[0m\n"

### Arguments ###
GOPATH ?= $(shell go env GOPATH)

### Test ###
.PHONY: test
test:
	@printf $(COLOR) "Running go test in shuffle and short mode ..."
	@go test \
		-shuffle=on \
		-count=1 \
		-short \
		-timeout=5m \
		./...
		-coverprofile=coverage.out


.PHONY: test-coverage
test-coverage:
	@go tool cover -func=./coverage.out
