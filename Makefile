# Developer helpers

.PHONY: test
test:
	go test

.PHONY: format
format:
	go mod tidy
	gofumports -l -w ./


# Run static code analysis
.PHONY: lint
lint:
	golangci-lint run
